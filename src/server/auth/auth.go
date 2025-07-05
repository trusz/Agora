package auth

import (
	"agora/src/user"
	"time"

	"golang.org/x/oauth2"
)

type AuthHandler struct {
	jwtSecret   string
	issuer      string
	oauthConfig *oauth2.Config
	userHandler *user.UserHandler
}

func NewAuthHandler(
	jwtSecret,
	issuer,
	azureTenantID,
	azureClientID,
	azureClientSecret string,
	redirectURL string,
	userHandler *user.UserHandler,
) *AuthHandler {
	return &AuthHandler{
		// I add a timestampt to make sure users are re-logged in every time
		// I restart the server for example in case of a new version.
		// A reson for this is that I check at login if the users exists in the database
		// and if I clear the database for some reason between restarts the user
		// is still logged in through the token but not in the database anymore
		// this makes sure that does not happen
		jwtSecret: jwtSecret + time.Now().Format("20060102150405"),
		// jwtSecret:   jwtSecret,
		issuer:      issuer,
		oauthConfig: generateOAuth2Config(azureClientID, azureClientSecret, azureTenantID, redirectURL),
		userHandler: userHandler,
	}
}

func generateOAuth2Config(clientID, clientSecret, tenantID string, redirectURL string) *oauth2.Config {
	return &oauth2.Config{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		RedirectURL:  redirectURL, //  in my case RedirectURL:  "http://localhost:8080/callback"
		Scopes:       []string{"User.Read"},
		Endpoint: oauth2.Endpoint{
			AuthURL:  "https://login.microsoftonline.com/" + tenantID + "/oauth2/v2.0/authorize",
			TokenURL: "https://login.microsoftonline.com/" + tenantID + "/oauth2/v2.0/token",
		},
	}
}
