# 🤖 Lychee Discord Bot

The official Discord bot for the [Lychee](https://github.com/MD-Mushfiqur123/lychee) community server.

## Features

- **👋 Auto-Welcome** — Greets new members via DM with setup instructions
- **🎖️ Early Adopter Role** — Auto-assigns the @Early Adopter role to new members
- **💬 FAQ Auto-Reply** — Detects common questions and answers automatically
- **🍈 `/lychee-status`** — Live server status, version, and loaded models
- **📦 `/lychee-models`** — List installed models with optional search
- **❓ `/help`** — Contextual help with interactive buttons
- **📊 `/poll`** — Create community polls with emoji reactions
- **🔗 `/github`** — Quick links to GitHub issues and PRs
- **🛡️ Anti-Spam** — Auto-deletes invite links and API key leaks
- **📋 Mod Logging** — Tracks message deletes and edits

## Quick Start

### 1. Prerequisites

- **Node.js** ≥ 18.x
- **npm** ≥ 9.x
- A Discord Bot Token ([create one here](https://discord.com/developers/applications))

### 2. Setup

```bash
cd community/discord-bot
cp .env.example .env
```

Edit `.env` with your bot credentials:

```env
DISCORD_TOKEN=your_bot_token_here
DISCORD_CLIENT_ID=your_application_id_here
DISCORD_GUILD_ID=your_server_id_here       # optional
LYCHEE_API_URL=http://localhost:11434       # or your Lychee server URL
```

### 3. Install & Run

```bash
npm install
npm run dev          # development (auto-reload)
# or
npm start            # production
```

### 4. Invite the Bot

1. Go to the [Discord Developer Portal](https://discord.com/developers/applications)
2. Select your application → **OAuth2 → URL Generator**
3. Scopes: `bot`, `applications.commands`
4. Permissions: `Send Messages`, `Embed Links`, `Read Message History`, `Add Reactions`, `Use Slash Commands`, `Manage Messages`
5. Open the generated URL and invite the bot to your server

## Commands

| Command | Description |
|---|---|
| `/lychee-status` | Show Lychee server status, uptime, and loaded models |
| `/lychee-models [query]` | List locally installed Lychee models (optional search filter) |
| `/help [topic]` | Get help — topics: install, models, api, commands, contributing, troubleshooting |
| `/poll <question> <opt1> <opt2> [opt3...5]` | Create a community poll with emoji reactions |
| `/github <type> <number>` | Link to a GitHub issue or pull request |

## Development

### Project Structure

```
discord-bot/
├── index.js           # Main bot code (auto-welcome, FAQ, slash commands, anti-spam)
├── package.json       # Node dependencies and scripts
├── .env.example       # Environment variable template
└── README.md          # This file
```

### Adding New Commands

1. Add your command definition to the `slashCommands` array in `index.js`
2. Follow the `SlashCommandBuilder` pattern from discord.js
3. Restart the bot — commands are registered automatically on startup

### FAQ Patterns

To add or modify auto-reply patterns, edit the `FAQ_PATTERNS` array in `index.js`. Each entry has:
- `pattern` — A RegExp that matches the message content
- `reply` — A function that returns the reply string

### Logging

The bot uses Winston for structured logging. Set `LOG_LEVEL=debug` in `.env` for verbose output.

## Deployment

### PM2 (Recommended)

```bash
npm install -g pm2
pm2 start index.js --name lychee-discord-bot
pm2 save
pm2 startup
```

### Docker

```bash
docker build -t lychee-discord-bot .
docker run -d --name lychee-bot --env-file .env --restart unless-stopped lychee-discord-bot
```

### Systemd (Linux)

```bash
sudo cp lychee-discord-bot.service /etc/systemd/system/
sudo systemctl enable lychee-discord-bot
sudo systemctl start lychee-discord-bot
```

## Environment Variables

| Variable | Required | Default | Description |
|---|---|---|---|
| `DISCORD_TOKEN` | ✅ | — | Bot token from Discord Developer Portal |
| `DISCORD_CLIENT_ID` | ✅ | — | Application ID (needed for slash commands) |
| `DISCORD_GUILD_ID` | ❌ | — | Server ID for instant guild commands (omit for global) |
| `LYCHEE_API_URL` | ❌ | `http://localhost:11434` | Lychee API endpoint |
| `MOD_LOG_CHANNEL_ID` | ❌ | — | Channel ID for moderation logs |
| `LOG_LEVEL` | ❌ | `info` | Winston log level |

## License

MIT — see the [main repository license](../../LICENSE).
