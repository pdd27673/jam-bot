package commands

import (
	"fmt"
	"strings"

	"github.com/bwmarrin/discordgo"
)

// HelpCommand lists all available commands
type HelpCommand struct {
	registry *Registry
}

// NewHelpCommand creates a new HelpCommand
func NewHelpCommand(r *Registry) *HelpCommand {
	return &HelpCommand{registry: r}
}

// Name returns the command name
func (c *HelpCommand) Name() string {
	return "help"
}

// Description returns the command description
func (c *HelpCommand) Description() string {
	return "Lists all available commands"
}

// Execute runs the command with the given session and message
func (c *HelpCommand) Execute(s *discordgo.Session, m *discordgo.MessageCreate, args []string) error {
	var builder strings.Builder
	builder.WriteString("Available commands:\n")
	for _, cmd := range c.registry.commands {
		builder.WriteString(fmt.Sprintf("`!%s` - %s\n", cmd.Name(), cmd.Description()))
	}

	_, err := s.ChannelMessageSend(m.ChannelID, builder.String())
	if err != nil {
		return fmt.Errorf("[ERROR] failed to send message: %w", err)
	}
	return nil
}
