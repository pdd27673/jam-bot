package spotify

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"strings"
	"time"

	"jam-bot/internal/config"

	"github.com/go-redis/redis/v8"
	"golang.org/x/oauth2"
)

type Song struct {
	Title  string `json:"title"`
	Artist string `json:"artist"`
	URI    string `json:"uri"`
}

type PlaybackState struct {
	CurrentSong   Song  `json:"current_song"`
	PositionMs    int   `json:"position_ms"`
	IsPlaying     bool  `json:"is_playing"`
	LastUpdatedAt int64 `json:"last_updated_at"` // Unix timestamp in seconds
}

type Session struct {
	ChannelID    string        `json:"channel_id"` // Changed from GuildID to ChannelID
	Participants []string      `json:"participants"`
	Queue        []Song        `json:"queue"`
	Playback     PlaybackState `json:"playback"`
}

// Session Key Prefix
const sessionKeyPrefix = "jam_session_channel:" // Updated prefix

// Service represents the Spotify service
type Service struct {
	config      *oauth2.Config
	redisClient *redis.Client
	SendDM      func(discordUserID, message string) error // Add SendDM function
}

// NewSpotifyService initializes the Spotify service with Redis and SendDM function
func NewSpotifyService(cfg *config.Config, sendDM func(discordUserID, message string) error) *Service {
	rdb := redis.NewClient(&redis.Options{
		Addr:     cfg.RedisAddr,
		Password: cfg.RedisPassword,
		DB:       cfg.RedisDB,
	})

	// Initialize OAuth2 config
	oauthCfg := &oauth2.Config{
		ClientID:     cfg.SpotifyClientID,
		ClientSecret: cfg.SpotifyClientSecret,
		RedirectURL:  cfg.SpotifyRedirectURI,
		Scopes:       []string{"user-read-playback-state", "user-modify-playback-state", "user-read-currently-playing"},
		Endpoint: oauth2.Endpoint{
			AuthURL:  "https://accounts.spotify.com/authorize",
			TokenURL: "https://accounts.spotify.com/api/token",
		},
	}

	return &Service{
		config:      oauthCfg,
		redisClient: rdb,
		SendDM:      sendDM,
	}
}

// StartAuthServer starts the authentication server on the configured port
func (s *Service) StartAuthServer() error {
	cfg, err := config.LoadConfig()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	return s.StartAuthServerWithPort(cfg.Port)
}

// StartAuthServerWithPort starts the auth server on a specified port
func (s *Service) StartAuthServerWithPort(port int) error {
	http.HandleFunc("/callback", s.callbackHandler)
	server := &http.Server{
		Addr: fmt.Sprintf(":%d", port),
	}

	log.Printf("[INFO] Starting auth server on port %d\n", port)
	return server.ListenAndServe()
}

// callbackHandler remains unchanged
func (s *Service) callbackHandler(w http.ResponseWriter, r *http.Request) {
	state := r.URL.Query().Get("state") // Discord user ID
	code := r.URL.Query().Get("code")
	// channelID := r.URL.Query().Get("channel_id") // Optional: can pass ChannelID if needed

	if state == "" || code == "" {
		http.Error(w, "Invalid callback parameters", http.StatusBadRequest)
		return
	}

	if err := s.HandleCallback(r.Context(), state, code); err != nil {
		http.Error(w, "Authentication failed", http.StatusInternalServerError)
		return
	}

	// Notify user of successful authentication
	fmt.Fprintf(w, "<script>window.close()</script>")
}

// GetAuthURL returns the Spotify OAuth2 URL for a given Discord user (state is DiscordUserID)
func (s *Service) GetAuthURL(discordUserID string) string {
	state := discordUserID
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

// CreateSession creates a new jam session for a channel
func (s *Service) CreateSession(ctx context.Context, channelID string) error {
	session := Session{
		ChannelID:    channelID,
		Participants: []string{},
		Queue:        []Song{},
		Playback: PlaybackState{
			CurrentSong:   Song{},
			PositionMs:    0,
			IsPlaying:     false,
			LastUpdatedAt: time.Now().Unix(),
		},
	}

	sessionData, err := json.Marshal(session)
	if err != nil {
		return fmt.Errorf("failed to marshal session: %w", err)
	}

	key := fmt.Sprintf("%s%s", sessionKeyPrefix, channelID)
	return s.redisClient.Set(ctx, key, sessionData, 0).Err()
}

// LoadSession loads a jam session from Redis based on ChannelID
func (s *Service) LoadSession(ctx context.Context, channelID string) (*Session, error) {
	key := fmt.Sprintf("%s%s", sessionKeyPrefix, channelID)
	sessionData, err := s.redisClient.Get(ctx, key).Result()
	if err == redis.Nil {
		return nil, fmt.Errorf("no active session for channel %s", channelID)
	} else if err != nil {
		return nil, fmt.Errorf("failed to get session from Redis: %w", err)
	}

	var session Session
	err = json.Unmarshal([]byte(sessionData), &session)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal session: %w", err)
	}

	return &session, nil
}

