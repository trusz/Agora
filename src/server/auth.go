package server

import (
	"agora/src/log"
	"agora/src/user"
	"context"
	"crypto/rsa"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math/big"
	"net/http"
	"strings"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/oauth2"
)

// AzureJWTClaims represents the claims in the JWT token from Azure AD
type AzureJWTClaims struct {
	jwt.RegisteredClaims
	Aio               string   `json:"aio,omitempty"`
	Name              string   `json:"name,omitempty"`
	Nonce             string   `json:"nonce,omitempty"`
	Oid               string   `json:"oid,omitempty"`
	PreferredUsername string   `json:"preferred_username,omitempty"`
	Rh                string   `json:"rh,omitempty"`
	Tid               string   `json:"tid,omitempty"`
	Uti               string   `json:"uti,omitempty"`
	Ver               string   `json:"ver,omitempty"`
	Roles             []string `json:"roles,omitempty"`
	Scope             string   `json:"scp,omitempty"`
	Groups            []string `json:"groups,omitempty"`
}

// ParseJWT parses a JWT token and returns the claims
// This function uses the jwt library but skips verification since we already
// trust the token from the OAuth2 exchange
func ParseJWT(tokenString string) (*AzureJWTClaims, error) {
	if tokenString == "" {
		return nil, errors.New("empty token")
	}

	// Parse token without verification - we already trust it from OAuth exchange
	token, err := jwt.ParseWithClaims(tokenString, &AzureJWTClaims{}, func(token *jwt.Token) (interface{}, error) {
		// Skip verification by returning a dummy key
		return []byte("skip-verification"), nil
	})

	if err != nil {
		// Check if it's a validation error that we want to ignore
		if strings.Contains(err.Error(), "signature is invalid") {
			// We're ignoring validation errors, so just get the claims
			if claims, ok := token.Claims.(*AzureJWTClaims); ok {
				return claims, nil
			}
		}
		return nil, fmt.Errorf("error parsing token: %w", err)
	}

	if claims, ok := token.Claims.(*AzureJWTClaims); ok {
		return claims, nil
	}

	return nil, errors.New("invalid token claims")
}

func startLogin(w http.ResponseWriter, r *http.Request) {
	oauth2Config := generateOAuth2Config()
	url := oauth2Config.AuthCodeURL("state-token", oauth2.AccessTypeOffline)
	http.Redirect(w, r, url, http.StatusFound)
}

func makeHandleCallback(fallbackCallback http.HandlerFunc, userHandler *user.UserHandler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		oauth2Config := generateOAuth2Config()
		code := r.URL.Query().Get("code")
		log.Debug.Printf("msg='received callback' code='%s'\n", code)
		if code == "" {
			fallbackCallback(w, r)
			return
		}

		token, err := oauth2Config.Exchange(context.Background(), code)
		if err != nil {
			http.Error(w, "Failed to exchange token: "+err.Error(), http.StatusInternalServerError)
			return
		}
		log.Debug.Printf("msg='token received' token='%s'\n", token.AccessToken)

		client := oauth2Config.Client(context.Background(), token)
		resp, err := client.Get("https://graph.microsoft.com/v1.0/me")
		if err != nil {
			http.Error(w, "Failed to get user info: "+err.Error(), http.StatusInternalServerError)
			return
		}
		defer resp.Body.Close()

		data, err := io.ReadAll(resp.Body)
		if err != nil {
			http.Error(w, "Failed to read user info: "+err.Error(), http.StatusInternalServerError)
			return
		}

		// Unmarshal the JSON data into the MSGraphUser struct
		var user MSGraphUser
		if err := json.Unmarshal(data, &user); err != nil {
			http.Error(w, "Failed to parse user info: "+err.Error(), http.StatusInternalServerError)
			return
		}

		log.Pretty(user)
		if !userHandler.UserExists(user.ID) {
			userHandler.AddUser(user.ID, user.DisplayName, user.Mail)
		}
		cookie := makeCookieOutOfOAuthToken(*token)
		http.SetCookie(w, &cookie)
		http.Redirect(w, r, "/", http.StatusPermanentRedirect)

	}
}

func makeCookieOutOfOAuthToken(token oauth2.Token) http.Cookie {
	cookie := http.Cookie{}
	cookie.Name = "token"
	cookie.Value = token.AccessToken
	cookie.Expires = token.Expiry
	cookie.Secure = false // TODO: Set to true if using HTTPS
	cookie.HttpOnly = true
	cookie.Path = "/"

	return cookie
}

func authMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie("token")
		if err != nil || cookie.Value == "" {
			log.Debug.Println("No valid token found in cookie, redirecting to login")
			tryingToLogin := r.URL.Path == "/login" ||
				r.URL.Path == "/callback" ||
				(r.URL.Path == "/" && r.URL.Query().Get("code") != "")
			if tryingToLogin {
				next.ServeHTTP(w, r)
				return
			}

			http.Redirect(w, r, "/login", http.StatusFound)
			return
		}
		token := oauth2.Token{AccessToken: cookie.Value}
		validate(token.AccessToken)
		// log.Debug.Printf("msg='token found in cookie' token='%s'\n", token.AccessToken)

		// Parse the JWT token
		// claims, err := ParseJWT(token.AccessToken)
		// if err == nil {
		// 	log.Debug.Printf("msg='parsed JWT token' user='%s' email='%s'\n",
		// 		claims.Name, claims.PreferredUsername)

		// 	// Add claims to request context
		// 	ctx := context.WithValue(r.Context(), "jwt_claims", claims)
		// 	r = r.WithContext(ctx)
		// } else {
		// 	log.Debug.Printf("msg='failed to parse JWT token' err='%s'\n", err.Error())
		// }

		next.ServeHTTP(w, r)
	})
}

func validate(tokenString string) error {
	config, _ := LoadAzureConfig()
	jwks := fetchJWKS(config.TenantID)

	kid, _ := getKID(tokenString)
	// log.Debug.Printf("msg='got kid from token' kid='%s'\n", kid)

	keyInfo, _ := getKeyForKID(jwks.Keys, kid)
	// log.Debug.Printf("msg='got key for kid' keyInfo='%#v' err='%s'\n", keyInfo, err)

	pubKey, _ := buildRSAPublicKey(keyInfo)
	log.Debug.Printf("msg='built RSA public key' key='%#v'\n", pubKey)

	// token, _ := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
	// 	// Check algorithm
	// 	if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
	// 		return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
	// 	}
	// 	return pubKey, nil
	// })
	// log.Debug.Printf("msg='parsed JWT token' token='%s' valid='%t'\n", tokenString, token.Valid)

	return nil
}

func fetchJWKS(tenantID string) jwks {

	jwksUrl := fmt.Sprintf("https://login.microsoftonline.com/%s/discovery/v2.0/keys", tenantID)
	resp, err := http.Get(jwksUrl)
	if err != nil {
		log.Error.Printf("msg='failed to fetch JWKS' err='%s'\n", err.Error())
	}

	defer resp.Body.Close()
	var j jwks
	if err := json.NewDecoder(resp.Body).Decode(&j); err != nil {
		log.Error.Printf("msg='failed to decode JWKS' err='%s'\n", err.Error())
	}

	return j
	// log.Debug.Printf("msg='fetched JWKS' keys=%#v\n", j.Keys)

}

type jwks struct {
	Keys []jwksKey `json:"keys"`
}

type jwksKey struct {
	Kty string `json:"kty"`
	Use string `json:"use"`
	Kid string `json:"kid"`
	N   string `json:"n"`
	E   string `json:"e"`
}

func getKID(tokenString string) (string, error) {
	parts := strings.Split(tokenString, ".")
	if len(parts) < 2 {
		return "", fmt.Errorf("malformed token")
	}
	headerPart, err := base64.RawURLEncoding.DecodeString(parts[0])
	if err != nil {
		return "", err
	}
	var header map[string]interface{}
	if err := json.Unmarshal(headerPart, &header); err != nil {
		return "", err
	}
	kid, ok := header["kid"].(string)
	if !ok {
		return "", fmt.Errorf("no kid in token header")
	}
	return kid, nil
}

func getKeyForKID(keys []jwksKey, kid string) (*jwksKey, error) {
	for _, k := range keys {
		if k.Kid == kid {
			return &k, nil
		}
	}
	return nil, fmt.Errorf("no key matching kid %s", kid)
}

func buildRSAPublicKey(jk *jwksKey) (*rsa.PublicKey, error) {
	nb, err := base64.RawURLEncoding.DecodeString(jk.N)
	if err != nil {
		return nil, fmt.Errorf("failed to decode n: %v", err)
	}
	eb, err := base64.RawURLEncoding.DecodeString(jk.E)
	if err != nil {
		return nil, fmt.Errorf("failed to decode e: %v", err)
	}
	e := 0
	for _, b := range eb {
		e = e<<8 + int(b)
	}
	pub := &rsa.PublicKey{
		N: new(big.Int).SetBytes(nb),
		E: e,
	}
	return pub, nil
}
