package spotify

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"jam-bot/internal/config"

	"github.com/go-redis/redis/v8"
	"golang.org/x/oauth2"
)

type Service struct {
	config      *oauth2.Config
	redisClient *redis.Client
	SendDM      func(discordUserID, message string) error // Add SendDM function
}

// NewSpotifyService initializes the Spotify service with Redis and SendDM function
func NewSpotifyService(cfg *config.Config, sendDM func(string, string) error) *Service {
	oauthConfig := &oauth2.Config{
		ClientID:     cfg.SpotifyClientID,
		ClientSecret: cfg.SpotifyClientSecret,
		RedirectURL:  cfg.SpotifyRedirectURI,
		Scopes: []string{
			"user-read-playback-state",
			"user-modify-playback-state",
			"user-read-currently-playing",
			"streaming",
			"app-remote-control",
		},
		Endpoint: oauth2.Endpoint{
			AuthURL:  "https://accounts.spotify.com/authorize",
			TokenURL: "https://accounts.spotify.com/api/token",
		},
	}

	redisClient := redis.NewClient(&redis.Options{
		Addr:     cfg.RedisAddr,
		Password: cfg.RedisPassword,
		DB:       cfg.RedisDB,
	})

	// Test Redis connection
	ctx := context.Background()
	if err := redisClient.Ping(ctx).Err(); err != nil {
		log.Fatalf("[ERROR] Unable to connect to Redis: %v", err)
	}

	log.Println("[INFO] Connected to Redis successfully")

	return &Service{
		config:      oauthConfig,
		redisClient: redisClient,
		SendDM:      sendDM, // Assign SendDM function
	}
}

// GetAuthURL returns the Spotify OAuth2 URL for a given Discord user (state is DiscordUserID)
func (s *Service) GetAuthURL(state string) string {
	return s.config.AuthCodeURL(state, oauth2.AccessTypeOffline)
}

// HandleCallback processes the OAuth2 callback and stores the token in Redis
func (s *Service) HandleCallback(ctx context.Context, state, code string) error {
	token, err := s.config.Exchange(ctx, code)
	if err != nil {
		return fmt.Errorf("failed to exchange code for token: %w", err)
	}

	// Serialize token to JSON
	tokenData, err := json.Marshal(token)
	if err != nil {
		return fmt.Errorf("failed to marshal token: %w", err)
	}

	// Save token in Redis with Discord user ID as key
	err = s.redisClient.Set(ctx, state, tokenData, time.Hour*24*30).Err()
	if err != nil {
		return fmt.Errorf("failed to save token to Redis: %w", err)
	}

	log.Printf("[INFO] Successfully authenticated user: %s", state)

	// Send success DM to the user
	if s.SendDM != nil {
		err = s.SendDM(state, "✅ **Spotify Authentication Successful!** Your Spotify account has been connected.")
		if err != nil {
			log.Printf("[ERROR] Failed to send DM to user %s: %v", state, err)
		}
	}

	return nil
}

// GetClient returns an authenticated http.Client for a given Discord user
func (s *Service) GetClient(ctx context.Context, discordUserID string) (*http.Client, error) {
	// Retrieve token from Redis
	tokenData, err := s.redisClient.Get(ctx, discordUserID).Result()
	if err == redis.Nil {
		return nil, fmt.Errorf("user not authenticated with Spotify")
	} else if err != nil {
		return nil, fmt.Errorf("failed to get token from Redis: %w", err)
	}

	// Deserialize token
	var token oauth2.Token
	err = json.Unmarshal([]byte(tokenData), &token)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal token: %w", err)
	}

	// Check if token is expired and refresh if necessary
	if token.Expiry.Before(time.Now()) && token.RefreshToken != "" {
		ts := s.config.TokenSource(ctx, &token)
		newToken, err := ts.Token()
		if err != nil {
			return nil, fmt.Errorf("failed to refresh token: %w", err)
		}

		// Serialize new token
		newTokenData, err := json.Marshal(newToken)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal new token: %w", err)
		}

		// Save new token to Redis
		err = s.redisClient.Set(ctx, discordUserID, newTokenData, time.Hour*24*30).Err()
		if err != nil {
			return nil, fmt.Errorf("failed to save refreshed token to Redis: %w", err)
		}

		token = *newToken
	}

	return s.config.Client(ctx, &token), nil
}

// IsAuthenticated checks if a Discord user is already authenticated with Spotify
func (s *Service) IsAuthenticated(ctx context.Context, discordUserID string) (bool, error) {
	tokenData, err := s.redisClient.Get(ctx, discordUserID).Result()
	if err == redis.Nil {
		return false, nil // No token found
	} else if err != nil {
		return false, fmt.Errorf("failed to get token from Redis: %w", err)
	}

	var token oauth2.Token
	err = json.Unmarshal([]byte(tokenData), &token)
	if err != nil {
		return false, fmt.Errorf("failed to unmarshal token: %w", err)
	}

	// Optionally, check if the token is expired and refresh it
	if token.Expiry.Before(time.Now()) && token.RefreshToken == "" {
		return false, nil // Token expired and no refresh token available
	}

	return true, nil
}
