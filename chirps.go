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

func handlePostChirps(db *database.DB, jwtSecret []byte) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		rid := getRequestID(w)
		tokenString := strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer ")
		idStr, err := checkUserAuthenticated(tokenString, jwtSecret)
		log.Println(rid, "handlePostChirps for AuthorId", idStr, tokenString)
		if err != nil {
			respondWithError(w, 401, "Authentication failed", err)
			return
		}

		authorId, err := strconv.Atoi(idStr)
		if err != nil {
			respondWithError(w, 401, "Given user ID is not a number", err)
			return
		}

		type parameters struct {
			Body string `json:"body"`
		}

		decoder := json.NewDecoder(r.Body)
		params := parameters{}
		err = decoder.Decode(&params)
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
		chirp, err := db.CreateChirp(body, authorId)
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

		if s := r.URL.Query().Get("author_id"); len(s) > 0 {
			authorId, err := strconv.Atoi(s)
			if err != nil {
				respondWithError(w, 404, "Given author id does not look like a number", err)
			}
			authorChirps := make([]database.Chirp, 0)
			for _, c := range chirps {
				if c.AuthorId == authorId {
					authorChirps = append(authorChirps, c)
				}
			}
			chirps = authorChirps

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

func handleDeleteChirp(db *database.DB, jwtSecret []byte) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		rid := getRequestID(w)

		tokenString := strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer ")
		idStr, err := checkUserAuthenticated(tokenString, jwtSecret)
		log.Println(rid, "handleDeleteChirp for AuthorId", idStr, tokenString)
		if err != nil {
			respondWithError(w, 401, "Authentication failed", err)
			return
		}

		authorId, err := strconv.Atoi(idStr)
		if err != nil {
			respondWithError(w, 401, "Given user ID is not a number", err)
			return
		}

		chirpId, err := strconv.Atoi(r.PathValue("id"))
		if err != nil {
			respondWithError(w, 401, "Given chirp ID is not a number", err)
			return
		}

		chirp, err := db.GetChirp(chirpId)
		if err != nil {
			respondWithError(w, 404, "Couldn't retrieve chirp", err)
		}
		if chirp.AuthorId != authorId {
			respondWithError(w, 403, "User not authenticated to delete this chirp", nil)
		}
		w.WriteHeader(200)
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