// SaveSession saves a jam session to Redis based on ChannelID
func (s *Service) SaveSession(ctx context.Context, session *Session) error {
	sessionData, err := json.Marshal(session)
	if err != nil {
		return fmt.Errorf("failed to marshal session: %w", err)
	}

	key := fmt.Sprintf("%s%s", sessionKeyPrefix, session.ChannelID)
	return s.redisClient.Set(ctx, key, sessionData, 0).Err()
}

// DeleteSession deletes a jam session from Redis based on ChannelID
func (s *Service) DeleteSession(ctx context.Context, channelID string) error {
	key := fmt.Sprintf("%s%s", sessionKeyPrefix, channelID)
	return s.redisClient.Del(ctx, key).Err()
}

// AddUserToSession adds a user to the jam session for a specific channel
func (s *Service) AddUserToSession(ctx context.Context, channelID, userID string) error {
	session, err := s.LoadSession(ctx, channelID)
	if err != nil {
		// If no session exists, create one
		if err.Error() == fmt.Sprintf("no active session for channel %s", channelID) {
			err = s.CreateSession(ctx, channelID)
			if err != nil {
				return fmt.Errorf("failed to create session: %w", err)
			}
			session, err = s.LoadSession(ctx, channelID)
			if err != nil {
				return fmt.Errorf("failed to load session after creation: %w", err)
			}
		} else {
			return err
		}
	}

	// Check if user is already in session
	for _, id := range session.Participants {
		if id == userID {
			return fmt.Errorf("user is already in the session")
		}
	}

	session.Participants = append(session.Participants, userID)
	return s.SaveSession(ctx, session)
}

// RemoveUserFromSession removes a user from the jam session for a specific channel
func (s *Service) RemoveUserFromSession(ctx context.Context, channelID, userID string) error {
	session, err := s.LoadSession(ctx, channelID)
	if err != nil {
		return err
	}

	// Find and remove the user from participants
	for i, id := range session.Participants {
		if id == userID {
			session.Participants = append(session.Participants[:i], session.Participants[i+1:]...)
			break
		}
	}

	return s.SaveSession(ctx, session)
}

// GetSessionParticipants retrieves all users in the jam session for a specific channel
func (s *Service) GetSessionParticipants(ctx context.Context, channelID string) ([]string, error) {
	session, err := s.LoadSession(ctx, channelID)
	if err != nil {
		return nil, err
	}

	return session.Participants, nil
}

// LoadAllSessions retrieves all active jam sessions from Redis
func (s *Service) LoadAllSessions(ctx context.Context) ([]Session, error) {
	var sessions []Session
	iter := s.redisClient.Scan(ctx, 0, fmt.Sprintf("%s*", sessionKeyPrefix), 0).Iterator()
	for iter.Next(ctx) {
		key := iter.Val()
		sessionData, err := s.redisClient.Get(ctx, key).Result()
		if err != nil {
			log.Printf("[ERROR] Failed to get session data for key %s: %v", key, err)
			continue
		}

		var session Session
		err = json.Unmarshal([]byte(sessionData), &session)
		if err != nil {
			log.Printf("[ERROR] Failed to unmarshal session data for key %s: %v", key, err)
			continue
		}

		sessions = append(sessions, session)
	}

	if err := iter.Err(); err != nil {
		return nil, fmt.Errorf("error iterating Redis keys: %w", err)
	}

	return sessions, nil
}

// AddSongToQueue adds a song to the session's queue
func (s *Service) AddSongToQueue(ctx context.Context, channelID string, song Song) error {
	session, err := s.LoadSession(ctx, channelID)
	if err != nil {
		return fmt.Errorf("failed to load session: %w", err)
	}

	session.Queue = append(session.Queue, song)
	return s.SaveSession(ctx, session)
}

