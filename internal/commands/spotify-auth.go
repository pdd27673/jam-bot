package commands

import (
	"fmt"
	"jam-bot/internal/spotify"

	"github.com/bwmarrin/discordgo"
)

type SpotifyAuthCommand struct {
	spotifyService *spotify.Service
}

func NewSpotifyAuthCommand(spotifyService *spotify.Service) *SpotifyAuthCommand {
	return &SpotifyAuthCommand{spotifyService: spotifyService}
}

func (c *SpotifyAuthCommand) Name() string {
	return "auth"
}

func (c *SpotifyAuthCommand) Description() string {
	return "Authenticate with Spotify"
}

func (c *SpotifyAuthCommand) Execute(s *discordgo.Session, m *discordgo.MessageCreate, args []string) error {
	authURL := c.spotifyService.GetAuthURL(m.Author.ID)

	dm, err := s.UserChannelCreate(m.Author.ID)
	if err != nil {
		return fmt.Errorf("failed to create DM channel: %w", err)
	}

	_, err = s.ChannelMessageSend(dm.ID, fmt.Sprintf(
		"Please authenticate with Spotify by clicking this link:\n%s",
		authURL,
	))

	if err != nil {
		return fmt.Errorf("failed to send auth message: %w", err)
	}

	return nil
}
