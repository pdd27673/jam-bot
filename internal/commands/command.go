package commands

import (
	"github.com/bwmarrin/discordgo"
)

// Command defines the interface for bot commands
type Command interface {
	// Name returns the command name
	Name() string
	// Description returns the command description
	Description() string
	// Execute runs the command with the given session and message
	Execute(s *discordgo.Session, m *discordgo.MessageCreate, args []string) error
}
