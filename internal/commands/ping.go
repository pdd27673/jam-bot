package commands

import (
	"errors"

	"github.com/bwmarrin/discordgo"
)

// PingCommand responds with "Pong!"
type PingCommand struct{}

// Name returns the command name
func (c *PingCommand) Name() string {
	return "ping"
}

// Description returns the command description
func (c *PingCommand) Description() string {
	return "Responds with 'Pong!'"
}

// Execute runs the command with the given session and message
func (c *PingCommand) Execute(s *discordgo.Session, m *discordgo.MessageCreate, args []string) error {
	_, err := s.ChannelMessageSend(m.ChannelID, "Pong!")
	if err != nil {
		return errors.New("[ERROR] failed to send message:" + err.Error())
	}
	return nil
}
