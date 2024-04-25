package main

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"

	"github.com/madsbv/boot-dev-go-server/internal/database"
)

func handlePostPolkaWebhooks(db *database.DB, polkaSecret string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		rid := getRequestID(w)
		tokenString := strings.TrimPrefix(r.Header.Get("Authorization"), "ApiKey ")
		type parameters struct {
			Event string `json:"event"`
			Data  struct {
				UserId int `json:"user_id"`
			}
		}
		decoder := json.NewDecoder(r.Body)
		params := parameters{}
		err := decoder.Decode(&params)
		log.Println(rid, "handlePostPolkaWebhooks", params.Event, params.Data.UserId, tokenString)
		if err != nil {
			respondWithError(w, 500, "Failed to decode request body", err)
			return
		}

		if params.Event != "user.upgraded" {
			w.WriteHeader(200)
			log.Println(rid, "Non-upgrade event detected")
			return
		}

		// TODO
		auth := true

		if tokenString != polkaSecret && auth {
			respondWithError(w, 401, "Invalid Polka key", nil)
			return
		}

		err = db.UpgradeUser(params.Data.UserId)
		if err != nil {
			respondWithError(w, 404, "User not found", err)
			return
		}
		log.Println(rid, "Upgraded user", params.Data.UserId)
		w.WriteHeader(200)
		return
	})
}
