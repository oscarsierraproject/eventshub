package v1rest

// Author: Sebastian Oleksiak (oscarsierraproject@protonmail.com)
// License: The Unlicense
// Created: August 18, 2024

import (
	"errors"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

func createJWT(username string) (string, error) {
	// Create a JSON Web Token (JWT) based on an open standard (RFC 7519) based on the provided username.
	//
	// The username parameter is the user's identifier.
	// Returns a string representing the JWT token and an error if the token creation process fails.
	token := jwt.New(jwt.SigningMethodHS512)
	claims := token.Claims.(jwt.MapClaims)

	claims["exp"] = time.Now().Add(1 * time.Minute).Unix()
	claims["authorized"] = true
	claims["user"] = username

	secret := os.Getenv("GOCALENDAR_TOKEN_SECRET")
	if secret == "" {
		err := errors.New("failed to obtain token secret")
		panic(err)
	}

	tokenStr, err := token.SignedString([]byte(secret))
	if err != nil {
		fmt.Errorf("something went wrong during token creation process: %s", err.Error())
		return "", err
	}

	return tokenStr, nil
}

func validateJWT(w http.ResponseWriter, r *http.Request) (err error) {

	if r.Header["Token"] == nil {
		return errors.New("token is not present in HEADER")
	}

	keyFunc := func(token *jwt.Token) (interface{}, error) {
		/* Receive the parsed token.
		 * Return the cryptographic key for verifying the signature.
		 */
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("there was an error in parsing")
		}

		secret := os.Getenv("GOCALENDAR_TOKEN_SECRET")
		if secret == "" {
			err := errors.New("failed to obtain token secret")
			panic(err)
		}

		return []byte(secret), nil
	}

	token, err := jwt.Parse(r.Header["Token"][0], keyFunc)
	if token == nil || err != nil {
		return errors.New("There was an error during token parsing.")
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return errors.New("There was an error during claims parsing.")
	}

	exp := claims["exp"].(float64)
	if int64(exp) < time.Now().Local().Unix() {
		return errors.New("Token has expired.")
	}

	return nil
}
