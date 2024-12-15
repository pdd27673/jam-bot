package bot

import (
	"fmt"
	"jam-bot/internal/commands"
	"jam-bot/internal/config"
	"jam-bot/internal/spotify"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/bwmarrin/discordgo"
)

var dg *discordgo.Session
var cmdRegistry *commands.Registry

// sendDM sends a direct message to a user
func sendDM(userID, message string) error {
	channel, err := dg.UserChannelCreate(userID)
	if err != nil {
		return fmt.Errorf("failed to create DM channel: %w", err)
	}

	_, err = dg.ChannelMessageSend(channel.ID, message)
	if err != nil {
		return fmt.Errorf("failed to send DM message: %w", err)
	}

	return nil
}

// StartBot initializes the Discord session and starts the bot
func StartBot() error {
	cfg, err := config.LoadConfig()
	if err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	// Initialize Spotify service with sendDM function
	spotifyService := spotify.NewSpotifyService(cfg, sendDM)

	// Start auth server in a goroutine
	go func() {
		if err := spotifyService.StartAuthServer(8080); err != nil {
			log.Printf("[ERROR] Auth server failed: %v", err)
		}
	}()

	dg, err = discordgo.New("Bot " + cfg.DiscordToken)
	if err != nil {
		return fmt.Errorf("error creating Discord session: %w", err)
	}

	// Initialize and register commands
	cmdRegistry = commands.NewRegistry()
	cmdRegistry.Register(&commands.PingCommand{})
	cmdRegistry.Register(commands.NewHelpCommand(cmdRegistry))
	cmdRegistry.Register(commands.NewSpotifyAuthCommand(spotifyService))
	cmdRegistry.Register(commands.NewSpotifyStatusCommand(spotifyService))

	// Add message handler
	dg.AddHandler(func(s *discordgo.Session, m *discordgo.MessageCreate) {
		err := cmdRegistry.ExecuteCommand(s, m, cfg.BotPrefix)
		if err != nil {
			fmt.Println("[ERROR]", err)
		}
	})

	// Open a websocket connection to Discord and begin listening
	err = dg.Open()
	if err != nil {
		return fmt.Errorf("error opening connection to Discord: %w", err)
	}
	log.Println("[INFO] bot is now running. Press CTRL+C to exit.")

	// Wait until CTRL+C or other termination signal is received
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	<-sc

	// Cleanly close down the Discord session
	err = dg.Close()
	if err != nil {
		return fmt.Errorf("[ERROR] error closing connection to Discord: %s", err)
	}

	return nil
}
