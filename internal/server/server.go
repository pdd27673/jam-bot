package server

import (
	"jam-bot/internal/config"
	"log"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
)

func StartServer(cfg *config.Config) {
	router := mux.NewRouter()

	router.HandleFunc("/api/v1/health", healthCheckHandler).Methods(http.MethodGet)
	router.HandleFunc("/api/v1/docs", documentationHandler).Methods(http.MethodGet)
	port := cfg.ServerPort

	log.Printf("[INFO] Starting HTTP server on port %s\n", strconv.Itoa(port))
	if err := http.ListenAndServe(":"+strconv.Itoa(port), router); err != nil {
		log.Fatalf("[ERROR] HTTP server failed: %v", err)
	}
}
