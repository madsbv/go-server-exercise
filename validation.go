package main

import (
	"encoding/json"
	"log"
	"net/http"
)

func validateChirp(w http.ResponseWriter, r *http.Request) {
	// We always return JSON from this method
	w.Header().Set("Content-Type", "application/json")
	type parameters struct {
		Body string `json:"body"`
	}
	type returnValid struct {
		Valid bool `json:"valid"`
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

	// Chirp is valid, return 200
	respBody := returnValid{Valid: true}
	respondWithJSON(w, 200, respBody)
}
