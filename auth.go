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

const expirationAccessSeconds = 60 * 60            // 1 hour
const expirationRefreshSeconds = 60 * 60 * 24 * 60 // 60 days
const accessIssuer = "chirpy-access"
const refreshIssuer = "chirpy-refresh"

func handlePostLogin(db *database.DB, jwtSecret []byte) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		rid := getRequestID(w)

		type parameters struct {
			Email      string `json:"email"`
			Password   string `json:"password"`
			Expiration int    `json:"expires_in_seconds"`
		}

		type response struct {
			Email        string `json:"email"`
			Id           int    `json:"id"`
			IsChirpyRed  bool   `json:"is_chirpy_red"`
			Token        string `json:"token"`
			RefreshToken string `json:"refresh_token"`
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

		expiration := expirationAccessSeconds
		if params.Expiration > 0 && params.Expiration < expiration {
			expiration = params.Expiration
		}

		jwt, err := newToken(fmt.Sprint(user.Id), accessIssuer, expiration, jwtSecret)
		if err != nil {
			log.Printf(rid, "Error creating jwt", err, params)
			respondWithError(w, 500, "Error handling request", err)
			return
		}

		jwtRefresh, err := newToken(fmt.Sprint(user.Id), refreshIssuer, expirationRefreshSeconds, jwtSecret)

		respondWithJSON(w, 200, response{
			Email:        user.Email,
			Id:           user.Id,
			IsChirpyRed:  user.IsChirpyRed,
			Token:        jwt,
			RefreshToken: jwtRefresh,
		})
	})
}

func newToken(id, issuer string, expirationSeconds int, key []byte) (string, error) {
	return jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.RegisteredClaims{
		Issuer:    issuer,
		IssuedAt:  jwt.NewNumericDate(time.Now()),
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Duration(expirationSeconds) * time.Second)),
		Subject:   fmt.Sprint(id),
	}).SignedString(key)
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

		token, err := validateToken(tokenString, accessIssuer, jwtSecret)
		if err != nil {
			respondWithError(w, 401, "Invalid token", err)
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

func handlePostRefresh(db *database.DB, jwtSecret []byte) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		rid := getRequestID(w)
		tokenString := strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer ")
		log.Println(rid, "handlePostRefresh", tokenString)

		revoked, err := db.TokenRevoked(tokenString)
		if err != nil {
			respondWithError(w, 500, "Potential database error", err)
			return
		}
		if revoked {
			respondWithError(w, 401, "Token is revoked", err)
			return
		}

		token, err := validateToken(tokenString, refreshIssuer, jwtSecret)
		if err != nil {
			respondWithError(w, 401, "Invalid token", err)
			return
		}

		var idStr string
		idStr, err = token.Claims.GetSubject()
		if err != nil {
			respondWithError(w, 401, "No user ID given", err)
			return
		}

		jwt, err := newToken(idStr, accessIssuer, expirationAccessSeconds, jwtSecret)
		if err != nil {
			respondWithError(w, 500, "Error creating access token", err)
		}

		type response struct {
			Token string `json:"token"`
		}

		respondWithJSON(w, 200, response{Token: jwt})
	})
}

func validateToken(tokenString string, requiredIssuer string, key []byte) (*jwt.Token, error) {
	// The keyFunc should take the parsed but unverified token, do any checks to make sure the token is of a valid format, and then return the signing key to verify the authenticity of the token against.
	return jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if issuer, err := token.Claims.GetIssuer(); issuer != requiredIssuer || err != nil {
			return nil, fmt.Errorf("Invalid token type %v, %v", issuer, err)
		}
		return key, nil
	}, jwt.WithValidMethods([]string{"HS256"}))
}

func checkUserAuthenticated(tokenString string, key []byte) (string, error) {
	token, err := validateToken(tokenString, accessIssuer, key)
	if err != nil {
		return "", err
	}
	return token.Claims.GetSubject()
}

func handlePostRevoke(db *database.DB, jwtSecret []byte) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		rid := getRequestID(w)
		tokenString := strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer ")
		log.Println(rid, "handlePostRevoke", tokenString)

		token, err := validateToken(tokenString, refreshIssuer, jwtSecret)
		if err != nil {
			respondWithError(w, 401, "Invalid token", err)
			return
		}
		issuedAt, err := token.Claims.GetIssuedAt()
		if err != nil {
			respondWithError(w, 401, "Invalid token", err)
			return
		}

		err = db.RevokeToken(tokenString, issuedAt.Time)
		if err != nil {
			respondWithError(w, 500, "Potential database error", err)
			return
		}

		w.WriteHeader(200)
	})
}
