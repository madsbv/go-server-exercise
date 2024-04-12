package main

import (
	"log"
	"net/http"
)

func main() {
	filepathRoot := "/app/"
	port := "8080"

	smux := http.NewServeMux()
	smux.Handle(filepathRoot, http.StripPrefix("/app", http.FileServer(http.Dir("."))))
	smux.HandleFunc("/healthz", healthz)

	corsMux := middlewareCors(smux)

	server := http.Server{
		Addr:    "localhost:" + port,
		Handler: corsMux,
	}
	log.Printf("Serving files from %s on port: %s\n", filepathRoot, port)
	log.Fatal(server.ListenAndServe())

}

func healthz(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
	log.Println("Health check request answered")
}

func middlewareCors(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS, PUT, DELETE")
		w.Header().Set("Access-Control-Allow-Headers", "*")
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}
		next.ServeHTTP(w, r)
	})
}
