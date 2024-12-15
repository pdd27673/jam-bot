package bot

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"jam-bot/internal/commands"
	"jam-bot/internal/config"
	"jam-bot/internal/spotify"
	"jam-bot/internal/utils"

	"github.com/bwmarrin/discordgo"
)

var dg *discordgo.Session
var cmdRegistry *commands.Registry

// StartBot initializes the discord session and starts the bot
func StartBot() error {
	cfg, err := config.LoadConfig()
	if err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	if cfg.DiscordToken == "" {
		return fmt.Errorf("discord token not provided")
	}

	dg, err = discordgo.New("Bot " + cfg.DiscordToken)
	if err != nil {
		return fmt.Errorf("error creating Discord session: %w", err)
	}

	spotifyService := spotify.NewSpotifyService(
		cfg.SpotifyClientID,
		cfg.SpotifyClientSecret,
		cfg.SpotifyRedirectURI,
	)

	// start auth server in a separate goroutine
	go func() {
		if err := spotifyService.StartAuthServer(8080); err != nil {
			log.Printf("[ERROR] failed to start auth server: %s", err)
		}
	}()

	// Initialize and register commands
	cmdRegistry = commands.NewRegistry()
	cmdRegistry.Register(&commands.PingCommand{})
	cmdRegistry.Register(commands.NewHelpCommand(cmdRegistry))
	cmdRegistry.Register(commands.NewSpotifyAuthCommand(spotifyService))

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
		return fmt.Errorf("error opening connection to Discord: %s", err)
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
