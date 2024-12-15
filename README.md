# Spotify Jam Discord Bot

A Discord bot that enables collaborative music sessions using Spotify. Users can join sessions, manage a shared queue, and control playback synchronized across their Spotify accounts.

## Features

- **Session Management**: Create and manage music sessions that multiple users can join.
- **Shared Queue**: Add songs to a shared queue that plays for all session members.
- **Playback Control**: Play, pause, and skip songs in the session.
- **Voting System**: Vote to skip the current song or add songs to the queue.
- **Spotify Integration**: Authenticate with Spotify to control playback on your account.

## Commands

- `!join` - Join the current music session.
- `!leave` - Leave the current session.
- `!play` - Resume playback.
- `!pause` - Pause playback.
- `!add [song name]` - Add a song to the queue.
- `!skip` - Skip the current song.
- `!queue` - Display the current song queue.
- `!vote_skip` - Initiate a vote to skip the current song.

## Installation

### Prerequisites

- **Go 1.22 or higher** installed on your machine.
- **Discord Bot Token**:
  - Create an application on the [Discord Developer Portal](https://discord.com/developers/applications).
  - Add a bot to your application and copy the bot token.
- **Spotify Developer Credentials**:
  - Create an app on the [Spotify Developer Dashboard](https://developer.spotify.com/dashboard/applications).
  - Note down the Client ID and Client Secret.
  - Set the Redirect URI to `http://localhost:8080/callback` or your desired URI.

### Steps

1. **Clone the Repository**:

   ```bash
   git clone https://github.com/yourusername/jam-bot.git
   cd jam-bot
   ```

2. **Set Up Environment Variables**:

   Create a `.env` file in the root directory and add your credentials:

   ```env
   DISCORD_TOKEN=your_discord_bot_token
   SPOTIFY_CLIENT_ID=your_spotify_client_id
   SPOTIFY_CLIENT_SECRET=your_spotify_client_secret
   SPOTIFY_REDIRECT_URI=your_redirect_uri
   ```

3. **Install Dependencies**:

   ```bash
   go mod download
   ```

4. **Build the Bot**:

   ```bash
   go build -o discord-bot .
   ```

5. **Run the Bot**:

   ```bash
   ./discord-bot
   ```

6. **Invite the Bot to Your Server**:

   - Generate an OAuth2 URL with the necessary permissions from the Discord Developer Portal.
   - Use the URL to invite the bot to your Discord server.

## Usage

- **Start a Session**:
  - Use `!start_session` (if implemented) to initiate a music session.
  
- **Join a Session**:

  - Type `!join` in a channel where the bot is active.
  - Authenticate with Spotify when prompted.

- **Control Playback**:

  - Use `!play` and `!pause` to control playback.
  - Add songs with `!add [song name]`.

- **Manage the Queue**:

  - View the queue with `!queue`.
  - Vote to skip songs with `!vote_skip`.

## Development Plan Overview

The bot is designed with a modular structure for scalability and maintainability. Key components include:

- **Command Handling**: Modular commands with a registry pattern for easy addition of new commands.
- **Spotify Integration**: Handles authentication and communication with the Spotify API.
- **Session Management**: Manages user sessions and synchronization of playback.
- **Queue Management**: Handles song queues and playback order.
- **Voting System**: Implements collaborative decision-making features like skipping songs.

## Contributing

Contributions are welcome! Please follow these steps:

1. **Fork the Repository**:

   ```bash
   git fork https://github.com/yourusername/jam-bot.git
   ```

2. **Create a Feature Branch**:

   ```bash
   git checkout -b feature/your-feature-name
   ```

3. **Commit Your Changes**:

   ```bash
   git commit -am 'Add new feature'
   ```

4. **Push to the Branch**:

   ```bash
   git push origin feature/your-feature-name
   ```

5. **Open a Pull Request**.

## License

This project is open-source and available under the MIT License.

## Contact

For questions or support, please open an issue on the [GitHub repository](https://github.com/pdd27673/jam-bot/issues).