// internal/spotify/service.go
package spotify

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"sync"

	"golang.org/x/oauth2"
)

type Service struct {
	config *oauth2.Config
	tokens map[string]*oauth2.Token // Map Discord user ID to Spotify token
	mu     sync.RWMutex
}

func NewSpotifyService(clientID, clientSecret, redirectURI string) *Service {
	config := &oauth2.Config{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		RedirectURL:  redirectURI,
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

	return &Service{
		config: config,
		tokens: make(map[string]*oauth2.Token),
	}
}

// GetAuthURL returns the Spotify OAuth2 URL for a given Discord user
func (s *Service) GetAuthURL(state string) string {
	return s.config.AuthCodeURL(state)
}

// HandleCallback processes the OAuth2 callback and stores the token
func (s *Service) HandleCallback(ctx context.Context, state, code string) error {
	token, err := s.config.Exchange(ctx, code)
	if err != nil {
		return fmt.Errorf("failed to exchange code for token: %w", err)
	}

	s.mu.Lock()
	s.tokens[state] = token
	s.mu.Unlock()

	log.Printf("[INFO] Successfully authenticated user: %s", state)

	return nil
}

// GetClient returns an authenticated http.Client for a given Discord user
func (s *Service) GetClient(ctx context.Context, discordUserID string) (*http.Client, error) {
	s.mu.RLock()
	token, exists := s.tokens[discordUserID]
	s.mu.RUnlock()

	if !exists {
		return nil, fmt.Errorf("user not authenticated with Spotify")
	}

	return s.config.Client(ctx, token), nil
}
