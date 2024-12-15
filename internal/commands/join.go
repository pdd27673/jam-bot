package commands

import (
	"context"
	"fmt"
	"jam-bot/internal/spotify"

	"github.com/bwmarrin/discordgo"
)

type JoinCommand struct {
	spotifyService *spotify.Service
}

func NewJoinCommand(spotifyService *spotify.Service) *JoinCommand {
	return &JoinCommand{spotifyService: spotifyService}
}

func (c *JoinCommand) Name() string {
	return "join"
}

func (c *JoinCommand) Description() string {
	return "Connects you to the current jam session."
}

func (c *JoinCommand) Execute(s *discordgo.Session, m *discordgo.MessageCreate, args []string) error {
	ctx := context.Background()
	channelID := m.ChannelID
	userID := m.Author.ID

	if channelID == "" {
		_, err := s.ChannelMessageSend(m.ChannelID, "❌ This command can only be used within a channel.")
		if err != nil {
			return fmt.Errorf("failed to send message: %w", err)
		}
		return nil
	}

	// Check if the user is authenticated
	isAuth, err := c.spotifyService.IsAuthenticated(ctx, userID)
	if err != nil {
		return fmt.Errorf("failed to check authentication status: %w", err)
	}

	if !isAuth {
		// Inform the user to authenticate first
		dm, err := s.UserChannelCreate(userID)
		if err != nil {
			return fmt.Errorf("failed to create DM channel: %w", err)
		}

		_, err = s.ChannelMessageSend(dm.ID, "You need to authenticate with Spotify first. Use `!auth` to authenticate.")
		if err != nil {
			return fmt.Errorf("failed to send message: %w", err)
		}

		return nil
	}

	// Add the user to the session
	err = c.spotifyService.AddUserToSession(ctx, channelID, userID)
	if err != nil {
		// If user is already in session, inform them
		if err.Error() == "user is already in the session" {
			_, err = s.ChannelMessageSend(m.ChannelID, "You are already part of the jam session!")
			if err != nil {
				return fmt.Errorf("failed to send message: %w", err)
			}
			return nil
		}
		return fmt.Errorf("failed to add user to session: %w", err)
	}

	// Confirm to the user via message in the channel
	_, err = s.ChannelMessageSend(m.ChannelID, "✅ You have joined the jam session!")
	if err != nil {
		return fmt.Errorf("failed to send confirmation message: %w", err)
	}

	return nil
}
