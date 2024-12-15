package commands

import (
	"errors"
	"strings"

	"github.com/bwmarrin/discordgo"
)

// Registry holds all registered commands
type Registry struct {
	commands map[string]Command
}

// NewRegistry creates a new command registry
func NewRegistry() *Registry {
	return &Registry{
		commands: make(map[string]Command),
	}
}

// Register adds a command to the registry
func (r *Registry) Register(c Command) {
	r.commands[strings.ToLower(c.Name())] = c
}

// Get returns a command from the registry
func (r *Registry) Get(name string) (Command, error) {
	cmd, exists := r.commands[strings.ToLower(name)]
	if !exists {
		return nil, errors.New("command not found")
	}
	return cmd, nil
}

// ExecuteCommand finds and executes a command based on the input
func (r *Registry) ExecuteCommand(s *discordgo.Session, m *discordgo.MessageCreate, prefix string) error {
	if !strings.HasPrefix(m.Content, prefix) {
		return nil
	}

	content := strings.TrimPrefix(m.Content, prefix)
	args := strings.Fields(content)
	if len(args) == 0 {
		return errors.New("empty command")
	}

	cmdName := strings.ToLower(args[0])
	cmdArgs := args[1:]

	cmd, err := r.Get(cmdName)
	if err != nil {
		return err
	}

	return cmd.Execute(s, m, cmdArgs)
}

// Add helper method to get all commands
func (r *Registry) GetCommands() map[string]*Command {
	result := make(map[string]*Command)
	for k, v := range r.commands {
		result[k] = &v
	}
	return result
}
