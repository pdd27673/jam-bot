package bot

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/bwmarrin/discordgo"
	"github.com/joho/godotenv"
)

var dg *discordgo.Session

// StartBot initializes the discord session and starts the bot
func StartBot() error {
	// load env variables from .env file
	err := godotenv.Load()
	if err != nil {
		log.Println("[INFO] no .env file found. Proceeding with environment variables.")
	}

	token := os.Getenv("DISCORD_TOKEN")
	if token == "" {
		return fmt.Errorf("[ERROR] DISCORD_TOKEN not found in environment variables")
	}

	// create a new discord session
	dg, err = discordgo.New("Bot " + token)
	if err != nil {
		return fmt.Errorf("[ERROR] error creating Discord session: %s", err)
	}

	// register messageCreate as a callback for the messageCreate events
	dg.AddHandler(messageCreate)

	// open a websocket connection to Discord and begin listening
	err = dg.Open()
	if err != nil {
		return fmt.Errorf("[ERROR] error opening connection to Discord: %s", err)
	}
	log.Println("[INFO] bot is now running. Press CTRL+C to exit.")

	// wait until ctrl+c or other termination signal is received
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
	<-sc

	// cleanly close down the Discord session
	err = dg.Close()
	if err != nil {
		return fmt.Errorf("[ERROR] error closing connection to Discord: %s", err)
	}

	return nil
}