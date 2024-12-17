package main

import (
	"jam-bot/internal/bot"
	"log"
)

func main() {
	// cfg, err := config.LoadConfig()
	// if err != nil {
	// 	log.Fatalf("[ERROR] Failed to load config: %v", err)
	// }

	// Initialize Spotify service
	// spotifyService := spotify.NewSpotifyService(cfg, bot.SendDM)

	err := bot.StartBot()
	if err != nil {
		log.Fatalf("[ERROR] Error starting the bot: %v", err)
	}

	// Start the unified HTTP server
	// server.StartServer(cfg, spotifyService)
}
