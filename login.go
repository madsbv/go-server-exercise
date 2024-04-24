package main

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/madsbv/boot-dev-go-server/internal/database"
)

func handleLogin(db *database.DB) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// We always return JSON from this method
		w.Header().Set("Content-Type", "application/json")
		type parameters struct {
			Email    string `json:"email"`
			Password string `json:"password"`
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

		respondWithJSON(w, 200, user)
	})
}
