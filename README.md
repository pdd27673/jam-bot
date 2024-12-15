# Jam Bot

Jam Bot is a Discord bot designed to facilitate synchronized music sessions using Spotify. It enables users to create and manage music queues, control playback, and collaborate seamlessly within Discord servers.

## Features

- **Session Management**: Create and join music sessions to synchronize playback across multiple users.
- **Spotify Integration**: Authenticate with Spotify to access and control music playback.
- **Queue Management**: Add, view, and manage a queue of songs collaboratively.
- **Playback Controls**: Play, pause, and navigate through tracks with simple commands.
- **Voting System**: Implement collaborative decision-making features like voting to skip songs.

## Current Status

The Jam Bot is currently in an active development phase with the following features implemented:

- **Authentication**:
  - Users can authenticate with Spotify using the `!auth` command.
  - OAuth2 flow is integrated to securely connect user accounts with Spotify.
  
- **Session Commands**:
  - `!start_session`: Initiate a new music session.
  - `!join`: Join an existing music session.
  - `!leave`: Leave the current music session.

- **Playback Commands**:
  - `!play`: Resume playback.
  - `!pause`: Pause playback.
  - `!add [song name]`: Add a song to the queue.
  - `!queue`: View the current song queue.
  - `!vote_skip`: Vote to skip the current song.

- **Queue Management**:
  - Songs are added to a Redis-backed queue ensuring synchronization across users.
  
- **Docker Deployment**:
  - Dockerfile configured for deployment on Render.com with multi-stage builds.
  - Redis integration for session and queue management.

## What’s Left to Do

The following tasks are planned to enhance the Jam Bot's functionality and stability:

1. **Fix "Invalid Track URI" Error**:
   - Debug and resolve the issue causing invalid track URIs during synchronization.
   - Ensure all URIs retrieved from Spotify are correctly formatted and stored.

2. **Enhance Error Handling**:
   - Implement comprehensive error handling to provide meaningful feedback to users.
   - Prevent the bot from crashing due to unexpected errors.

3. **Improve Logging and Monitoring**:
   - Integrate advanced logging to monitor bot activities and quickly identify issues.
   - Set up alerting mechanisms for critical failures.

4. **Expand Command Set**:
   - Add more playback controls such as `!next`, `!previous`, and `!shuffle`.
   - Implement user-specific commands for personalized experiences.

5. **Optimize Redis Usage**:
   - Refine data structures and access patterns in Redis for better performance.
   - Implement session expiration and cleanup mechanisms.

6. **User Interface Enhancements**:
   - Improve Discord message embeds for a more user-friendly interface.
   - Provide visual feedback during authentication and other critical actions.

7. **Scalability Improvements**:
   - Optimize the bot for handling a larger number of simultaneous sessions.
   - Ensure stability under increased load and concurrent user interactions.

8. **Testing and QA**:
   - Develop unit and integration tests to ensure code quality.
   - Conduct thorough testing with multiple users to validate synchronization and playback features.

9. **Documentation**:
   - Expand the README with detailed setup instructions, FAQs, and troubleshooting tips.
   - Document codebase thoroughly for easier contributions and maintenance.

## Development Plan Overview

The bot is designed with a modular structure for scalability and maintainability. Key components include:

- **Command Handling**: Modular commands with a registry pattern for easy addition of new commands.
- **Spotify Integration**: Handles authentication and communication with the Spotify API.
- **Session Management**: Manages user sessions and synchronization of playback.
- **Queue Management**: Handles song queues and playback order.
- **Voting System**: Implements collaborative decision-making features like skipping songs.

## Getting Started

### Prerequisites

- Docker installed on your machine.
- A Spotify Developer account to obtain Client ID and Client Secret.
- A Discord application with a bot token.

### Installation

1. **Clone the Repository**:

    ```bash
    git clone https://github.com/pdd27673/jam-bot.git
    cd jam-bot
    ```

2. **Configure Environment Variables**:

    Create a `.env` file in the root directory and add the following:

    ```env
    DISCORD_TOKEN=your_discord_bot_token
    SPOTIFY_CLIENT_ID=your_spotify_client_id
    SPOTIFY_CLIENT_SECRET=your_spotify_client_secret
    SPOTIFY_REDIRECT_URI=https://your-service.onrender.com/callback
    REDISADDR=redis_host:redis_port
    REDISPASSWORD=your_redis_password
    REDISDB=0
    PORT=8080
    ```

3. **Build and Run with Docker**:

    ```bash
    docker build -t jambot:latest .
    docker run -d -p 8080:8080 --env-file .env jambot:latest
    ```

### Deployment on Render.com

1. **Create a New Web Service**:
   - Log in to [Render.com](https://render.com/) and create a new web service.
   - Connect your GitHub repository and select the branch to deploy.

2. **Configure Build and Start Commands**:
   - **Build Command**: `docker build -t jambot:latest .`
   - **Start Command**: `./discord-bot`

3. **Add Redis as an Add-On**:
   - In your Render service dashboard, navigate to the **"Add-Ons"** section.
   - Select **"Redis"** and add it to your project.
   - Note the `REDISURL` provided.

4. **Set Environment Variables**:
   - Navigate to the **"Environment"** tab in your Render service.
   - Add the required environment variables as shown in the Installation section.

5. **Update Spotify OAuth2 Settings**:
   - Go to the [Spotify Developer Dashboard](https://developer.spotify.com/dashboard/).
   - Update the **Redirect URI** to `https://<your-service-name>.onrender.com/callback`.

6. **Deploy the Service**:
   - Push your changes to GitHub.
   - Render will automatically build and deploy your application.
   - Monitor the deployment logs for any issues.

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