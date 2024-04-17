package main

import (
	"log"
	"net/http"
)

func main() {
	filepathRoot := "/app/"
	port := "8080"

	smux := http.NewServeMux()
	apiCfg := apiConfig{}
	smux.Handle(filepathRoot, apiCfg.middlewareMetricsInc(http.StripPrefix("/app", http.FileServer(http.Dir(".")))))
	smux.HandleFunc("/healthz", healthz)
	smux.HandleFunc("/metrics", apiCfg.metrics)
	smux.HandleFunc("/reset", apiCfg.reset)

	corsMux := middlewareCors(smux)

	server := http.Server{
		Addr:    "localhost:" + port,
		Handler: corsMux,
	}
	log.Printf("Serving files from %s on port: %s\n", filepathRoot, port)
	log.Fatal(server.ListenAndServe())
}

type apiConfig struct {
	fileserverHits int
}
