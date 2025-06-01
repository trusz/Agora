package auth

import (
	"agora/src/log"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/oauth2"
)

func (ah *AuthHandler) HandleLogin(w http.ResponseWriter, r *http.Request) {
	url := ah.oauthConfig.AuthCodeURL("state-token", oauth2.AccessTypeOffline)
	http.Redirect(w, r, url, http.StatusFound)
}

func (ah *AuthHandler) MakeHandleCallback(fallbackHandler http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		code := r.URL.Query().Get("code")
		if code == "" {
			fallbackHandler(w, r)
			return
		}

		token, err := ah.oauthConfig.Exchange(context.Background(), code)
		if err != nil {
			http.Error(w, "Failed to exchange token: "+err.Error(), http.StatusInternalServerError)
			return
		}

		client := ah.oauthConfig.Client(context.Background(), token)
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
		if !ah.userHandler.UserExists(user.ID) {
			ah.userHandler.AddUser(user.ID, user.DisplayName, user.Mail)
		}

		newToken, jwtString := ah.createJWT(user.ID, user.DisplayName, user.Mail, ah.issuer)
		log.Debug.Printf("msg='JWT created' jwt='%#v'\n", newToken)
		expiry, err := newToken.Claims.GetExpirationTime()
		if err != nil {
			log.Error.Printf("Failed to get expiration time from JWT claims: %v\n", err)
			// Fallback to 1 hour from now if we can't extract expiration
			expiry = jwt.NewNumericDate(time.Now().Add(time.Hour * 1))
		}

		cookie := makeCookieOutOfOAuthToken(jwtString, expiry.Time)
		http.SetCookie(w, &cookie)
		http.Redirect(w, r, "/", http.StatusPermanentRedirect)

	}
}

func makeCookieOutOfOAuthToken(tokenString string, expiry time.Time) http.Cookie {
	cookie := http.Cookie{}
	cookie.Name = "token"
	cookie.Value = tokenString
	cookie.Expires = expiry
	cookie.Secure = false // TODO: Set to true if using HTTPS
	cookie.HttpOnly = true
	cookie.Path = "/"

	return cookie
}

// MSGraphUser represents a user object returned by Microsoft Graph API
type MSGraphUser struct {
	ODataContext      string   `json:"@odata.context"`
	BusinessPhones    []string `json:"businessPhones"`
	DisplayName       string   `json:"displayName"`
	GivenName         string   `json:"givenName"`
	JobTitle          *string  `json:"jobTitle"`
	Mail              string   `json:"mail"`
	MobilePhone       *string  `json:"mobilePhone"`
	OfficeLocation    *string  `json:"officeLocation"`
	PreferredLanguage string   `json:"preferredLanguage"`
	Surname           string   `json:"surname"`
	UserPrincipalName string   `json:"userPrincipalName"`
	ID                string   `json:"id"`
}
