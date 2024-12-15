// config/config.go
package config

import (
	"fmt"

	"github.com/spf13/viper"
)

type Config struct {
	DiscordToken        string
	SpotifyClientID     string
	SpotifyClientSecret string
	SpotifyRedirectURI  string
	BotPrefix           string
}

func LoadConfig() (*Config, error) {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")

	viper.AutomaticEnv()
	viper.BindEnv("DISCORD_TOKEN")
	viper.BindEnv("SPOTIFY_CLIENT_ID")
	viper.BindEnv("SPOTIFY_CLIENT_SECRET")
	viper.BindEnv("SPOTIFY_REDIRECT_URI")

	viper.SetDefault("BotPrefix", "!")

	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			// Config file was found but another error was produced
			return nil, fmt.Errorf("failed to read configuration file: %w", err)
		}
		// Config file not found; proceed with environment variables
		fmt.Println("[INFO] No configuration file found; using environment variables.")
	}

	config := &Config{
		DiscordToken:        viper.GetString("DISCORD_TOKEN"),
		SpotifyClientID:     viper.GetString("SPOTIFY_CLIENT_ID"),
		SpotifyClientSecret: viper.GetString("SPOTIFY_CLIENT_SECRET"),
		SpotifyRedirectURI:  viper.GetString("SPOTIFY_REDIRECT_URI"),
		BotPrefix:           viper.GetString("BotPrefix"),
	}

	return config, nil
}
