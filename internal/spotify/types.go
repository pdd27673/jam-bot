package spotify

type Song struct {
	Title  string `json:"title"`
	Artist string `json:"artist"`
	URI    string `json:"uri"`
}

type PlaybackState struct {
	CurrentSong   Song  `json:"current_song"`
	PositionMs    int   `json:"position_ms"`
	IsPlaying     bool  `json:"is_playing"`
	LastUpdatedAt int64 `json:"last_updated_at"`
}

type Session struct {
	GuildID      string        `json:"guild_id"`
	Participants []string      `json:"participants"`
	Queue        []Song        `json:"queue"`
	Playback     PlaybackState `json:"playback"`
}
