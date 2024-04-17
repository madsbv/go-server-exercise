package main

import (
	"fmt"
	"io"
	"net/http"
)

func (cfg *apiConfig) middlewareMetricsInc(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, h *http.Request) {
		cfg.fileserverHits += 1
		next.ServeHTTP(w, h)
	},
	)
}

func (cfg *apiConfig) metrics(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	io.WriteString(w, fmt.Sprintf("Hits: %v", cfg.fileserverHits))
}

func (cfg *apiConfig) reset(_ http.ResponseWriter, _ *http.Request) {
	cfg.fileserverHits = 0
}
