package main

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"

	"github.com/madsbv/boot-dev-go-server/internal/database"
)

type User = database.User

func handlePostUsers(db *database.DB) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// We always return JSON from this method
		w.Header().Set("Content-Type", "application/json")
		type parameters struct {
			Email string `json:"email"`
		}

		decoder := json.NewDecoder(r.Body)
		params := parameters{}
		err := decoder.Decode(&params)
		log.Printf("Handling: %s", params.Email)
		if err != nil {
			log.Printf("Error decoding user parameters: %s", err)
			respondWithError(w, 500, "Failed to decode request")
			return
		}

		user, err := db.CreateUser(params.Email)
		if err != nil {
			log.Printf("Database error when creating user: %v", err)
			respondWithError(w, 500, "Error handling request")
			return
		}

		respondWithJSON(w, 201, user)
	})
}

func handleGetAllUsers(db *database.DB) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		users, err := db.GetSortedUsers()
		if err != nil {
			log.Println("Error getting list of users")
			respondWithError(w, 500, "Error handling request")
		}
		respondWithJSON(w, 200, users)
	})
}

func handleGetUser(db *database.DB) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestedId := r.PathValue("id")
		id, err := strconv.Atoi(requestedId)
		if err != nil {
			log.Printf("Error serving GetUser request for requested id %v: Looks like it is not an integer", requestedId)
		}

		user, err := db.GetUser(id)
		if err != nil {
			log.Println("Error getting user with id", id, err)
			// NOTE: It might be worth distinguishing between internal database error, and invalid id in request. How to do that?
			respondWithError(w, 404, "Error handling request")
		}

		respondWithJSON(w, 200, user)
	})
}
