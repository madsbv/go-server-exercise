package main

import (
	"net/http"

	"github.com/madsbv/boot-dev-go-server/internal/database"
)

func initRoutes(db *database.DB, apiCfg *apiConfig, filepathRoot string) *http.ServeMux {
	smux := http.NewServeMux()

	smux.Handle(filepathRoot, apiCfg.middlewareMetricsInc(http.FileServer(http.Dir("."))))

	smux.HandleFunc("GET /api/healthz", healthz)
	smux.HandleFunc("GET /admin/metrics", apiCfg.metrics)
	smux.HandleFunc("GET /api/reset", apiCfg.reset)

	smux.Handle("POST /api/chirps", handlePostChirps(db, apiCfg.jwtSecret))
	smux.Handle("GET /api/chirps", handleGetAllChirps(db))
	smux.Handle("GET /api/chirps/{id}", handleGetChirp(db))
	smux.Handle("DELETE /api/chirps/{id}", handleDeleteChirp(db, apiCfg.jwtSecret))

	smux.Handle("POST /api/users", handlePostUsers(db))
	smux.Handle("GET /api/users", handleGetAllUsers(db))
	smux.Handle("GET /api/users/{id}", handleGetUser(db))
	smux.Handle("PUT /api/users", handlePutUsers(db, apiCfg.jwtSecret))

	smux.Handle("POST /api/login", handlePostLogin(db, apiCfg.jwtSecret))
	smux.Handle("POST /api/refresh", handlePostRefresh(db, apiCfg.jwtSecret))
	smux.Handle("POST /api/revoke", handlePostRevoke(db, apiCfg.jwtSecret))

	smux.Handle("POST /api/polka/webhooks", handlePostPolkaWebhooks(db, apiCfg.polkaSecret))

	return smux
}
