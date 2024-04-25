package main

import (
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"
)

// Log information about every HTTP call, identified by the requestID
func logging(logger *log.Logger, next http.Handler) http.Handler {
	start := time.Now()
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestID := getRequestID(w)
		if requestID == "" {
			requestID = "unknown"
		}
		logger.Println(requestID, r.Method, r.URL.Path, time.Since(start))
		// logger.Println(requestID, r.Method, r.URL.Path, r.RemoteAddr, r.UserAgent(), time.Since(start))
		next.ServeHTTP(w, r)
	})
}

// Ensure that every request has a unique requestID available so we can keep track of individual requests
func tracing(next http.Handler) http.Handler {
	rg := requestIDGenerator{}
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestID := r.Header.Get("X-Request-Id")
		if requestID == "" {
			requestID = rg.getID()
		}
		w.Header().Set("X-Request-Id", requestID)
		next.ServeHTTP(w, r)
	})
}

// Locking on a single mutex here is not great, but it should work for now. The blocking operation is very fast.
type requestIDGenerator struct {
	nextID int
	mux    sync.Mutex
}

func (rg *requestIDGenerator) getID() string {
	rg.mux.Lock()
	defer rg.mux.Unlock()
	newID := rg.nextID
	rg.nextID++
	return fmt.Sprint(newID)
}

func getRequestID(w http.ResponseWriter) string {
	return w.Header().Get("X-Request-Id")
}
