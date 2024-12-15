package commands

import (
	"context"
	"fmt"
	"jam-bot/internal/spotify"

	"github.com/bwmarrin/discordgo"
)

type SpotifyStatusCommand struct {
	spotifyService *spotify.Service
}

func NewSpotifyStatusCommand(spotifyService *spotify.Service) *SpotifyStatusCommand {
	return &SpotifyStatusCommand{spotifyService: spotifyService}
}

func (c *SpotifyStatusCommand) Name() string {
	return "status"
}

func (c *SpotifyStatusCommand) Description() string {
	return "Check Spotify authentication status"
}

func (c *SpotifyStatusCommand) Execute(s *discordgo.Session, m *discordgo.MessageCreate, args []string) error {
	ctx := context.Background()
	isAuth, err := c.spotifyService.IsAuthenticated(ctx, m.Author.ID)
	if err != nil {
		return fmt.Errorf("failed to check authentication status: %w", err)
	}

	dm, err := s.UserChannelCreate(m.Author.ID)
	if err != nil {
		return fmt.Errorf("failed to create DM channel: %w", err)
	}

	if isAuth {
		_, err = s.ChannelMessageSend(dm.ID, "You are already authenticated with Spotify!")
	} else {
		_, err = s.ChannelMessageSend(dm.ID, "You are not authenticated with Spotify. Use `!auth` to authenticate.")
	}

	if err != nil {
		return fmt.Errorf("failed to send status message: %w", err)
	}

	return nil
}
