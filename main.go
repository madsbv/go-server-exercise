package main

import (
	"flag"
	"log"
	"net/http"
	"os"
	"sync"

	"github.com/joho/godotenv"
	"github.com/madsbv/boot-dev-go-server/internal/database"
)

func main() {
	filepathRoot := "/app/"
	dbg := flag.Bool("debug", false, "Enable debug mode")
	flag.Parse()
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	jwtSecret := []byte(os.Getenv("JWT_SECRET"))
	apiCfg := apiConfig{}
	apiCfg.jwtSecret = jwtSecret

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

	logger := log.New(os.Stdout, "http: ", log.LstdFlags)
	logger.Printf("Serving files from %s on port: %s\n", filepathRoot, port)

	smux := initRoutes(db, &apiCfg, filepathRoot)

	server := http.Server{
		Addr:     "localhost:" + port,
		Handler:  tracing(logging(logger, middlewareCors(smux))),
		ErrorLog: logger,
	}
	logger.Fatal(server.ListenAndServe())
}

type apiConfig struct {
	fileserverHits struct {
		count int
		mux   sync.RWMutex
	}
	jwtSecret []byte
}
