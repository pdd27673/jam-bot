package server

import (
	"jam-bot/internal/config"
	"jam-bot/internal/spotify"
	"log"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
)

func StartServer(cfg *config.Config, spotifyService *spotify.Service) {
	router := mux.NewRouter()

	// Register all handlers
	router.HandleFunc("/api/v1/health", healthCheckHandler).Methods(http.MethodGet)
	router.HandleFunc("/api/v1/docs", documentationHandler).Methods(http.MethodGet)
	router.HandleFunc("/callback", spotifyService.CallbackHandler).Methods(http.MethodGet)

	port := cfg.Port

	log.Printf("[INFO] Starting unified HTTP server on port %s\n", strconv.Itoa(port))
	if err := http.ListenAndServe(":"+strconv.Itoa(port), router); err != nil {
		log.Fatalf("[ERROR] HTTP server failed: %v", err)
	}
}
