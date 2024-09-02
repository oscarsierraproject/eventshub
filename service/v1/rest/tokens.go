package v1rest

// Author: Sebastian Oleksiak (oscarsierraproject@protonmail.com)
// License: The Unlicense
// Created: August 18, 2024

import (
	"errors"
	"net/http"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

const (
	tokenLifeTime time.Duration = 2 * time.Minute
)

// Create a JSON Web Token (JWT) based on an open standard (RFC 7519) based on the provided username.
// The username parameter is the user's identifier.
// Returns a string representing the JWT token and an error if the token creation process fails.
func createJWT(username string) (string, error) {
	token := jwt.New(jwt.SigningMethodHS512)

	if claims, ok := token.Claims.(jwt.MapClaims); ok {
		claims["exp"] = time.Now().Add(tokenLifeTime).Unix()
		claims["authorized"] = true
		claims["user"] = username
	} else {
		return "", errors.New("failed to obtain token claims")
	}

	secret := os.Getenv("GOCALENDAR_TOKEN_SECRET")
	if secret == "" {
		panic(errors.New("failed to obtain token secret"))
	}

	tokenStr, err := token.SignedString([]byte(secret))
	if err != nil {
		return "", err
	}

	return tokenStr, nil
}

func validateJWT(_ http.ResponseWriter, r *http.Request) (err error) {
	if r.Header["Token"] == nil {
		return errors.New("failed to obtain token from HEADER")
	}

	// Receive the parsed token.
	// Return the cryptographic key for verifying the signature.
	keyFunc := func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unsupported signing method")
		}

		secret := os.Getenv("GOCALENDAR_TOKEN_SECRET")
		if secret == "" {
			panic(errors.New("failed to obtain token secret"))
		}

		return []byte(secret), nil
	}

	token, err := jwt.Parse(r.Header["Token"][0], keyFunc)
	if token == nil || err != nil {
		return errors.New("there was an error during token parsing")
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return errors.New("there was an error during claims parsing")
	}

	exp, ok := claims["exp"].(float64)
	if !ok {
		return errors.New("failed to obtain token expiration time")
	}

	if int64(exp) < time.Now().Local().Unix() {
		return errors.New("token has expired")
	}

	return nil
}
