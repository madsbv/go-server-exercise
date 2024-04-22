package main

import (
	"encoding/json"
	"log"
	"net/http"
	"slices"
	"strings"
)

func handleValidateChirp(w http.ResponseWriter, r *http.Request) {
	// We always return JSON from this method
	w.Header().Set("Content-Type", "application/json")
	type parameters struct {
		Body string `json:"body"`
	}
	type returnValid struct {
		CleanedBody string `json:"cleaned_body"`
	}

	decoder := json.NewDecoder(r.Body)
	params := parameters{}
	err := decoder.Decode(&params)
	log.Printf("Handling: %s", params.Body)
	if err != nil {
		log.Printf("Error decoding chirp parameters: %s", err)
		respondWithError(w, 500, "Failed to decode request")
		return
	}

	if l := len(params.Body); l > 140 {
		log.Printf("Received chirp with %d > 140 characters, rejected", l)
		respondWithError(w, 400, "Chirp is too long")
		return
	}

	// Chirp has valid length, proceed to clean it up
	// Replace the following with '****': kerfuffle, sharbert, fornax
	badWords := []string{"kerfuffle", "sharbert", "fornax"}

	body := strings.TrimSpace(params.Body)
	words := strings.Split(body, " ")
	for i, w := range words {
		if slices.Contains(badWords, strings.ToLower(w)) {
			words[i] = "****"
		}
	}
	CleanedBody := strings.Join(words, " ")

	respBody := returnValid{CleanedBody}
	respondWithJSON(w, 200, respBody)
}
