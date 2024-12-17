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
	RedisAddr           string
	RedisPassword       string
	RedisDB             int
	Port                int // Added Port field
}

func LoadConfig() (*Config, error) {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")
	viper.AddConfigPath("./internal/config") // Ensure the correct path

	viper.AutomaticEnv()
	viper.BindEnv("DISCORD_TOKEN")
	viper.BindEnv("SPOTIFY_CLIENT_ID")
	viper.BindEnv("SPOTIFY_CLIENT_SECRET")
	viper.BindEnv("SPOTIFY_REDIRECT_URI")
	viper.BindEnv("BOT_PREFIX")
	viper.BindEnv("REDISADDR") // Environment variables are case-insensitive
	viper.BindEnv("REDISPASSWORD")
	viper.BindEnv("REDISDB")
	viper.BindEnv("PORT")        // Bind PORT environment variable
	viper.BindEnv("SERVER_PORT") // Bind PORT environment variable

	viper.SetDefault("BotPrefix", "!")
	viper.SetDefault("RedisAddr", "localhost:6379")
	viper.SetDefault("RedisPassword", "")
	viper.SetDefault("RedisDB", 0)
	viper.SetDefault("Port", 8080) // Default port

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
		RedisAddr:           viper.GetString("RedisAddr"),
		RedisPassword:       viper.GetString("RedisPassword"),
		RedisDB:             viper.GetInt("RedisDB"),
		Port:                viper.GetInt("Port"), // Load Port
	}

	return config, nil
}
