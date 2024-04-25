package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/madsbv/boot-dev-go-server/internal/database"
)

func handleLogin(db *database.DB, jwtSecret []byte) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		type parameters struct {
			Email      string `json:"email"`
			Password   string `json:"password"`
			Expiration int    `json:"expires_in_seconds"`
		}

		type response struct {
			Email string `json:"email"`
			Id    int    `json:"id"`
			Token string `json:"token"`
		}

		decoder := json.NewDecoder(r.Body)
		params := parameters{}
		err := decoder.Decode(&params)
		log.Printf("Handling: %s", params.Email)
		if err != nil {
			log.Printf("Error decoding user parameters: %s\n%v", err, params)
			respondWithError(w, 500, "Failed to decode request")
			return
		}

		user, err := db.ValidateLogin(params.Email, params.Password)
		if err != nil {
			log.Printf("Error while validating login for user: %v", err)
			respondWithError(w, 401, "Error handling request")
			return
		}

		expiration := 60 * 60 * 24
		if params.Expiration > 0 && params.Expiration < expiration {
			expiration = params.Expiration
		}

		jwt, err := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.RegisteredClaims{
			Issuer:    "chirpy",
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Duration(expiration) * time.Second)),
			Subject:   string(user.Id),
		}).SignedString(jwtSecret)
		if err != nil {
			log.Printf("Error creating jwt for user:", err, params)
			respondWithError(w, 500, "Error handling request")
			return
		}

		respondWithJSON(w, 200, response{
			Email: user.Email,
			Id:    user.Id,
			Token: jwt,
		})

	})
}

func handlePutUsers(db *database.DB, jwtSecret []byte) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		type parameters struct {
			Email    string `json:"email"`
			Password string `json:"password"`
		}
		decoder := json.NewDecoder(r.Body)
		params := parameters{}
		err := decoder.Decode(&params)
		log.Printf("Handling request to update email to: %s", params.Email)
		if err != nil {
			log.Printf("Error decoding user parameters: %s\n%v", err, params)
			respondWithError(w, 500, "Failed to decode request")
			return
		}

		tokenString := strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer ")
		log.Printf("Received token: %v", tokenString)

		// The keyFunc should take the parsed but unverified token, do any checks to make sure the token is of a valid format, and then return the signing key to verify the authenticity of the token against.
		token, err := jwt.ParseWithClaims(tokenString, jwt.RegisteredClaims{Issuer: "chirpy"}, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("Unexpected signing method: %v", token.Header["alg"])
			}

			return jwtSecret, nil
		})
		if err != nil {
			respondWithError(w, 401, "Invalid token")
			return
		}

		var idStr string
		idStr, err = token.Claims.GetSubject()
		if err != nil {
			respondWithError(w, 401, "Invalid token")
			return
		}

		var id int
		id, err = strconv.Atoi(idStr)
		if err != nil {
			respondWithError(w, 401, "Invalid token")
			return
		}

		var user database.SafeUser
		user, err = db.UpdateUser(id, params.Email, params.Password)
		if err != nil {
			respondWithError(w, 500, "Failed to decode request")
			return
		}
		respondWithJSON(w, 200, user)
	})
}
