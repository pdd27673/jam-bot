package commands

import (
	"context"
	"fmt"
	"jam-bot/internal/spotify"
	"strings"

	"github.com/bwmarrin/discordgo"
)

type QueueCommand struct {
	spotifyService *spotify.Service
}

func NewQueueCommand(spotifyService *spotify.Service) *QueueCommand {
	return &QueueCommand{spotifyService: spotifyService}
}

func (c *QueueCommand) Name() string {
	return "queue"
}

func (c *QueueCommand) Description() string {
	return "Displays the current song queue."
}

func (c *QueueCommand) Execute(s *discordgo.Session, m *discordgo.MessageCreate, args []string) error {
	ctx := context.Background()
	channelID := m.ChannelID

	if channelID == "" {
		_, err := s.ChannelMessageSend(m.ChannelID, "❌ This command can only be used within a channel.")
		if err != nil {
			return fmt.Errorf("failed to send message: %w", err)
		}
		return nil
	}

	// Retrieve the current session
	session, err := c.spotifyService.LoadSession(ctx, channelID)
	if err != nil {
		return fmt.Errorf("failed to load session: %w", err)
	}

	if len(session.Queue) == 0 {
		_, err := s.ChannelMessageSend(m.ChannelID, "🎶 The queue is currently empty.")
		if err != nil {
			return fmt.Errorf("failed to send message: %w", err)
		}
		return nil
	}

	// Build the queue list
	var queueList []string
	for idx, song := range session.Queue {
		queueList = append(queueList, fmt.Sprintf("%d. **%s** by **%s**", idx+1, song.Title, song.Artist))
	}

	queueMessage := "🎶 **Current Queue:**\n" + strings.Join(queueList, "\n")

	_, err = s.ChannelMessageSend(m.ChannelID, queueMessage)
	if err != nil {
		return fmt.Errorf("failed to send queue message: %w", err)
	}

	return nil
}
