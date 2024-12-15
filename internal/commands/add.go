package commands

import (
	"context"
	"fmt"
	"jam-bot/internal/spotify"
	"strings"

	"github.com/bwmarrin/discordgo"
)

type AddCommand struct {
	spotifyService *spotify.Service
}

func NewAddCommand(spotifyService *spotify.Service) *AddCommand {
	return &AddCommand{spotifyService: spotifyService}
}

func (c *AddCommand) Name() string {
	return "add"
}

func (c *AddCommand) Description() string {
	return "Adds a song to the jam session queue. Usage: !add [song name]"
}

func (c *AddCommand) Execute(s *discordgo.Session, m *discordgo.MessageCreate, args []string) error {
	ctx := context.Background()
	channelID := m.ChannelID

	if channelID == "" {
		_, err := s.ChannelMessageSend(m.ChannelID, "❌ This command can only be used within a channel.")
		if err != nil {
			return fmt.Errorf("failed to send message: %w", err)
		}
		return nil
	}

	if len(args) == 0 {
		_, err := s.ChannelMessageSend(m.ChannelID, "❌ Please provide a song name. Usage: `!add [song name]`")
		if err != nil {
			return fmt.Errorf("failed to send usage message: %w", err)
		}
		return nil
	}

	songName := strings.Join(args, " ")

	// Search for the song using Spotify API
	song, err := c.spotifyService.SearchSong(ctx, m.Author.ID, songName)
	if err != nil {
		_, sendErr := s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("❌ Failed to find the song: %v", err))
		if sendErr != nil {
			return fmt.Errorf("failed to send error message: %w", sendErr)
		}
		return fmt.Errorf("failed to search song: %w", err)
	}

	// Add the song to the session queue
	err = c.spotifyService.AddSongToQueue(ctx, channelID, song) // Updated to use channelID
	if err != nil {
		_, sendErr := s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("❌ Failed to add the song to the queue: %v", err))
		if sendErr != nil {
			return fmt.Errorf("failed to send error message: %w", sendErr)
		}
		return fmt.Errorf("failed to add song to queue: %w", err)
	}

	// Confirm to the user
	_, err = s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("✅ **%s** by **%s** has been added to the queue.", song.Title, song.Artist))
	if err != nil {
		return fmt.Errorf("failed to send confirmation message: %w", err)
	}

	return nil
}
