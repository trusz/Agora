package auth

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/AzureAD/microsoft-authentication-library-for-go/apps/public"
	"github.com/google/uuid"
)

// Config holds Azure AD authentication configuration
type Config struct {
	TenantID     string
	ClientID     string
	ClientSecret string
	RedirectURI  string
}

// Auth handles authentication with Azure AD
type Auth struct {
	db     *sql.DB
	config Config
}

// NewAuth creates a new Auth instance
func NewAuth(db *sql.DB, config Config) *Auth {
	return &Auth{
		db:     db,
		config: config,
	}
}

// Session stores user session information
type Session struct {
	ID        string
	UserID    string
	Token     string
	ExpiresAt time.Time
}

// LoginHandler handles the login process
func (a *Auth) LoginHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		// Show login page
		http.ServeFile(w, r, "src/templates/auth/login.html")
		return
	}

	// Handle login form submission - initiating Azure AD auth
	state := uuid.New().String()

	// Store state in session cookie for validation
	stateCookie := http.Cookie{
		Name:     "auth_state",
		Value:    state,
		HttpOnly: true,
		Secure:   r.TLS != nil,
		MaxAge:   int(time.Hour.Seconds()),
		Path:     "/",
	}
	http.SetCookie(w, &stateCookie)

	// Redirect to Azure login
	azureAuthority := fmt.Sprintf("https://login.microsoftonline.com/%s", a.config.TenantID)

	publicClientApp, err := public.New(a.config.ClientID, public.WithAuthority(azureAuthority))
	if err != nil {
		log.Printf("Error creating Azure AD client: %v", err)
		http.Error(w, "Authentication error", http.StatusInternalServerError)
		return
	}

	// We're not using this directly in a web app, but initiating the redirect ourselves
	authURL := fmt.Sprintf(
		"%s/oauth2/v2.0/authorize?client_id=%s&response_type=code&redirect_uri=%s&scope=openid%%20profile%%20email&state=%s",
		azureAuthority,
		a.config.ClientID,
		a.config.RedirectURI,
		state,
	)

	http.Redirect(w, r, authURL, http.StatusFound)
}

