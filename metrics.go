package main

import (
	"fmt"
	"io"
	"net/http"
)

func (cfg *apiConfig) middlewareMetricsInc(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, h *http.Request) {
		cfg.fileserverHits.mux.Lock()
		defer cfg.fileserverHits.mux.Unlock()
		cfg.fileserverHits.count += 1
		next.ServeHTTP(w, h)
	},
	)
}

func (cfg *apiConfig) metrics(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	cfg.fileserverHits.mux.RLock()
	defer cfg.fileserverHits.mux.RUnlock()
	io.WriteString(w, fmt.Sprintf(`<html>

<body>
    <h1>Welcome, Chirpy Admin</h1>
    <p>Chirpy has been visited %d times!</p>
</body>

</html>
`, cfg.fileserverHits.count))
}

func (cfg *apiConfig) reset(_ http.ResponseWriter, _ *http.Request) {
	cfg.fileserverHits.mux.Lock()
	defer cfg.fileserverHits.mux.Unlock()
	cfg.fileserverHits.count = 0
}
