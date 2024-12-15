package commands

import (
	"context"
	"fmt"
	"jam-bot/internal/spotify"

	"github.com/bwmarrin/discordgo"
)

type PauseCommand struct {
	spotifyService *spotify.Service
}

func NewPauseCommand(spotifyService *spotify.Service) *PauseCommand {
	return &PauseCommand{spotifyService: spotifyService}
}

func (c *PauseCommand) Name() string {
	return "pause"
}

func (c *PauseCommand) Description() string {
	return "Pauses the current playback."
}

func (c *PauseCommand) Execute(s *discordgo.Session, m *discordgo.MessageCreate, args []string) error {
	ctx := context.Background()
	channelID := m.ChannelID

	if channelID == "" {
		_, err := s.ChannelMessageSend(m.ChannelID, "❌ This command can only be used within a channel.")
		if err != nil {
			return fmt.Errorf("failed to send message: %w", err)
		}
		return nil
	}

	// Pause playback
	err := c.spotifyService.PausePlayback(ctx, channelID)
	if err != nil {
		_, sendErr := s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("❌ Failed to pause playback: %v", err))
		if sendErr != nil {
			return fmt.Errorf("failed to send error message: %w", sendErr)
		}
		return fmt.Errorf("failed to pause playback: %w", err)
	}

	// Confirm to the user
	_, err = s.ChannelMessageSend(m.ChannelID, "✅ Playback has been paused.")
	if err != nil {
		return fmt.Errorf("failed to send confirmation message: %w", err)
	}

	return nil
}
