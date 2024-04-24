package main

import (
	"flag"
	"log"
	"net/http"
	"os"
	"sync"

	"github.com/madsbv/boot-dev-go-server/internal/database"
)

func main() {
	dbg := flag.Bool("debug", false, "Enable debug mode")
	flag.Parse()

	filepathRoot := "/app/"
	port := "8080"
	dbPath := "database.json"

	if *dbg {
		log.Println("In debug mode: Deleting existing database for testing purposes")
		_ = os.Remove(dbPath)
	}

	db, err := database.NewDB(dbPath)
	if err != nil {
		log.Fatal("Failed to create database connection: ", err)
	}

	smux := http.NewServeMux()
	apiCfg := apiConfig{}
	smux.Handle(filepathRoot, apiCfg.middlewareMetricsInc(http.FileServer(http.Dir("."))))
	smux.HandleFunc("GET /api/healthz", healthz)
	smux.HandleFunc("GET /admin/metrics", apiCfg.metrics)
	smux.HandleFunc("GET /api/reset", apiCfg.reset)
	smux.Handle("POST /api/chirps", handlePostChirps(db))
	smux.Handle("GET /api/chirps", handleGetAllChirps(db))
	smux.Handle("GET /api/chirps/{id}", handleGetChirp(db))
	smux.Handle("POST /api/users", handlePostUsers(db))
	smux.Handle("GET /api/users", handleGetAllUsers(db))
	smux.Handle("GET /api/users/{id}", handleGetUser(db))

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
