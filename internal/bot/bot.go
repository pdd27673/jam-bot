package bot

import (
	"context"
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
	cmdRegistry.Register(commands.NewJoinCommand(spotifyService))
	cmdRegistry.Register(commands.NewLeaveCommand(spotifyService))
	cmdRegistry.Register(commands.NewAddCommand(spotifyService))
	cmdRegistry.Register(commands.NewQueueCommand(spotifyService))
	cmdRegistry.Register(commands.NewUsersCommand(spotifyService))
	cmdRegistry.Register(commands.NewPlayCommand(spotifyService))
	cmdRegistry.Register(commands.NewPauseCommand(spotifyService))
	cmdRegistry.Register(commands.NewRemoveCommand(spotifyService))

	// Load and validate existing sessions
	err = loadAndValidateSessions(spotifyService)
	if err != nil {
		log.Printf("[ERROR] Failed to load and validate sessions: %v", err)
	}

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

// loadAndValidateSessions loads all existing sessions and validates participant authentication
func loadAndValidateSessions(spotifyService *spotify.Service) error {
	ctx := context.Background()
	sessions, err := spotifyService.LoadAllSessions(ctx)
	if err != nil {
		return fmt.Errorf("failed to load sessions: %w", err)
	}

	for _, session := range sessions {
		channelID := session.ChannelID
		for _, userID := range session.Participants {
			isAuth, err := spotifyService.IsAuthenticated(ctx, userID)
			if err != nil {
				log.Printf("[ERROR] Failed to check authentication for user %s: %v", userID, err)
				continue
			}

			if !isAuth {
				// Remove user from session
				err = spotifyService.RemoveUserFromSession(ctx, channelID, userID)
				if err != nil {
					log.Printf("[ERROR] Failed to remove unauthenticated user %s from session: %v", userID, err)
					continue
				}

				// Notify the user via DM
				err = spotifyService.SendDM(userID, "❌ You have been removed from the jam session because you are no longer authenticated with Spotify. Please re-authenticate using `!auth` if you wish to join again.")
				if err != nil {
					log.Printf("[ERROR] Failed to send DM to user %s: %v", userID, err)
				}

				log.Printf("[INFO] Removed unauthenticated user %s from channel %s session.", userID, channelID)
			}
		}
	}

	log.Println("[INFO] All sessions loaded and validated successfully.")
	return nil
}
