package main

import (
	"log"
	"net/http"
	"sync"
)

func main() {
	filepathRoot := "/app/"
	port := "8080"

	smux := http.NewServeMux()
	apiCfg := apiConfig{}
	smux.Handle(filepathRoot, apiCfg.middlewareMetricsInc(http.StripPrefix("/app", http.FileServer(http.Dir(".")))))
	smux.HandleFunc("GET /api/healthz", healthz)
	smux.HandleFunc("GET /admin/metrics", apiCfg.metrics)
	smux.HandleFunc("GET /api/reset", apiCfg.reset)
	smux.HandleFunc("POST /api/validate_chirp", validateChirp)

	corsMux := middlewareCors(smux)

	server := http.Server{
		Addr:    "localhost:" + port,
		Handler: corsMux,
	}
	log.Printf("Serving files from %s on port: %s\n", filepathRoot, port)
	log.Fatal(server.ListenAndServe())
}

type apiConfig struct {
	fileserverHits struct {
		count int
		mux   sync.RWMutex
	}
}
