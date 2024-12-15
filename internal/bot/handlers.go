package bot

import (
	"fmt"
	"strings"

	"jam-bot/internal/utils" // Add this import

	"github.com/bwmarrin/discordgo"
)

// messageCreate is called whenever a new message is created in a channel
func messageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {
	// ignore empty messages
	if strings.TrimSpace(m.Content) == "" {
		return
	}

	// ignore messages created by the bot
	if m.Author.ID == s.State.User.ID {
		return
	}

	// check if the message starts withq the bot prefix
	if !strings.HasPrefix(m.Content, utils.DISCORD_BOT_PREFIX) {
		return
	}

	// split the message into command and arguments
	args := strings.Fields(m.Content)
	if len(args) == 0 {
		return
	}

	command := strings.ToLower(args[0][len(utils.DISCORD_BOT_PREFIX):]) // remove the prefix from the command and convert to lowercase

	switch command {
	case string(utils.COMMAND_PING):
		handlePingCommand(s, m)
	case string(utils.COMMAND_HELP):
		handleHelpCommand(s, m)
	default:
		handleUnknownCommand(s, m)
	}
}

// handlePingCommand responds to the ping command
func handlePingCommand(s *discordgo.Session, m *discordgo.MessageCreate) {
	_, err := s.ChannelMessageSend(m.ChannelID, "Pong!")
	if err != nil {
		fmt.Println("[ERROR] error sending message: ", err)
	}
}

// handleHelpCommand responds to the help command
func handleHelpCommand(s *discordgo.Session, m *discordgo.MessageCreate) {
	_, err := s.ChannelMessageSend(m.ChannelID, "Available commands:\n`!ping` - Responds with 'Pong!'\n`!help` - Displays this help message")
	if err != nil {
		fmt.Println("[ERROR] error sending message: ", err)
	}
}

// handleUnknownCommand responds to unknown commands
func handleUnknownCommand(s *discordgo.Session, m *discordgo.MessageCreate) {
	_, err := s.ChannelMessageSend(m.ChannelID, "Unknown command. Type `!help` for a list of available commands.")
	if err != nil {
		fmt.Println("[ERROR] error sending message: ", err)
	}
}
