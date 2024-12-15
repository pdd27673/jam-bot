package commands

import (
	"context"
	"fmt"
	"jam-bot/internal/spotify"

	"github.com/bwmarrin/discordgo"
)

type UsersCommand struct {
	spotifyService *spotify.Service
}

func NewUsersCommand(spotifyService *spotify.Service) *UsersCommand {
	return &UsersCommand{spotifyService: spotifyService}
}

func (c *UsersCommand) Name() string {
	return "users"
}

func (c *UsersCommand) Description() string {
	return "Lists all users currently in the jam session."
}

func (c *UsersCommand) Execute(s *discordgo.Session, m *discordgo.MessageCreate, args []string) error {
	ctx := context.Background()
	channelID := m.ChannelID

	if channelID == "" {
		_, err := s.ChannelMessageSend(m.ChannelID, "❌ This command can only be used within a channel.")
		if err != nil {
			return fmt.Errorf("failed to send message: %w", err)
		}
		return nil
	}

	// Retrieve session participants using ChannelID
	participants, err := c.spotifyService.GetSessionParticipants(ctx, channelID)
	if err != nil {
		return fmt.Errorf("failed to get session participants: %w", err)
	}

	if len(participants) == 0 {
		_, err := s.ChannelMessageSend(m.ChannelID, "👥 There are currently no users in the jam session.")
		if err != nil {
			return fmt.Errorf("failed to send message: %w", err)
		}
		return nil
	}

	// Fetch user details from Discord
	var userMentions []string
	for _, userID := range participants {
		user, err := s.User(userID)
		if err != nil {
			userMentions = append(userMentions, fmt.Sprintf("Unknown User (%s)", userID))
			continue
		}
		userMentions = append(userMentions, user.Mention())
	}

	usersMessage := "👥 **Current Jam Session Participants:**\n" + "<" + stringJoin(userMentions, ">, <") + ">"

	_, err = s.ChannelMessageSend(m.ChannelID, usersMessage)
	if err != nil {
		return fmt.Errorf("failed to send users message: %w", err)
	}

	return nil
}

// Helper function to join strings with a separator
func stringJoin(elements []string, sep string) string {
	result := ""
	for i, elem := range elements {
		if i > 0 {
			result += sep
		}
		result += elem
	}
	return result
}
