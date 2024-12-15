package commands

import (
	"context"
	"fmt"
	"jam-bot/internal/spotify"
	"strconv"

	"github.com/bwmarrin/discordgo"
)

type RemoveCommand struct {
	spotifyService *spotify.Service
}

func NewRemoveCommand(spotifyService *spotify.Service) *RemoveCommand {
	return &RemoveCommand{spotifyService: spotifyService}
}

func (c *RemoveCommand) Name() string {
	return "remove"
}

func (c *RemoveCommand) Description() string {
	return "Removes a song from the queue by its position. Usage: !remove [position]"
}

func (c *RemoveCommand) Execute(s *discordgo.Session, m *discordgo.MessageCreate, args []string) error {
	ctx := context.Background()
	channelID := m.ChannelID

	if channelID == "" {
		_, err := s.ChannelMessageSend(m.ChannelID, "❌ This command can only be used within a channel.")
		if err != nil {
			return fmt.Errorf("failed to send message: %w", err)
		}
		return nil
	}

	if len(args) != 1 {
		_, err := s.ChannelMessageSend(m.ChannelID, "❌ Please provide the position of the song to remove. Usage: `!remove [position]`")
		if err != nil {
			return fmt.Errorf("failed to send usage message: %w", err)
		}
		return nil
	}

	position, err := strconv.Atoi(args[0])
	if err != nil || position < 1 {
		_, err := s.ChannelMessageSend(m.ChannelID, "❌ Invalid position. Please provide a positive integer.")
		if err != nil {
			return fmt.Errorf("failed to send error message: %w", err)
		}
		return nil
	}

	// Remove the song from the queue
	err = c.spotifyService.RemoveSongFromQueue(ctx, channelID, position-1)
	if err != nil {
		_, sendErr := s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("❌ Failed to remove the song: %v", err))
		if sendErr != nil {
			return fmt.Errorf("failed to send error message: %w", sendErr)
		}
		return fmt.Errorf("failed to remove song: %w", err)
	}

	// Confirm to the user
	_, err = s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("✅ Song at position **%d** has been removed from the queue.", position))
	if err != nil {
		return fmt.Errorf("failed to send confirmation message: %w", err)
	}

	return nil
}
