package auth

import (
	"agora/src/user"

	"golang.org/x/oauth2"
)

type AuthHandler struct {
	jwtSecret   string
	issuer      string
	oauthConfig *oauth2.Config
	userHandler *user.UserHandler
}

func NewAuthHandler(jwtSecret, issuer, azureTenantID, azureClientID, azureClientSecret string, userHandler *user.UserHandler) *AuthHandler {
	return &AuthHandler{
		jwtSecret:   jwtSecret,
		issuer:      issuer,
		oauthConfig: generateOAuth2Config(azureClientID, azureClientSecret, azureTenantID),
		userHandler: userHandler,
	}
}

func generateOAuth2Config(clientID, clientSecret, tenantID string) *oauth2.Config {
	return &oauth2.Config{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		RedirectURL:  "http://localhost:54324/login/callback", //  in my case RedirectURL:  "http://localhost:8080/callback"
		Scopes:       []string{"User.Read"},
		Endpoint: oauth2.Endpoint{
			AuthURL:  "https://login.microsoftonline.com/" + tenantID + "/oauth2/v2.0/authorize",
			TokenURL: "https://login.microsoftonline.com/" + tenantID + "/oauth2/v2.0/token",
		},
	}
}
