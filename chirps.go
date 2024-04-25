package main

import (
	"encoding/json"
	"log"
	"net/http"
	"slices"
	"strconv"
	"strings"

	"github.com/madsbv/boot-dev-go-server/internal/database"
)

type Chirp = database.Chirp

func handlePostChirps(db *database.DB) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		type parameters struct {
			Body string `json:"body"`
		}

		decoder := json.NewDecoder(r.Body)
		params := parameters{}
		err := decoder.Decode(&params)
		log.Printf("Handling: %s", params.Body)
		if err != nil {
			log.Printf("Error decoding chirp parameters: %s", err)
			respondWithError(w, 500, "Failed to decode request", err)
			return
		}

		if l := len(params.Body); l > 140 {
			log.Printf("Received chirp with %d > 140 characters, rejected", l)
			respondWithError(w, 400, "Chirp is too long", err)
			return
		}

		// Chirp has valid length, proceed to clean it up
		body := cleanBadWords(strings.TrimSpace(params.Body))
		chirp, err := db.CreateChirp(body)
		if err != nil {
			log.Printf("Database error when creating chirp: %v", err)
			respondWithError(w, 500, "Error handling request", err)
			return
		}

		respondWithJSON(w, 201, chirp)
	})
}

func handleGetAllChirps(db *database.DB) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		chirps, err := db.GetSortedChirps()
		if err != nil {
			log.Println("Error getting list of chirps")
			respondWithError(w, 500, "Error handling request", err)
		}
		respondWithJSON(w, 200, chirps)
	})
}

func handleGetChirp(db *database.DB) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestedId := r.PathValue("id")
		id, err := strconv.Atoi(requestedId)
		if err != nil {
			log.Printf("Error serving GetChirp request for requested id %v: Looks like it is not an integer", requestedId)
		}

		chirp, err := db.GetChirp(id)
		if err != nil {
			log.Println("Error getting chirp with id", id, err)
			// NOTE: It might be worth distinguishing between internal database error, and invalid id in request. How to do that?
			respondWithError(w, 404, "Error handling request", err)
		}

		respondWithJSON(w, 200, chirp)
	})
}

func cleanBadWords(body string) string {
	// Replace the following words with '****'
	badWords := []string{"kerfuffle", "sharbert", "fornax"}

	words := strings.Split(body, " ")
	for i, w := range words {
		if slices.Contains(badWords, strings.ToLower(w)) {
			words[i] = "****"
		}
	}
	return strings.Join(words, " ")
}