// SearchSong searches for a song using Spotify API and returns the first result
func (s *Service) SearchSong(ctx context.Context, userId, query string) (Song, error) {
	client, err := s.GetClient(ctx, userId) // Using ChannelID to get the client
	if err != nil {
		return Song{}, fmt.Errorf("failed to get Spotify client: %w", err)
	}

	searchURL := "https://api.spotify.com/v1/search"
	params := url.Values{}
	params.Add("q", query)
	params.Add("type", "track")
	params.Add("limit", "1")

	req, err := http.NewRequestWithContext(ctx, "GET", fmt.Sprintf("%s?%s", searchURL, params.Encode()), nil)
	if err != nil {
		return Song{}, fmt.Errorf("failed to create search request: %w", err)
	}

	resp, err := client.Do(req)
	if err != nil {
		return Song{}, fmt.Errorf("search request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := ioutil.ReadAll(resp.Body)
		return Song{}, fmt.Errorf("search request returned status %d: %s", resp.StatusCode, string(bodyBytes))
	}

	var searchResult struct {
		Tracks struct {
			Items []struct {
				Name    string `json:"name"`
				Artists []struct {
					Name string `json:"name"`
				} `json:"artists"`
				URI string `json:"uri"`
			} `json:"items"`
		} `json:"tracks"`
	}

	err = json.NewDecoder(resp.Body).Decode(&searchResult)
	if err != nil {
		return Song{}, fmt.Errorf("failed to decode search response: %w", err)
	}

	if len(searchResult.Tracks.Items) == 0 {
		return Song{}, fmt.Errorf("no results found for query: %s", query)
	}

	firstTrack := searchResult.Tracks.Items[0]
	artistNames := []string{}
	for _, artist := range firstTrack.Artists {
		artistNames = append(artistNames, artist.Name)
	}

	return Song{
		Title:  firstTrack.Name,
		Artist: strings.Join(artistNames, ", "),
		URI:    firstTrack.URI,
	}, nil
}

// RemoveSongFromQueue removes a song from the session's queue at the specified index
func (s *Service) RemoveSongFromQueue(ctx context.Context, channelID string, index int) error {
	session, err := s.LoadSession(ctx, channelID)
	if err != nil {
		return fmt.Errorf("failed to load session: %w", err)
	}

	if index < 0 || index >= len(session.Queue) {
		return fmt.Errorf("song position out of range")
	}

	removedSong := session.Queue[index]
	session.Queue = append(session.Queue[:index], session.Queue[index+1:]...)

	// If the removed song is currently playing, reset playback state
	if session.Playback.IsPlaying && session.Playback.CurrentSong.URI == removedSong.URI {
		session.Playback.IsPlaying = false
		session.Playback.PositionMs = 0       // Reset position
		session.Playback.CurrentSong = Song{} // Clear current song
	}

	err = s.SaveSession(ctx, session)
	if err != nil {
		return fmt.Errorf("failed to save session: %w", err)
	}

	return nil
}

// ///////////////////////
// //////////////////////
// PlayRequest represents the payload for the Spotify Play API
type PlayRequest struct {
	URIs       []string `json:"uris,omitempty"`
	PositionMs int      `json:"position_ms,omitempty"`
}

// Device represents a Spotify device
type Device struct {
	ID            string `json:"id"`
	IsActive      bool   `json:"is_active"`
	IsRestricted  bool   `json:"is_restricted"`
	Name          string `json:"name"`
	Type          string `json:"type"`
	VolumePercent int    `json:"volume_percent"`
}

// DevicesResponse represents the response from the Spotify Devices API
type DevicesResponse struct {
	Devices []Device `json:"devices"`
}

// StartPlayback starts playback of the queued songs for all participants in the channel's session
func (s *Service) StartPlayback(ctx context.Context, channelID string) error {
	session, err := s.LoadSession(ctx, channelID)
	if err != nil {
		return fmt.Errorf("failed to load session: %w", err)
	}

	if len(session.Queue) == 0 {
		return fmt.Errorf("the queue is empty")
	}

	currentSong := session.Queue[0]

	// Reset position if starting a new song
	if session.Playback.CurrentSong.URI != currentSong.URI {
		session.Playback.PositionMs = 0
		session.Playback.CurrentSong = currentSong
	} else if session.Playback.IsPlaying {
		// Only update position if it's the same song and was playing
		elapsed := time.Now().Unix() - session.Playback.LastUpdatedAt
		session.Playback.PositionMs += int(elapsed * 1000)
	}

	// Iterate over each participant and start playback
	for _, userID := range session.Participants {
		client, err := s.GetClient(ctx, userID)
		if err != nil {
			s.SendDM(userID, fmt.Sprintf("❌ Failed to start playback: %v", err))
			continue
		}

		// Retrieve user's active devices
		devices, err := s.GetUserDevices(ctx, client)
		if err != nil {
			s.SendDM(userID, fmt.Sprintf("❌ Unable to retrieve devices: %v", err))
			continue
		}

		// Select the first available active device
		if len(devices) == 0 {
			s.SendDM(userID, "❌ No active Spotify devices found. Please open Spotify on one of your devices.")
			continue
		}

		deviceID := devices[0].ID // Selecting the first device

		// Prepare the playback request with position_ms
		playReq := &PlayRequest{
			URIs:       []string{currentSong.URI},
			PositionMs: session.Playback.PositionMs,
		}

		playReqBody, err := json.Marshal(playReq)
		if err != nil {
			s.SendDM(userID, fmt.Sprintf("❌ Failed to marshal playback request: %v", err))
			continue
		}

		// Create the playback API request
		playURL := "https://api.spotify.com/v1/me/player/play?device_id=" + deviceID
		req, err := http.NewRequestWithContext(ctx, "PUT", playURL, strings.NewReader(string(playReqBody)))
		if err != nil {
			s.SendDM(userID, fmt.Sprintf("❌ Failed to create playback request: %v", err))
			continue
		}
		req.Header.Set("Content-Type", "application/json")

		// Execute the playback request
		resp, err := client.Do(req)
		if err != nil {
			s.SendDM(userID, fmt.Sprintf("❌ Playback request failed: %v", err))
			continue
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusNoContent {
			bodyBytes, _ := ioutil.ReadAll(resp.Body)
			s.SendDM(userID, fmt.Sprintf("❌ Failed to start playback (Status %d): %s", resp.StatusCode, string(bodyBytes)))
			continue
		}

		// Optionally, notify the user of successful playback start
		s.SendDM(userID, fmt.Sprintf("✅ Now playing **%s** by **%s** from %d ms.", currentSong.Title, currentSong.Artist, session.Playback.PositionMs))
	}

	// Update the session's playback state
	session.Playback.IsPlaying = true
	session.Playback.LastUpdatedAt = time.Now().Unix()
	session.Playback.CurrentSong = currentSong

	err = s.SaveSession(ctx, session)
	if err != nil {
		return fmt.Errorf("failed to save session: %w", err)
	}

	// quit := make(chan struct{})
	// s.StartSyncTicker(channelID, 10*time.Second, quit)

	return nil
}

// PausePlayback pauses the current playback for all participants in the channel's session
func (s *Service) PausePlayback(ctx context.Context, channelID string) error {
	session, err := s.LoadSession(ctx, channelID)
	if err != nil {
		return fmt.Errorf("failed to load session: %w", err)
	}

	if !session.Playback.IsPlaying {
		return fmt.Errorf("playback is already paused")
	}

	// Calculate the current playback position
	elapsed := time.Now().Unix() - session.Playback.LastUpdatedAt
	session.Playback.PositionMs += int(elapsed * 1000) // Convert seconds to milliseconds

	// Iterate over each participant and pause playback
	for _, userID := range session.Participants {
		client, err := s.GetClient(ctx, userID)
		if err != nil {
			s.SendDM(userID, fmt.Sprintf("❌ Failed to pause playback: %v", err))
			continue
		}

		// Retrieve user's active devices
		devices, err := s.GetUserDevices(ctx, client)
		if err != nil {
			s.SendDM(userID, fmt.Sprintf("❌ Unable to retrieve devices: %v", err))
			continue
		}

		// Select the first available active device
		if len(devices) == 0 {
			s.SendDM(userID, "❌ No active Spotify devices found. Please open Spotify on one of your devices.")
			continue
		}

		deviceID := devices[0].ID // Selecting the first device

		// Create the pause API request
		pauseURL := "https://api.spotify.com/v1/me/player/pause?device_id=" + deviceID
		req, err := http.NewRequestWithContext(ctx, "PUT", pauseURL, nil)
		if err != nil {
			s.SendDM(userID, fmt.Sprintf("❌ Failed to create pause request: %v", err))
			continue
		}

		// Execute the pause request
		resp, err := client.Do(req)
		if err != nil {
			s.SendDM(userID, fmt.Sprintf("❌ Pause request failed: %v", err))
			continue
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusNoContent {
			bodyBytes, _ := ioutil.ReadAll(resp.Body)
			s.SendDM(userID, fmt.Sprintf("❌ Failed to pause playback (Status %d): %s", resp.StatusCode, string(bodyBytes)))
			continue
		}

		// Optionally, notify the user of successful pause
		s.SendDM(userID, "✅ Playback has been paused.")
	}

	// Update the session's playback state
	session.Playback.IsPlaying = false
	session.Playback.LastUpdatedAt = time.Now().Unix()

	err = s.SaveSession(ctx, session)
	if err != nil {
		return fmt.Errorf("failed to save session: %w", err)
	}

	return nil
}

// GetUserDevices retrieves the available devices for a user
func (s *Service) GetUserDevices(ctx context.Context, client *http.Client) ([]Device, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", "https://api.spotify.com/v1/me/player/devices", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create devices request: %w", err)
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("devices request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := ioutil.ReadAll(resp.Body)
		return nil, fmt.Errorf("devices request returned status %d: %s", resp.StatusCode, string(bodyBytes))
	}

	var devicesResp DevicesResponse
	err = json.NewDecoder(resp.Body).Decode(&devicesResp)
	if err != nil {
		return nil, fmt.Errorf("failed to decode devices response: %w", err)
	}

	return devicesResp.Devices, nil
}

// StartSyncTicker initiates a ticker to periodically synchronize playback positions
func (s *Service) StartSyncTicker(channelID string, interval time.Duration, quit <-chan struct{}) {
	ticker := time.NewTicker(interval)
	go func() {
		for {
			select {
			case <-ticker.C:
				err := s.SynchronizePlayback(channelID)
				if err != nil {
					fmt.Printf("Failed to synchronize playback for channel %s: %v\n", channelID, err)
				}
			case <-quit:
				ticker.Stop()
				return
			}
		}
	}()
}

// SynchronizePlayback ensures all users are playing the current song from the same position
func (s *Service) SynchronizePlayback(channelID string) error {
	session, err := s.LoadSession(context.Background(), channelID)
	if err != nil {
		return fmt.Errorf("failed to load session: %w", err)
	}

	if !session.Playback.IsPlaying {
		return nil // No need to synchronize if not playing
	}

	// Calculate the current playback position
	elapsed := time.Now().Unix() - session.Playback.LastUpdatedAt
	currentPosition := session.Playback.PositionMs + int(elapsed*1000)

	// Iterate over each participant and update playback position if necessary
	for _, userID := range session.Participants {
		client, err := s.GetClient(context.Background(), userID)
		if err != nil {
			s.SendDM(userID, fmt.Sprintf("❌ Failed to synchronize playback: %v", err))
			continue
		}

		// Retrieve user's active devices
		devices, err := s.GetUserDevices(context.Background(), client)
		if err != nil {
			s.SendDM(userID, fmt.Sprintf("❌ Unable to retrieve devices for synchronization: %v", err))
			continue
		}

		if len(devices) == 0 {
			s.SendDM(userID, "❌ No active Spotify devices found for synchronization.")
			continue
		}

		deviceID := devices[0].ID // Selecting the first device

		// Prepare the playback request with synchronized position
		playReq := &PlayRequest{
			URIs:       []string{session.Playback.CurrentSong.URI},
			PositionMs: currentPosition,
		}

		playReqBody, err := json.Marshal(playReq)
		if err != nil {
			s.SendDM(userID, fmt.Sprintf("❌ Failed to marshal synchronization playback request: %v", err))
			continue
		}

		// Create the playback API request
		playURL := "https://api.spotify.com/v1/me/player/play?device_id=" + deviceID
		req, err := http.NewRequestWithContext(context.Background(), "PUT", playURL, strings.NewReader(string(playReqBody)))
		if err != nil {
			s.SendDM(userID, fmt.Sprintf("❌ Failed to create synchronization playback request: %v", err))
			continue
		}
		req.Header.Set("Content-Type", "application/json")

		// Execute the playback request
		resp, err := client.Do(req)
		if err != nil {
			s.SendDM(userID, fmt.Sprintf("❌ Synchronization playback request failed: %v", err))
			continue
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusNoContent {
			bodyBytes, _ := ioutil.ReadAll(resp.Body)
			s.SendDM(userID, fmt.Sprintf("❌ Failed to synchronize playback (Status %d): %s", resp.StatusCode, string(bodyBytes)))
			continue
		}

		// Optionally, notify the user of successful synchronization
		s.SendDM(userID, fmt.Sprintf("🔄 Synchronized playback to **%s** at %d ms.", session.Playback.CurrentSong.Title, currentPosition))
	}

	return nil
}