// LogoutHandler handles the logout process
func (a *Auth) LogoutHandler(w http.ResponseWriter, r *http.Request) {
	// Clear session cookie
	sessionCookie := http.Cookie{
		Name:     "session_token",
		Value:    "",
		HttpOnly: true,
		Secure:   r.TLS != nil,
		MaxAge:   -1,
		Path:     "/",
	}
	http.SetCookie(w, &sessionCookie)

	// Redirect to home page
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

// CallbackHandler processes the callback from Azure AD
func (a *Auth) CallbackHandler(w http.ResponseWriter, r *http.Request) {
	// Verify state to prevent CSRF
	stateCookie, err := r.Cookie("auth_state")
	if err != nil {
		http.Error(w, "Invalid authentication state", http.StatusBadRequest)
		return
	}

	stateParam := r.URL.Query().Get("state")
	if stateCookie.Value != stateParam {
		http.Error(w, "State mismatch", http.StatusBadRequest)
		return
	}

	// Clear state cookie
	http.SetCookie(w, &http.Cookie{
		Name:     "auth_state",
		Value:    "",
		HttpOnly: true,
		Secure:   r.TLS != nil,
		MaxAge:   -1,
		Path:     "/",
	})

	// Extract authorization code
	code := r.URL.Query().Get("code")
	if code == "" {
		http.Error(w, "No authorization code received", http.StatusBadRequest)
		return
	}

	// Exchange code for token
	token, err := a.exchangeCodeForToken(code)
	if err != nil {
		log.Printf("Error exchanging code for token: %v", err)
		http.Error(w, "Authentication error", http.StatusInternalServerError)
		return
	}

	// Create or update user
	user, err := a.processUserLogin(token)
	if err != nil {
		log.Printf("Error processing user login: %v", err)
		http.Error(w, "Authentication error", http.StatusInternalServerError)
		return
	}

	// Create session
	sessionToken := uuid.New().String()
	expiresAt := time.Now().Add(24 * time.Hour)

	// Store session in database
	_, err = a.db.Exec(
		"INSERT INTO sessions (id, user_id, token, expires_at) VALUES (?, ?, ?, ?)",
		uuid.New().String(),
		user.ID,
		sessionToken,
		expiresAt,
	)
	if err != nil {
		log.Printf("Error storing session: %v", err)
		http.Error(w, "Authentication error", http.StatusInternalServerError)
		return
	}

	// Set session cookie
	http.SetCookie(w, &http.Cookie{
		Name:     "session_token",
		Value:    sessionToken,
		HttpOnly: true,
		Secure:   r.TLS != nil,
		Expires:  expiresAt,
		Path:     "/",
	})

	// Redirect to home page
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

// TokenResponse represents the response from Azure AD token endpoint
type TokenResponse struct {
	AccessToken  string `json:"access_token"`
	IDToken      string `json:"id_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int    `json:"expires_in"`
}

// exchangeCodeForToken exchanges an authorization code for tokens
func (a *Auth) exchangeCodeForToken(code string) (string, error) {
	azureAuthority := fmt.Sprintf("https://login.microsoftonline.com/%s", a.config.TenantID)

	// We would typically use the MSAL library here, but for web apps,
	// we need to handle the token exchange manually
	tokenURL := fmt.Sprintf("%s/oauth2/v2.0/token", azureAuthority)

	// Create form data
	formData := fmt.Sprintf(
		"client_id=%s&client_secret=%s&code=%s&grant_type=authorization_code&redirect_uri=%s",
		a.config.ClientID,
		a.config.ClientSecret,
		code,
		a.config.RedirectURI,
	)

	// Make POST request
	req, err := http.NewRequest("POST", tokenURL, strings.NewReader(formData))
	if err != nil {
		return "", err
	}
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("error exchanging code for token: %s", resp.Status)
	}

	// Parse response
	var tokenResp TokenResponse
	err = json.NewDecoder(resp.Body).Decode(&tokenResp)
	if err != nil {
		return "", err
	}

	return tokenResp.IDToken, nil
}

// UserClaims represents the claims from the ID token
type UserClaims struct {
	Sub           string `json:"sub"`
	Name          string `json:"name"`
	Email         string `json:"email"`
	PreferredName string `json:"preferred_username"`
}

// processUserLogin processes a user login, creating or updating the user as needed
func (a *Auth) processUserLogin(idToken string) (*User, error) {
	// Parse ID token to get user info
	claims, err := parseIDToken(idToken)
	if err != nil {
		return nil, err
	}

	// Check if user exists
	var user User
	err = a.db.QueryRow(
		"SELECT id, name, email, username FROM users WHERE email = ?",
		claims.Email,
	).Scan(&user.ID, &user.Name, &user.Email, &user.Username)

	if err == sql.ErrNoRows {
		// Create new user
		user = User{
			ID:       uuid.New().String(),
			Name:     claims.Name,
			Email:    claims.Email,
			Username: claims.PreferredName,
			Token:    idToken,
		}

		_, err = a.db.Exec(
			"INSERT INTO users (id, name, email, username) VALUES (?, ?, ?, ?)",
			user.ID, user.Name, user.Email, user.Username,
		)
		if err != nil {
			return nil, err
		}
	} else if err != nil {
		return nil, err
	}

	// Update token
	user.Token = idToken

	return &user, nil
}

// parseIDToken parses the ID token to extract user claims
func parseIDToken(idToken string) (*UserClaims, error) {
	// In a real application, we would validate the token signature
	// and expiration, but for simplicity, we'll just parse it

	parts := strings.Split(idToken, ".")
	if len(parts) != 3 {
		return nil, fmt.Errorf("invalid token format")
	}

	// Decode payload
	payload, err := base64UrlDecode(parts[1])
	if err != nil {
		return nil, err
	}

	// Parse claims
	var claims UserClaims
	err = json.Unmarshal(payload, &claims)
	if err != nil {
		return nil, err
	}

	return &claims, nil
}

// base64UrlDecode decodes a base64url-encoded string
func base64UrlDecode(input string) ([]byte, error) {
	// Add padding if needed
	padding := 4 - (len(input) % 4)
	if padding < 4 {
		input += strings.Repeat("=", padding)
	}

	// Replace base64url encoding with standard base64 encoding
	input = strings.ReplaceAll(input, "-", "+")
	input = strings.ReplaceAll(input, "_", "/")

	return base64.StdEncoding.DecodeString(input)
}

// AuthMiddleware verifies a user is authenticated
func (a *Auth) AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Get session cookie
		cookie, err := r.Cookie("session_token")
		if err != nil {
			// Redirect to login page
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}

		// Get session from database
		var session Session
		err = a.db.QueryRow(
			"SELECT id, user_id, token, expires_at FROM sessions WHERE token = ?",
			cookie.Value,
		).Scan(&session.ID, &session.UserID, &session.Token, &session.ExpiresAt)

		if err != nil || time.Now().After(session.ExpiresAt) {
			// Invalid or expired session
			// Clear cookie
			http.SetCookie(w, &http.Cookie{
				Name:     "session_token",
				Value:    "",
				HttpOnly: true,
				Secure:   r.TLS != nil,
				MaxAge:   -1,
				Path:     "/",
			})

			// Redirect to login page
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}

		// Get user from database
		var user User
		err = a.db.QueryRow(
			"SELECT id, name, email, username FROM users WHERE id = ?",
			session.UserID,
		).Scan(&user.ID, &user.Name, &user.Email, &user.Username)

		if err != nil {
			// User not found
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}

		// Add user to request context
		ctx := context.WithValue(r.Context(), userContextKey, &user)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// User is a type key for use in context
type userContextKey int

const userContextKey userContextKey = iota

// GetUser gets the user from request context
func GetUser(r *http.Request) (*User, bool) {
	user, ok := r.Context().Value(userContextKey).(*User)
	return user, ok
}

// LoadConfig loads Azure AD configuration from environment variables
func LoadConfig() Config {
	return Config{
		TenantID:     os.Getenv("AZURE_TENANT_ID"),
		ClientID:     os.Getenv("AZURE_CLIENT_ID"),
		ClientSecret: os.Getenv("AZURE_CLIENT_SECRET"),
		RedirectURI:  os.Getenv("AZURE_REDIRECT_URI"),
	}
}
