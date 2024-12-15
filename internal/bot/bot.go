package bot

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"jam-bot/internal/commands"
	"jam-bot/internal/utils"

	"github.com/bwmarrin/discordgo"
	"github.com/joho/godotenv"
)

var dg *discordgo.Session
var cmdRegistry *commands.Registry

// StartBot initializes the discord session and starts the bot
func StartBot() error {
	err := godotenv.Load("local.env")
	if err != nil {
		log.Println("[INFO] no .env file found. Proceeding with environment variables.")
	}

	token := os.Getenv("DISCORD_TOKEN")
	if token == "" {
		return fmt.Errorf("[ERROR] DISCORD_TOKEN not found in environment variables")
	}

	dg, err = discordgo.New("Bot " + token)
	if err != nil {
		return fmt.Errorf("[ERROR] error creating Discord session: %s", err)
	}

	// Initialize and register commands
	cmdRegistry = commands.NewRegistry()
	cmdRegistry.Register(&commands.PingCommand{})
	cmdRegistry.Register(commands.NewHelpCommand(cmdRegistry))

	// Add message handler
	dg.AddHandler(func(s *discordgo.Session, m *discordgo.MessageCreate) {
		err := cmdRegistry.ExecuteCommand(s, m, utils.DISCORD_BOT_PREFIX)
		if err != nil {
			fmt.Println("[ERROR]", err)
		}
	})

	// Open a websocket connection to Discord and begin listening
	err = dg.Open()
	if err != nil {
		return fmt.Errorf("[ERROR] error opening connection to Discord: %s", err)
	}
	log.Println("[INFO] bot is now running. Press CTRL+C to exit.")

	// Wait until ctrl+c or other termination signal is received
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
	<-sc

	// Cleanly close down the Discord session
	err = dg.Close()
	if err != nil {
		return fmt.Errorf("[ERROR] error closing connection to Discord: %s", err)
	}

	return nil
}
