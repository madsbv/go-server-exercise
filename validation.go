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
	type returnErr struct {
		Error string `json:"error"`
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
		respErr := returnErr{
			Error: "Failed to decode request",
		}

		dat, merr := json.Marshal(respErr)
		if merr != nil {
			log.Fatal("Failed to marshal static data, this should never happen")
		}

		w.WriteHeader(500)
		w.Write(dat)
		return
	}

	if l := len(params.Body); l > 140 {
		log.Printf("Received chirp with %d > 140 characters, rejected", l)
		respErr := returnErr{
			Error: "Chirp is too long",
		}

		dat, merr := json.Marshal(respErr)
		if merr != nil {
			log.Fatal("Failed to marshal static data, this should never happen")
		}

		w.WriteHeader(400)
		w.Write(dat)
		return
	}

	// Chirp is valid, return 200
	respBody := returnValid{Valid: true}
	dat, merr := json.Marshal(respBody)
	if merr != nil {
		log.Fatal("Failed to marshal static data, this should never happen")
	}
	w.WriteHeader(200)
	w.Write(dat)
}
