# 🍒 Getting Started with Lychee

Welcome! This guide will take you from zero to chatting with AI models on your own machine in about 10 minutes. No cloud, no API keys, no subscriptions — just you, your computer, and open-source AI.

> **Prerequisites:** You'll need **Go 1.22+** or a pre-built binary for your OS. Don't have Go? Jump straight to the [download page](https://github.com/MD-Mushfiqur123/lychee/releases) and grab the installer for Windows, macOS, or Linux.

---

## Step 1: Install Lychee 📥

Open your terminal and run:

```bash
go install github.com/MD-Mushfiqur123/lychee@latest
```

That's it. One command. Go will download, compile, and install Lychee into your `$GOPATH/bin` (usually `~/go/bin`).

Make sure `$GOPATH/bin` (or `%GOPATH%\bin` on Windows) is in your system `PATH` — the install script usually handles this, but if `lychee` isn't found after install, check that.

**Verify the install:**

```bash
lychee version
```

You should see something like:

```
Lychee  v0.9.0
  Commit:   a1b2c3d
  Build:    2026-06-17T12:00:00Z
  Go:       go1.23.4
  Platform: windows/amd64
```

> 💡 **Alternative installs:**
> - **Windows (PowerShell):** `irm https://raw.githubusercontent.com/MD-Mushfiqur123/lychee/main/scripts/install.ps1 | iex`
> - **macOS / Linux:** `curl -fsSL https://raw.githubusercontent.com/MD-Mushfiqur123/lychee/main/scripts/install.sh | sh`
> - **Docker:** `docker pull ghcr.io/md-mushfiqur123/lychee:latest`

---

## Step 2: Start the Lychee Server 🚀

Lychee runs as a local server that manages models, handles inference requests, and serves the web dashboard.

In a **new terminal window**, start the server:

```bash
lychee serve
```

You'll see output similar to:

```
time=2026-06-17T14:30:00.000+06:00 level=INFO msg="server config" env="map[...]"
time=2026-06-17T14:30:00.015+06:00 level=INFO msg="Lychee cloud disabled: true"
time=2026-06-17T14:30:00.082+06:00 level=INFO msg="Listening on 127.0.0.1:11434 (version 0.9.0)"
```

> 🟢 **Keep this terminal running!** The server needs to stay alive. Open a second terminal for the next steps, or run `lychee serve` in the background.

---

## Step 3: Pull Your First Model 🧠

Now let's download a model. We'll start with **Qwen2.5-3B** — it's small (~2 GB), fast, and performs remarkably well for its size. Perfect for learning.

In your second terminal, run:

```bash
lychee pull qwen2.5:3b
```

You'll see a progress bar as the model downloads:

```
pulling manifest ⠋
pulling 9f4b3a5c7d... 100% ▕████████████████████▏ 2.0 GB
pulling b2a8c1e3f9... 100% ▕████████████████████▏ 485 MB
pulling f0d3e6a1b8... 100% ▕████████████████████▏  32 KB
verifying sha256 digest ⠙
writing manifest ⠹
success ✔
```

> 🎉 **Congrats!** You just downloaded a full AI model to your machine. No internet needed from here on out — it's yours forever.

**Check what you have installed:**

```bash
lychee models
```

```
── Available Models ──

  • qwen2.5:3b    1.9 GB
```

---

## Step 4: Chat with Your Model 💬

Now the fun part. Send a message to your freshly downloaded model with a single command:

```bash
lychee run qwen2.5:3b "Hello! What can you tell me about yourself?"
```

You should see the model respond:

```
Hello! I'm Qwen2.5, an AI assistant created by Alibaba Cloud's Qwen team.
I'm a 3-billion-parameter language model designed to help with questions,
writing, coding, analysis, and creative tasks. I can communicate in multiple
languages and I'm running locally on your machine through Lychee — which means
our conversation is completely private and doesn't require an internet connection!
How can I help you today?
```

> 💡 **Pro tip:** Lychee automatically detects and launches the server if it's not already running. But keeping `lychee serve` running separately gives you the dashboard and better control.

**Try a coding example:**

```bash
lychee run qwen2.5:3b "Write a Python function that checks if a string is a palindrome"
```

**Or ask for a summary:**

```bash
lychee run qwen2.5:3b "Summarize the key differences between REST and GraphQL in 3 bullet points"
```

---

## Step 5: Enter Interactive Mode 🎮

For a richer experience, use the interactive REPL. It keeps the conversation going without re-typing the model name every time:

```bash
lychee interactive
```

You'll enter the Lychee REPL:

```
🤖 Lychee REPL — model: qwen2.5:3b
  Type /help for commands, /exit to quit, /model to switch

> Tell me a short joke about programming

  Why do programmers prefer dark mode?
  Because light attracts bugs! 🐛

> Now explain why that joke is funny

  The joke plays on two meanings of "bugs":
  1. Software bugs — errors or flaws in code that programmers
     spend a lot of time fixing.
  2. Real insects — which are actually attracted to light.
  
  The pun works because programmers often work late at night
  in dark-themed editors, and the unexpected twist of a "bug"
  being a literal insect is what makes it land.
```

**Useful REPL commands:**

| Command | What it does |
|:---|:---|
| `/exit` or `/quit` | Exit the REPL |
| `/help` | Show all available commands |
| `/model <name>` | Switch to a different model |
| `/models` | List all installed models |
| `/clear` | Clear the screen |
| `/system` | Show system info (GPU, RAM, CPU) |

**Switching models mid-conversation:**

```
> /model llama3.2
  ✓ Switched to: llama3.2
```

---

## Step 6: Explore the Dashboard 🌐

Lychee includes a built-in web dashboard. While the server is running (`lychee serve`), open your browser and navigate to:

```
http://localhost:11434
```

You'll see the **Lychee Server Dashboard** — a clean, dark-themed interface that shows:

- **🟢 Server status** — whether the server is running and healthy
- **📊 Active models** — which models are currently loaded in memory
- **🧠 Installed models** — all models you've pulled, with their sizes
- **⚡ GPU info** — GPU name, memory usage, and utilization (if applicable)
- **📈 Request stats** — recent inference activity

The dashboard is also a full **single-page application (SPA)** — you can use it to test chat completions, generate embeddings, and explore the API right from your browser.

> 🌐 The dashboard is served directly by Lychee — no extra install, no web server to configure. If you can reach `localhost:11434`, you're in.

---

## Step 7: Try Lychee Desktop (Windows GUI) 🖥️

Prefer a graphical interface? Lychee Desktop is a native Windows app that wraps the full Lychee engine in a clean, clickable UI.

**Download the latest release:**

🔗 **[Download Lychee Desktop](https://github.com/MD-Mushfiqur123/lychee-desktop/releases/latest)**

Grab `Lychee-Setup.exe`, run the installer, and you'll have:

- 📦 **One-click model browsing & pulling** — pick from a catalog, click download
- 💬 **Chat interface** — no terminal needed
- ⚙️ **Settings panel** — configure GPU, memory limits, and model storage
- 🔄 **Auto-updates** — stay on the latest version automatically

> 💡 The desktop app uses the same Lychee engine under the hood. Models you pull via CLI work in Desktop and vice versa — they share the same model directory (`~/.lychee/models`).

---

## Step 8: Next Steps 🗺️

You now have a fully functional local AI setup. Here's where to go next:

### 📚 Learn More

| Resource | Link |
|:---|:---|
| **Full Documentation** | [md-mushfiqur123.github.io/lychee-docs](https://md-mushfiqur123.github.io/lychee-docs/) |
| **CLI Reference** | [docs/cli-reference.md](https://github.com/MD-Mushfiqur123/lychee/blob/main/docs/cli-reference.md) |
| **API Documentation** | [docs/api.md](https://github.com/MD-Mushfiqur123/lychee/blob/main/docs/api.md) |
| **Models Guide** | [docs/models.md](https://github.com/MD-Mushfiqur123/lychee/blob/main/docs/models.md) |
| **Troubleshooting** | [docs/troubleshooting.md](https://github.com/MD-Mushfiqur123/lychee/blob/main/docs/troubleshooting.md) |
| **FAQ** | [docs/faq.mdx](https://md-mushfiqur123.github.io/lychee-docs/faq) |
| **Production Deployment** | [docs/production.md](https://github.com/MD-Mushfiqur123/lychee/blob/main/docs/production.md) |

### 💻 SDKs — Build with Lychee

Integrate Lychee into your own applications:

| Language | Install | Docs |
|:---|:---|:---|
| 🐍 **Python** | `pip install lychee-python` | [PyPI](https://pypi.org/project/lychee-python/) |
| 📦 **JavaScript / TypeScript** | `npm install lychee-js` | [npm](https://www.npmjs.com/package/lychee-js) |
| 💎 **Ruby** | `gem install lychee` | [RubyGems](https://rubygems.org/gems/lychee) |
| 🔵 **Go** | `go get github.com/MD-Mushfiqur123/lychee/sdk/go` | [pkg.go.dev](https://pkg.go.dev/github.com/MD-Mushfiqur123/lychee/sdk/go) |
| 🦀 **Rust** | [Source](https://github.com/MD-Mushfiqur123/lychee/tree/main/sdk/rust) | Cargo |
| 🟦 **VS Code Extension** | [Source](https://github.com/MD-Mushfiqur123/lychee/tree/main/sdk/vscode) | Marketplace |

### 🤝 Join the Community

| Platform | Link |
|:---|:---|
| 💬 **Discord** | [discord.gg/lychee](https://discord.gg/lychee) |
| 📖 **GitHub Discussions** | [github.com/MD-Mushfiqur123/lychee/discussions](https://github.com/MD-Mushfiqur123/lychee/discussions) |
| 🐛 **Report a Bug** | [github.com/MD-Mushfiqur123/lychee/issues](https://github.com/MD-Mushfiqur123/lychee/issues) |
| 🌟 **Star on GitHub** | [github.com/MD-Mushfiqur123/lychee](https://github.com/MD-Mushfiqur123/lychee) |

### 🔥 Try These Next

```bash
# Pull a bigger, more capable model
lychee pull llama3.2

# Pull from HuggingFace directly
lychee pull microsoft/Phi-3-mini-4k-instruct-gguf

# Use OpenAI-compatible API from your apps
curl http://localhost:11434/v1/chat/completions \
  -d '{"model":"qwen2.5:3b","messages":[{"role":"user","content":"Hello!"}]}'

# Try the Anthropic-compatible API
curl http://localhost:11434/v1/messages \
  -d '{"model":"qwen2.5:3b","messages":[{"role":"user","content":"Hello!"}]}'

# Launch Claude Code with Lychee as the backend
lychee launch claude

# Run OpenClaw (100+ skills AI assistant)
lychee launch openclaw

# Generate embeddings for search/RAG
curl http://localhost:11434/api/embed \
  -d '{"model":"nomic-embed-text","input":"Hello, world!"}'
```

---

## 🎉 You Did It!

You've installed Lychee, pulled an AI model, chatted with it, explored the dashboard, and know where to go next. You're now running a full AI stack **locally, privately, and for free** — no different than having your own personal ChatGPT that lives on your hard drive.

**What makes Lychee special:**

- 🔒 **100% private** — everything runs locally, your data never leaves your machine
- 🆓 **No API keys, no subscriptions** — all open-source models, forever free
- 🧩 **HuggingFace integration** — pull any GGUF model directly by repo path
- 🔌 **OpenAI & Anthropic API compatible** — drop-in replacement for existing apps
- 🌐 **Built-in dashboard** — manage everything from your browser
- 🖥️ **Desktop app available** — GUI option for Windows users
- 🛠️ **SDKs for 6 languages** — build apps on top of Lychee

Happy building! 🍒
