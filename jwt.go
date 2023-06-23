package main

import (
	"time"

	"github.com/dgrijalva/jwt-go"
)

// Define a secret key for signing the JWT tokens
var secretKey = []byte("Your_Secret_Key")

// User represents a user in the system
type User struct {
	ID       string `json:"id"`
	Username string `json:"username"`
	Password string `json:"password"`
}

// Claims represents the JWT claims
type Claims struct {
	Username string `json:"username"`
	jwt.StandardClaims
}

// Authenticate handles the authentication logic
func Authenticate(username, password string) bool {
	// Implement your authentication logic here
	// For simplicity, we'll use a hardcoded username and password
	return username == "admin" && password == "password"
}

// GenerateToken generates a JWT token for the given username
func GenerateToken(username string) (string, error) {
	expirationTime := time.Now().Add(24 * time.Hour)

	claims := &Claims{
		Username: username,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: expirationTime.Unix(),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(secretKey)
}
