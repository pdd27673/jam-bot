package commands

import (
	"context"
	"fmt"
	"jam-bot/internal/spotify"

	"github.com/bwmarrin/discordgo"
)

type LeaveCommand struct {
	spotifyService *spotify.Service
}

func NewLeaveCommand(spotifyService *spotify.Service) *LeaveCommand {
	return &LeaveCommand{spotifyService: spotifyService}
}

func (c *LeaveCommand) Name() string {
	return "leave"
}

func (c *LeaveCommand) Description() string {
	return "Removes you from the current jam session."
}

func (c *LeaveCommand) Execute(s *discordgo.Session, m *discordgo.MessageCreate, args []string) error {
	ctx := context.Background()
	guildID := m.GuildID
	userID := m.Author.ID

	// Remove the user from the session
	err := c.spotifyService.RemoveUserFromSession(ctx, guildID, userID)
	if err != nil {
		return fmt.Errorf("failed to remove user from session: %w", err)
	}

	// Confirm to the user via message in the channel
	_, err = s.ChannelMessageSend(m.ChannelID, "✅ You have left the jam session.")
	if err != nil {
		return fmt.Errorf("failed to send confirmation message: %w", err)
	}

	return nil
}
