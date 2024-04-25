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
		rid := getRequestID(w)

		expirationSeconds := 60 * 60 * 24

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
		log.Println(rid, "handleLogin email", params.Email)
		if err != nil {
			log.Println(rid, "Error decoding user parameters", err, params)
			respondWithError(w, 500, "Failed to decode request", err)
			return
		}

		user, err := db.ValidateLogin(params.Email, params.Password)
		if err != nil {
			log.Println(rid, "Error while validating login", err)
			respondWithError(w, 401, "Error handling request", err)
			return
		}

		if params.Expiration > 0 && params.Expiration < expirationSeconds {
			expirationSeconds = params.Expiration
		}

		jwt, err := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.RegisteredClaims{
			Issuer:    "chirpy",
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Duration(expirationSeconds) * time.Second)),
			Subject:   fmt.Sprint(user.Id),
		}).SignedString(jwtSecret)
		if err != nil {
			log.Printf(rid, "Error creating jwt", err, params)
			respondWithError(w, 500, "Error handling request", err)
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
		rid := getRequestID(w)
		type parameters struct {
			Email    string `json:"email"`
			Password string `json:"password"`
		}
		decoder := json.NewDecoder(r.Body)
		params := parameters{}
		err := decoder.Decode(&params)
		log.Println(rid, "handlePutUsers", params.Email)
		if err != nil {
			respondWithError(w, 500, "Failed to decode request body", err)
			return
		}

		tokenString := strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer ")
		log.Println(rid, "Received token", tokenString)

		// The keyFunc should take the parsed but unverified token, do any checks to make sure the token is of a valid format, and then return the signing key to verify the authenticity of the token against.
		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			return jwtSecret, nil
		}, jwt.WithValidMethods([]string{"HS256"}))
		if err != nil {
			respondWithError(w, 401, "Failed to parse token", err)
			return
		}

		var idStr string
		idStr, err = token.Claims.GetSubject()
		if err != nil {
			respondWithError(w, 401, "No user ID given", err)
			return
		}

		var id int
		id, err = strconv.Atoi(idStr)
		if err != nil {
			respondWithError(w, 401, "Given user ID is not a number", err)
			return
		}

		var user database.SafeUser
		user, err = db.UpdateUser(id, params.Email, params.Password)
		if err != nil {
			respondWithError(w, 500, "Failed to update user (ID might be incorrect)", err)
			return
		}
		respondWithJSON(w, 200, user)
	})
}
