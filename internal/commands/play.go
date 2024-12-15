package commands

import (
	"context"
	"fmt"
	"jam-bot/internal/spotify"

	"github.com/bwmarrin/discordgo"
)

type PlayCommand struct {
	spotifyService *spotify.Service
}

func NewPlayCommand(spotifyService *spotify.Service) *PlayCommand {
	return &PlayCommand{spotifyService: spotifyService}
}

func (c *PlayCommand) Name() string {
	return "play"
}

func (c *PlayCommand) Description() string {
	return "Starts playback of the queued songs."
}

func (c *PlayCommand) Execute(s *discordgo.Session, m *discordgo.MessageCreate, args []string) error {
	ctx := context.Background()
	channelID := m.ChannelID

	if channelID == "" {
		_, err := s.ChannelMessageSend(m.ChannelID, "❌ This command can only be used within a channel.")
		if err != nil {
			return fmt.Errorf("failed to send message: %w", err)
		}
		return nil
	}

	// Start playback
	err := c.spotifyService.StartPlayback(ctx, channelID)
	if err != nil {
		_, sendErr := s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("❌ Failed to start playback: %v", err))
		if sendErr != nil {
			return fmt.Errorf("failed to send error message: %w", sendErr)
		}
		return fmt.Errorf("failed to start playback: %w", err)
	}

	// Confirm to the user
	_, err = s.ChannelMessageSend(m.ChannelID, "✅ Playback has started.")
	if err != nil {
		return fmt.Errorf("failed to send confirmation message: %w", err)
	}

	return nil
}
