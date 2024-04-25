package main

import (
	"encoding/json"
	"log"
	"net/http"
)

func respondWithError(w http.ResponseWriter, code int, msg string, err error) {
	rid := getRequestID(w)
	log.Println(rid, "Responding with error message", msg)
	if err != nil {
		log.Println(rid, "\033[31;1mError:\033[0m", err)
	}

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
	rid := getRequestID(w)
	dat, err := json.Marshal(payload)
	if err != nil {
		respondWithError(w, 500, "Error encoding response payload", err)
	}
	log.Println(rid, "Responding with json payload and status code", code)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	w.Write(dat)
}
