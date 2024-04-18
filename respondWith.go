package main

import (
	"encoding/json"
	"log"
	"net/http"
)

func respondWithError(w http.ResponseWriter, code int, msg string) {
	type returnErr struct {
		Error string `json:"error"`
	}
	respErr := returnErr{
		Error: msg,
	}

	dat, merr := json.Marshal(respErr)
	if merr != nil {
		log.Fatal("Failed to marshal statically structured data, this should never happen")
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	w.Write(dat)
}

func respondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	dat, err := json.Marshal(payload)
	if err != nil {
		log.Printf("Error marshalling some payload")
		respondWithError(w, 500, "Something went wrong")
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	w.Write(dat)
}
