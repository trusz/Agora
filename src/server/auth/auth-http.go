package auth

import (
	"agora/src/log"
	"agora/src/user"
	"context"
	"net/http"
)

func (ah *AuthHandler) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie("token")
		cookieFound := err == nil && cookie.Value != ""

		tryingToLogin := r.URL.Path == "/login" ||
			r.URL.Path == "/callback" ||
			(r.URL.Path == "/" && r.URL.Query().Get("code") != "")

		if !cookieFound {
			log.Debug.Println("No valid token found in cookie, redirecting to login")
			if tryingToLogin {
				next.ServeHTTP(w, r)
				return
			}

			http.Redirect(w, r, "/login", http.StatusFound)
			return
		}

		token, err := ah.VerifyToken(cookie.Value)
		if err != nil {
			if tryingToLogin {
				next.ServeHTTP(w, r)
				return
			}
			log.Error.Printf("Invalid token: %v, redirecting to login\n", err)
			http.Redirect(w, r, "/login", http.StatusFound)
			return
		}

		claims, ok := token.Claims.(*CustomClaims)
		if !ok {
			log.Error.Println("Token claims are not of type CustomClaims, redirecting to login")
			http.Redirect(w, r, "/login", http.StatusFound)
			return
		}

		loggedInUser := user.User{
			ID:    claims.UserID,
			Name:  claims.Name,
			Email: claims.Email,
		}

		ctx := context.WithValue(r.Context(), "user", loggedInUser)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func ExtractUserFromContext(ctx context.Context) (user.User, bool) {
	loggedInUser, ok := ctx.Value("user").(user.User)
	if !ok {
		log.Error.Println("Could not extract user from context")
		return user.User{}, false
	}
	return loggedInUser, true
}
