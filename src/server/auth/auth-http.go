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

		user := user.User{
			ID:    claims.UserID,
			Name:  claims.Name,
			Email: claims.Email,
		}

		ctx := context.WithValue(r.Context(), "user", user)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func (ah *AuthHandler) MockMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user := user.User{
			ID:    "999",
			Name:  "John Local",
			Email: "john@localhost.com",
		}
		if !ah.userHandler.UserExists(user.ID) {
			ah.userHandler.AddUser(user.ID, "John Local", user.Email)
		}
		ctx := context.WithValue(r.Context(), "user", user)
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
