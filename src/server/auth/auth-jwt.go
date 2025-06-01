package auth

import (
	"agora/src/log"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// CustomClaims defines the structure of our JWT claims
type CustomClaims struct {
	UserID string `json:"userid"`
	Name   string `json:"name"`
	Email  string `json:"email"`
	jwt.RegisteredClaims
}

func (ah *AuthHandler) createJWT(userID string, name string, email string, issuer string) (jwt.Token, string) {
	// Create claims with RegisteredClaims for proper expiration handling
	now := time.Now()
	expirationTime := now.Add(time.Hour * 1)
	
	claims := CustomClaims{
		UserID: userID,
		Name:   name,
		Email:  email,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    issuer,
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(expirationTime),
		},
	}

	// Create token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// Sign token
	tokenString, err := token.SignedString([]byte(ah.jwtSecret))
	if err != nil {
		log.Error.Printf("Failed to sign token: %v\n", err)
		panic(err)
	}

	return *token, tokenString
}

func (ah *AuthHandler) VerifyToken(tokenString string) (jwt.Token, error) {
	token, err := jwt.ParseWithClaims(tokenString, &CustomClaims{}, func(token *jwt.Token) (interface{}, error) {
		// Check signing method
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(ah.jwtSecret), nil
	})

	if err != nil {
		log.Error.Printf("Invalid token: %v\n", err.Error())
		return jwt.Token{}, err
	}

	if claims, ok := token.Claims.(*CustomClaims); ok && token.Valid {
		log.Debug.Println("Token valid!")
		log.Debug.Printf("User ID: %v\n", claims.UserID)
		log.Debug.Printf("Name: %v\n", claims.Name)
		log.Debug.Printf("Email: %v\n", claims.Email)

		// We can safely access expiration time using RegisteredClaims methods
		expiry, err := claims.GetExpirationTime()
		if err == nil {
			log.Debug.Printf("Expires at: %v\n", expiry.Time)
		}
	} else {
		log.Debug.Println("Invalid token claims")
	}

	return *token, nil
}
