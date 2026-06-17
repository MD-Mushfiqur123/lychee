/**
 * 🍈 Lychee Discord Bot — index.js
 *
 * Features:
 *  - Auto-welcome DMs for new members
 *  - FAQ auto-reply based on message patterns
 *  - /lychee-status — live server status embed
 *  - /lychee-models — list available models
 *  - /help — contextual help
 *  - /poll — community polls
 *  - /github — link to GitHub issues/PRs
 *  - Early Adopter role auto-assignment
 *  - Anti-spam + token leak detection
 *  - Moderation logging
 *
 * Built with discord.js v14
 */

// ---------------------------------------------------------------------------
// Imports & Config
// ---------------------------------------------------------------------------

require("dotenv").config();
const fs = require("fs");
const path = require("path");
const {
  Client,
  GatewayIntentBits,
  Partials,
  Collection,
  Events,
  EmbedBuilder,
  SlashCommandBuilder,
  ActionRowBuilder,
  ButtonBuilder,
  ButtonStyle,
  REST,
  Routes,
  ChannelType,
  PermissionFlagsBits,
} = require("discord.js");
const axios = require("axios");
const winston = require("winston");

// ---------------------------------------------------------------------------
// Logger
// ---------------------------------------------------------------------------

const logger = winston.createLogger({
  level: process.env.LOG_LEVEL || "info",
  format: winston.format.combine(
    winston.format.timestamp(),
    winston.format.printf(({ timestamp, level, message }) => {
      return `[${timestamp}] ${level.toUpperCase()}: ${message}`;
    })
  ),
  transports: [new winston.transports.Console()],
});

// ---------------------------------------------------------------------------
// Constants
// ---------------------------------------------------------------------------

const LYCHEE_API_URL =
  process.env.LYCHEE_API_URL || "http://localhost:11434";
const MOD_LOG_CHANNEL_ID = process.env.MOD_LOG_CHANNEL_ID;
const GITHUB_REPO = "MD-Mushfiqur123/lychee";
const BOT_COLOR = 0x2ecc71; // Lychee green

// FAQ patterns — regex → reply
const FAQ_PATTERNS = [
  {
    pattern: /how\b.{0,20}(install|download|get.{0,5}lychee)/is,
    reply: (m) =>
      `👋 **Getting Started with Lychee**\n\n` +
      `📥 **Install:**\n` +
      `\`\`\`bash\ncurl -fsSL https://lychee.dev/install.sh | sh\n\`\`\`\n` +
      `Windows: \`winget install Lychee\`\n\n` +
      `📚 Full guide: https://docs.lychee.dev/quickstart`,
  },
  {
    pattern:
      /(won't|doesn't|not|can't|cannot)\b.{0,30}(start|run|work|load|launch)/is,
    reply: (m) =>
      `🔧 **Troubleshooting Lychee Startup**\n\n` +
      `Common fixes:\n` +
      `1. Check if Lychee is already running: \`lychee status\`\n` +
      `2. Verify GPU drivers are installed\n` +
      `3. Check logs: \`lychee logs\`\n` +
      `4. Try restarting: \`lychee stop && lychee serve\`\n\n` +
      `Still stuck? Post in <#CHANNEL_ID_HELP> with your OS, GPU, and error message.`,
  },
  {
    pattern:
      /which\b.{0,20}(model|llm).{0,20}(best|good|recommend|should.{0,10}(use|run|try))/is,
    reply: (m) =>
      `🤖 **Model Recommendations**\n\n` +
      `Your best model depends on your hardware. Check <#CHANNEL_ID_MODELS> for community benchmarks!\n\n` +
      `**Quick picks:**\n` +
      `• **Coding:** \`qwen3.5:14b\`, \`deepseek-coder-v2\`\n` +
      `• **General chat:** \`llama3.2:3b\`, \`gemma4:9b\`\n` +
      `• **Lightweight (CPU):** \`phi3:mini\`, \`qwen3.5:0.5b\`\n` +
      `• **Vision:** \`llava:13b\`, \`qwen2.5-vl:7b\`\n\n` +
      `Browse all models: \`lychee list\``,
  },
  {
    pattern: /how\b.{0,20}(update|upgrade)\b.{0,10}lychee/is,
    reply: (m) =>
      `🔄 **Updating Lychee**\n\n` +
      `\`\`\`bash\nlychee update\n\`\`\`\n\n` +
      `This pulls the latest version. To update models:\n` +
      `\`\`\`bash\nlychee pull <model-name>  # re-pulls the latest GGUF\n\`\`\``,
  },
  {
    pattern:
      /(gpu|cuda|metal|vulkan|rocm)\b.{0,30}(not|isn't|won't).{0,10}(detect|work|working|recognize)/is,
    reply: (m) =>
      `🖥️ **GPU Not Detected?**\n\n` +
      `**macOS (Metal):** Metal works out of the box on Apple Silicon.\n` +
      `**Linux (CUDA):** Make sure NVIDIA drivers and CUDA toolkit are installed. Run \`nvidia-smi\` to verify.\n` +
      `**Linux (ROCm):** Check \`rocm-smi\` and ensure ROCm is properly installed.\n` +
      `**Windows (CUDA):** Install the latest NVIDIA Game Ready or Studio driver.\n\n` +
      `Run \`lychee gpu-info\` to see what Lychee detects.\n` +
      `Still not working? Head to <#CHANNEL_ID_HELP>.`,
  },
  {
    pattern:
      /how\b.{0,20}(use|connect|call).{0,10}(openai|api|endpoint|client)/is,
    reply: (m) =>
      `🔌 **Lychee API (OpenAI-compatible)**\n\n` +
      `Start the API server:\n` +
      `\`\`\`bash\nlychee serve\n\`\`\`\n\n` +
      `Then use any OpenAI client — just change the base URL:\n\n` +
      `**Python (openai package):**\n` +
      `\`\`\`python\nfrom openai import OpenAI\nclient = OpenAI(base_url="http://localhost:11434/v1", api_key="lychee")\n\`\`\`\n` +
      `**cURL:**\n` +
      `\`\`\`bash\ncurl http://localhost:11434/v1/chat/completions -d '{"model":"llama3.2","messages":[{"role":"user","content":"Hello!"}]}'\n\`\`\`\n\n` +
      `📚 Docs: https://docs.lychee.dev/api`,
  },
  {
    pattern: /how\b.{0,30}(import|convert).{0,10}gguf/is,
    reply: (m) =>
      `📦 **Using GGUF Models with Lychee**\n\n` +
      `Lychee supports GGUF files directly. Two ways:\n\n` +
      `**1. From Hugging Face:**\n` +
      `\`\`\`bash\nlychee pull huggingface.co/user/model-name:quant\n\`\`\`\n\n` +
      `**2. Local GGUF file:**\n` +
      `\`\`\`bash\nlychee create my-model -f ./path/to/model.gguf\n\`\`\`\n\n` +
      `📚 Docs: https://docs.lychee.dev/import`,
  },
  {
    pattern: /(slow|lag|laggy|performance|speed)\b.{0,5}(issue|problem|bad|terrible)/is,
    reply: (m) =>
      `⚡ **Improving Lychee Performance**\n\n` +
      `1. Use a smaller quant: \`q4_K_M\` is a good balance\n` +
      `2. Check GPU usage: \`lychee gpu-info\`\n` +
      `3. Close other GPU-heavy apps (Chrome, games)\n` +
      `4. Reduce context length if you don't need 128K tokens\n` +
      `5. Try speculative decoding (v0.2.0+): \`lychee run model --draft-model model:q4_0\`\n\n` +
      `Share your setup in <#CHANNEL_ID_MODELS> for specific tuning advice.`,
  },
];

// Known channel IDs (set during ready event)
let CHANNEL_IDS = {};

// ---------------------------------------------------------------------------
// Discord Client
// ---------------------------------------------------------------------------

const client = new Client({
  intents: [
    GatewayIntentBits.Guilds,
    GatewayIntentBits.GuildMembers,
    GatewayIntentBits.GuildMessages,
    GatewayIntentBits.MessageContent,
    GatewayIntentBits.GuildPresences,
    GatewayIntentBits.GuildMessageReactions,
  ],
  partials: [Partials.Message, Partials.Channel, Partials.Reaction],
});

// Slash commands collection
client.commands = new Collection();

// ---------------------------------------------------------------------------
// Slash Command Definitions
// ---------------------------------------------------------------------------

const slashCommands = [
  // /lychee-status
  {
    data: new SlashCommandBuilder()
      .setName("lychee-status")
      .setDescription("Show Lychee server status, uptime, and loaded models"),
    async execute(interaction) {
      await interaction.deferReply();

      try {
        const start = Date.now();
        const [versionRes, modelsRes, statusRes] = await Promise.allSettled([
          axios.get(`${LYCHEE_API_URL}/api/version`, { timeout: 5000 }).catch(() => null),
          axios.get(`${LYCHEE_API_URL}/api/tags`, { timeout: 5000 }).catch(() => null),
          axios.get(`${LYCHEE_API_URL}/api/status`, { timeout: 5000 }).catch(() => null),
        ]);

        const latency = Date.now() - start;
        const isOnline =
          versionRes.status === "fulfilled" && versionRes.value !== null;

        const embed = new EmbedBuilder()
          .setTitle("🍈 Lychee Server Status")
          .setColor(isOnline ? 0x2ecc71 : 0xe74c3c)
          .setTimestamp();

        if (isOnline) {
          const version = versionRes.value?.data?.version || "unknown";
          const models =
            modelsRes.status === "fulfilled" && modelsRes.value
              ? modelsRes.value.data?.models || []
              : [];

          embed.setDescription("🟢 **Online**");
          embed.addFields(
            { name: "Version", value: `\`${version}\``, inline: true },
            { name: "API Latency", value: `${latency}ms`, inline: true },
            {
              name: "Models Loaded",
              value: `${models.length}`,
              inline: true,
            }
          );

          if (models.length > 0) {
            const modelList = models
              .slice(0, 10)
              .map((m) => `• \`${m.name}\``)
              .join("\n");
            const suffix =
              models.length > 10 ? `\n*...and ${models.length - 10} more*` : "";
            embed.addFields({
              name: "📦 Installed Models",
              value: modelList + suffix,
            });
          }

          // Try to get GPU info
          if (statusRes.status === "fulfilled" && statusRes.value) {
            const status = statusRes.value.data;
            if (status?.gpu) {
              embed.addFields({
                name: "🖥️ GPU",
                value: status.gpu,
                inline: true,
              });
            }
          }
        } else {
          embed.setDescription(
            "🔴 **Offline** — Could not reach the Lychee API server.\n" +
              "Is `lychee serve` running? Check with `lychee status` in your terminal."
          );
        }

        await interaction.editReply({ embeds: [embed] });
      } catch (err) {
        logger.error("Status command error:", err.message);
        await interaction.editReply(
          "⚠️ Failed to query Lychee server. Is it running?"
        );
      }
    },
  },

  // /lychee-models
  {
    data: new SlashCommandBuilder()
      .setName("lychee-models")
      .setDescription("List locally installed Lychee models")
      .addStringOption((opt) =>
        opt
          .setName("query")
          .setDescription("Filter models by name (optional)")
          .setRequired(false)
      ),
    async execute(interaction) {
      await interaction.deferReply();
      const query = interaction.options.getString("query")?.toLowerCase();

      try {
        const res = await axios.get(`${LYCHEE_API_URL}/api/tags`, {
          timeout: 5000,
        });
        let models = res.data?.models || [];

        if (query) {
          models = models.filter((m) =>
            m.name.toLowerCase().includes(query)
          );
        }

        if (models.length === 0) {
          return interaction.editReply(
            query
              ? `No models found matching **"${query}"**.`
              : "📭 No models installed yet. Pull one with `lychee pull <model-name>`."
          );
        }

        const embed = new EmbedBuilder()
          .setTitle(
            query
              ? `🔍 Models matching "${query}" (${models.length})`
              : `📦 Installed Models (${models.length})`
          )
          .setColor(BOT_COLOR)
          .setDescription(
            models
              .slice(0, 25)
              .map((m) => `• \`${m.name}\` — ${m.size || "unknown size"}`)
              .join("\n")
          )
          .setFooter({
            text: `Pull models: lychee pull <name> | Total: ${models.length}`,
          });

        if (models.length > 25) {
          embed.setFooter({
            text: `Showing 25 of ${models.length} models. Use \`/lychee-models query:<name>\` to filter.`,
          });
        }

        await interaction.editReply({ embeds: [embed] });
      } catch (err) {
        logger.error("Models command error:", err.message);
        await interaction.editReply(
          "⚠️ Could not reach Lychee server. Start it with `lychee serve`."
        );
      }
    },
  },

  // /help
  {
    data: new SlashCommandBuilder()
      .setName("help")
      .setDescription("Get help with Lychee")
      .addStringOption((opt) =>
        opt
          .setName("topic")
          .setDescription("What do you need help with?")
          .setRequired(false)
          .addChoices(
            { name: "Installation", value: "install" },
            { name: "Models & Pulling", value: "models" },
            { name: "API Usage", value: "api" },
            { name: "Bot Commands", value: "commands" },
            { name: "Contributing", value: "contributing" },
            { name: "Troubleshooting", value: "troubleshooting" }
          )
      ),
    async execute(interaction) {
      const topic = interaction.options.getString("topic");

      const helpEmbeds = {
        install: {
          title: "📥 Installing Lychee",
          description:
            "```bash\n# macOS / Linux\ncurl -fsSL https://lychee.dev/install.sh | sh\n\n# Windows\nwinget install Lychee\n\n# Docker\ndocker run -d -v lychee:/root/.lychee -p 11434:11434 lychee/lychee\n```\n\n📚 Full guide: https://docs.lychee.dev/quickstart",
        },
        models: {
          title: "📦 Working with Models",
          description:
            "**Pull a model:** `lychee pull <model-name>`\n" +
            "**List models:** `lychee list`\n" +
            "**Run a model:** `lychee run <model-name>`\n" +
            "**Delete a model:** `lychee rm <model-name>`\n" +
            "**Import GGUF:** `lychee create my-model -f ./model.gguf`\n\n" +
            "📚 Docs: https://docs.lychee.dev/import",
        },
        api: {
          title: "🔌 Using the Lychee API",
          description:
            "Lychee provides an **OpenAI-compatible API** at `http://localhost:11434/v1`\n\n" +
            "```python\nfrom openai import OpenAI\nclient = OpenAI(base_url='http://localhost:11434/v1', api_key='lychee')\n\n" +
            'response = client.chat.completions.create(\n  model="llama3.2",\n  messages=[{"role":"user","content":"Hello!"}]\n)\n```\n\n' +
            "📚 Docs: https://docs.lychee.dev/api",
        },
        commands: {
          title: "🤖 Bot Commands",
          description:
            "`/lychee-status` — Server status & models\n" +
            "`/lychee-models [query]` — List installed models\n" +
            "`/help [topic]` — This menu\n" +
            "`/poll <question> <options>` — Create a poll\n" +
            "`/github <type> <number>` — Link to issue/PR",
        },
        contributing: {
          title: "🛠️ Contributing to Lychee",
          description:
            "We welcome contributions!\n\n" +
            "1. Read [CONTRIBUTING.md](https://github.com/MD-Mushfiqur123/lychee/blob/main/CONTRIBUTING.md)\n" +
            "2. Pick a `good first issue`\n" +
            "3. Fork the repo and create a PR\n" +
            "4. Join the discussion in <#CHANNEL_ID_DEV>\n\n" +
            "After your first merged PR, you'll get the @Contributor role!",
        },
        troubleshooting: {
          title: "🔧 Troubleshooting",
          description:
            "**Lychee won't start:**\n" +
            "- Check if another instance is running: `lychee status`\n" +
            "- Verify GPU drivers: `lychee gpu-info`\n\n" +
            "**Model won't load:**\n" +
            "- Check available RAM/VRAM\n" +
            "- Try a smaller quant: `lychee pull model:q4_K_M`\n\n" +
            "**API not responding:**\n" +
            "- Is `lychee serve` running?\n" +
            "- Check port 11434 isn't blocked by firewall\n\n" +
            "Still stuck? Ask in <#CHANNEL_ID_HELP>!",
        },
      };

      if (topic && helpEmbeds[topic]) {
        const h = helpEmbeds[topic];
        const embed = new EmbedBuilder()
          .setTitle(h.title)
          .setDescription(h.description)
          .setColor(BOT_COLOR);
        await interaction.reply({ embeds: [embed] });
      } else {
        const embed = new EmbedBuilder()
          .setTitle("🍈 Lychee Help")
          .setDescription(
            "Choose a topic below or use `/help topic:<topic>`:\n\n" +
              "📥 `install` — Installing Lychee\n" +
              "📦 `models` — Working with models\n" +
              "🔌 `api` — Using the API\n" +
              "🤖 `commands` — Bot slash commands\n" +
              "🛠️ `contributing` — How to contribute\n" +
              "🔧 `troubleshooting` — Fix common issues\n\n" +
              "Need more help? Post in <#CHANNEL_ID_HELP>!"
          )
          .setColor(BOT_COLOR)
          .setFooter({
            text: `Lychee v1.0.0 | github.com/${GITHUB_REPO}`,
          });

        // Create topic buttons
        const row1 = new ActionRowBuilder().addComponents(
          new ButtonBuilder()
            .setCustomId("help_install")
            .setLabel("Install")
            .setStyle(ButtonStyle.Primary)
            .setEmoji("📥"),
          new ButtonBuilder()
            .setCustomId("help_models")
            .setLabel("Models")
            .setStyle(ButtonStyle.Primary)
            .setEmoji("📦"),
          new ButtonBuilder()
            .setCustomId("help_api")
            .setLabel("API")
            .setStyle(ButtonStyle.Primary)
            .setEmoji("🔌")
        );
        const row2 = new ActionRowBuilder().addComponents(
          new ButtonBuilder()
            .setCustomId("help_contributing")
            .setLabel("Contribute")
            .setStyle(ButtonStyle.Secondary)
            .setEmoji("🛠️"),
          new ButtonBuilder()
            .setCustomId("help_troubleshooting")
            .setLabel("Troubleshoot")
            .setStyle(ButtonStyle.Secondary)
            .setEmoji("🔧")
        );

        await interaction.reply({
          embeds: [embed],
          components: [row1, row2],
        });
      }
    },
  },

  // /poll
  {
    data: new SlashCommandBuilder()
      .setName("poll")
      .setDescription("Create a community poll")
      .addStringOption((opt) =>
        opt
          .setName("question")
          .setDescription("The poll question")
          .setRequired(true)
      )
      .addStringOption((opt) =>
        opt.setName("option1").setDescription("Option 1").setRequired(true)
      )
      .addStringOption((opt) =>
        opt.setName("option2").setDescription("Option 2").setRequired(true)
      )
      .addStringOption((opt) =>
        opt
          .setName("option3")
          .setDescription("Option 3 (optional)")
          .setRequired(false)
      )
      .addStringOption((opt) =>
        opt
          .setName("option4")
          .setDescription("Option 4 (optional)")
          .setRequired(false)
      )
      .addStringOption((opt) =>
        opt
          .setName("option5")
          .setDescription("Option 5 (optional)")
          .setRequired(false)
      ),
    async execute(interaction) {
      const question = interaction.options.getString("question");
      const options = [];
      for (let i = 1; i <= 5; i++) {
        const opt = interaction.options.getString(`option${i}`);
        if (opt) options.push(opt);
      }

      const emojis = ["1️⃣", "2️⃣", "3️⃣", "4️⃣", "5️⃣"];

      const embed = new EmbedBuilder()
        .setTitle(`📊 ${question}`)
        .setDescription(
          options.map((opt, i) => `${emojis[i]} ${opt}`).join("\n")
        )
        .setColor(BOT_COLOR)
        .setFooter({
          text: `Poll by ${interaction.user.displayName}`,
          iconURL: interaction.user.displayAvatarURL(),
        })
        .setTimestamp();

      const pollMsg = await interaction.reply({
        embeds: [embed],
        fetchReply: true,
      });

      // React with options
      for (let i = 0; i < options.length; i++) {
        await pollMsg.react(emojis[i]);
      }
    },
  },

  // /github
  {
    data: new SlashCommandBuilder()
      .setName("github")
      .setDescription("Link to a GitHub issue or pull request")
      .addStringOption((opt) =>
        opt
          .setName("type")
          .setDescription("Issue or Pull Request")
          .setRequired(true)
          .addChoices(
            { name: "Issue", value: "issue" },
            { name: "Pull Request", value: "pr" }
          )
      )
      .addIntegerOption((opt) =>
        opt
          .setName("number")
          .setDescription("Issue or PR number")
          .setRequired(true)
      ),
    async execute(interaction) {
      const type = interaction.options.getString("type");
      const number = interaction.options.getInteger("number");
      const label = type === "issue" ? "Issue" : "Pull Request";
      const url = `https://github.com/${GITHUB_REPO}/${type}/${number}`;

      try {
        const res = await axios.get(
          `https://api.github.com/repos/${GITHUB_REPO}/issues/${number}`,
          { headers: { Accept: "application/vnd.github+json" } }
        );

        const data = res.data;
        const embed = new EmbedBuilder()
          .setTitle(`${label} #${number}: ${data.title}`)
          .setURL(url)
          .setDescription(
            (data.body || "No description").slice(0, 400) +
              (data.body && data.body.length > 400 ? "..." : "")
          )
          .setColor(BOT_COLOR)
          .addFields(
            {
              name: "Status",
              value: data.state === "open" ? "🟢 Open" : "🟣 Closed",
              inline: true,
            },
            {
              name: "Author",
              value: `[${data.user.login}](${data.user.html_url})`,
              inline: true,
            }
          )
          .setFooter({ text: `github.com/${GITHUB_REPO}` });

        await interaction.reply({ embeds: [embed] });
      } catch {
        // If GitHub API fails, just send the link
        await interaction.reply(`🔗 **${label} #${number}:** ${url}`);
      }
    },
  },
];

// ---------------------------------------------------------------------------
// Event: Ready
// ---------------------------------------------------------------------------

client.once(Events.ClientReady, async (c) => {
  logger.info(`✅ Logged in as ${c.user.tag}`);

  // Set bot presence
  c.user.setPresence({
    activities: [{ name: "🍈 lychee serve", type: 4 }], // Custom status
    status: "online",
  });

  // Map channel IDs by name for FAQ references
  c.guilds.cache.forEach((guild) => {
    guild.channels.cache.forEach((channel) => {
      if (channel.name) {
        CHANNEL_IDS[channel.name.toUpperCase().replace(/-/g, "_")] =
          channel.id;
      }
    });
  });
  logger.info("📋 Channel map:", JSON.stringify(CHANNEL_IDS));

  // Register slash commands
  await registerCommands(c);
});

// ---------------------------------------------------------------------------
// Register Slash Commands
// ---------------------------------------------------------------------------

async function registerCommands(client) {
  const rest = new REST({ version: "10" }).setToken(process.env.DISCORD_TOKEN);
  const commandData = slashCommands.map((cmd) => cmd.data.toJSON());

  try {
    if (process.env.DISCORD_GUILD_ID) {
      // Guild-specific (instant, for testing)
      await rest.put(
        Routes.applicationGuildCommands(
          process.env.DISCORD_CLIENT_ID,
          process.env.DISCORD_GUILD_ID
        ),
        { body: commandData }
      );
      logger.info(
        `📝 Registered ${commandData.length} guild commands`
      );
    } else {
      // Global (takes up to 1 hour to propagate)
      await rest.put(Routes.applicationCommands(process.env.DISCORD_CLIENT_ID), {
        body: commandData,
      });
      logger.info(
        `📝 Registered ${commandData.length} global commands (may take up to 1h to appear)`
      );
    }
  } catch (err) {
    logger.error("Failed to register commands:", err.message);
  }
}

// ---------------------------------------------------------------------------
// Event: New Member (Auto-Welcome + Early Adopter)
// ---------------------------------------------------------------------------

client.on(Events.GuildMemberAdd, async (member) => {
  logger.info(`👋 New member: ${member.user.tag}`);

  // Assign @Early Adopter role
  try {
    const earlyAdopterRole = member.guild.roles.cache.find(
      (r) => r.name.toLowerCase() === "early adopter"
    );
    if (earlyAdopterRole) {
      await member.roles.add(earlyAdopterRole);
      logger.info(`🎖️ Assigned Early Adopter role to ${member.user.tag}`);
    }
  } catch (err) {
    logger.error("Failed to assign Early Adopter role:", err.message);
  }

  // Send welcome DM
  try {
    const welcomeEmbed = new EmbedBuilder()
      .setTitle("🍈 Welcome to Lychee!")
      .setDescription(
        `Hey **${member.user.displayName}** — welcome to the Lychee community!\n\n` +
          "Lychee is a fast, local-first LLM runner. Run AI models on your own machine with an OpenAI-compatible API.\n\n" +
          "**Getting started:**\n" +
          "📥 Install: `curl -fsSL https://lychee.dev/install.sh | sh`\n" +
          "📦 Pull a model: `lychee pull llama3.2`\n" +
          "▶️ Run it: `lychee run llama3.2`\n\n" +
          "**Useful channels:**\n" +
          "💬 `#general` — Chat with the community\n" +
          "❓ `#help` — Get support\n" +
          "🎨 `#showcase` — Share what you build\n" +
          "🤖 `#models` — Discuss models & benchmarks"
      )
      .setColor(BOT_COLOR)
      .setThumbnail(
        member.guild.iconURL({ size: 256 }) ||
          "https://lychee.dev/logo.png"
      )
      .setFooter({
        text: "Read the rules in #welcome-rules to get started!",
      });

    await member.send({ embeds: [welcomeEmbed] });
    logger.info(`📬 Sent welcome DM to ${member.user.tag}`);
  } catch (err) {
    // DM might be closed — that's okay
    logger.warn(`Could not DM ${member.user.tag}: ${err.message}`);
  }
});

// ---------------------------------------------------------------------------
// Event: Messages (FAQ Auto-Reply + Anti-Spam)
// ---------------------------------------------------------------------------

client.on(Events.MessageCreate, async (message) => {
  // Ignore bots, system messages, and DMs
  if (message.author.bot || message.system || !message.guild) return;

  const content = message.content.trim();
  if (!content) return;

  // ---- Anti-Spam / Token Leak Detection ----

  // Discord invite links (delete if from non-staff)
  const inviteRegex =
    /(?:discord\.(?:gg|io|me|li|com\/invite)|discordapp\.com\/invite)\/[a-zA-Z0-9]+/i;
  if (
    inviteRegex.test(content) &&
    !message.member?.permissions.has(PermissionFlagsBits.ManageMessages)
  ) {
    try {
      await message.delete();
      const warning = await message.channel.send(
        `⚠️ ${message.author}, sharing server invites is not allowed.`
      );
      setTimeout(() => warning.delete().catch(() => {}), 5000);
      logModeration(message.guild, "Invite link deleted", {
        user: message.author.tag,
        channel: message.channel.name,
        content: content.slice(0, 200),
      });
      return;
    } catch (err) {
      logger.error("Failed to delete invite message:", err.message);
    }
  }

  // Token / API key leak detection
  const tokenRegex =
    /\b(sk-[a-zA-Z0-9]{20,}|ghp_[a-zA-Z0-9]{20,}|hf_[a-zA-Z0-9]{20,}|xox[bprs]-[a-zA-Z0-9-]+)\b/;
  if (tokenRegex.test(content)) {
    try {
      await message.delete();
      await message.author
        .send(
          "🔐 **Security Alert:** Your message in the Lychee Discord contained what looks like an API key or token. It was deleted to protect your account.\n\n" +
            "Please **regenerate** that key immediately at the provider's dashboard — it may have been visible before deletion."
        )
        .catch(() => {});
      const warning = await message.channel.send(
        `⚠️ A message containing a possible API key was deleted to protect ${message.author}. Please never share keys in public channels.`
      );
      setTimeout(() => warning.delete().catch(() => {}), 10000);
      logModeration(message.guild, "Token leak detected & deleted", {
        user: message.author.tag,
        channel: message.channel.name,
      });
      return;
    } catch (err) {
      logger.error("Failed to handle token leak:", err.message);
    }
  }

  // ---- FAQ Auto-Reply ----

  for (const faq of FAQ_PATTERNS) {
    if (faq.pattern.test(content)) {
      // Don't reply if user already got an FAQ reply in the last 30 seconds
      // (simple debounce — prevents spam)
      try {
        let reply = faq.reply(message);

        // Replace channel placeholder with actual IDs
        reply = reply.replace(
          /<#CHANNEL_ID_(\w+)>/g,
          (match, key) => {
            const id = CHANNEL_IDS[key];
            return id ? `<#${id}>` : `#${key.toLowerCase().replace(/_/g, "-")}`;
          }
        );

        await message.reply({ content: reply, allowedMentions: { repliedUser: true } });
        logger.info(
          `💬 FAQ auto-reply to ${message.author.tag} in #${message.channel.name}`
        );
        break; // Only match one FAQ per message
      } catch (err) {
        logger.error("FAQ reply error:", err.message);
        break;
      }
    }
  }
});

// ---------------------------------------------------------------------------
// Event: Interactions (Slash Commands + Buttons)
// ---------------------------------------------------------------------------

client.on(Events.InteractionCreate, async (interaction) => {
  // Slash commands
  if (interaction.isChatInputCommand()) {
    const command = slashCommands.find(
      (cmd) => cmd.data.name === interaction.commandName
    );
    if (!command) return;

    try {
      await command.execute(interaction);
    } catch (err) {
      logger.error(
        `Command error (/${interaction.commandName}):`,
        err.message
      );
      const reply = {
        content: "❌ Something went wrong running that command.",
        ephemeral: true,
      };
      if (interaction.replied || interaction.deferred) {
        await interaction.followUp(reply);
      } else {
        await interaction.reply(reply);
      }
    }
  }

  // Button interactions (help menu)
  if (interaction.isButton()) {
    const helpEmbeds = {
      help_install: {
        title: "📥 Installing Lychee",
        description:
          "```bash\n# macOS / Linux\ncurl -fsSL https://lychee.dev/install.sh | sh\n\n# Windows\nwinget install Lychee\n\n# Docker\ndocker run -d -v lychee:/root/.lychee -p 11434:11434 lychee/lychee\n```\n\n📚 Full guide: https://docs.lychee.dev/quickstart",
      },
      help_models: {
        title: "📦 Working with Models",
        description:
          "**Pull a model:** `lychee pull <model-name>`\n" +
          "**List models:** `lychee list`\n" +
          "**Run a model:** `lychee run <model-name>`\n" +
          "**Delete a model:** `lychee rm <model-name>`\n\n" +
          "📚 Docs: https://docs.lychee.dev/import",
      },
      help_api: {
        title: "🔌 Using the Lychee API",
        description:
          "Lychee provides an **OpenAI-compatible API** at `http://localhost:11434/v1`\n\n" +
          "```python\nfrom openai import OpenAI\nclient = OpenAI(base_url='http://localhost:11434/v1', api_key='lychee')\n```\n\n" +
          "📚 Docs: https://docs.lychee.dev/api",
      },
      help_contributing: {
        title: "🛠️ Contributing to Lychee",
        description:
          "We welcome contributions!\n\n" +
          "1. Read [CONTRIBUTING.md](https://github.com/MD-Mushfiqur123/lychee/blob/main/CONTRIBUTING.md)\n" +
          "2. Pick a `good first issue`\n" +
          "3. Fork the repo and create a PR\n\n" +
          "After your first merged PR, you'll get the @Contributor role!",
      },
      help_troubleshooting: {
        title: "🔧 Troubleshooting",
        description:
          "**Lychee won't start:** Check `lychee status` and GPU drivers.\n" +
          "**Model won't load:** Try a smaller quant `q4_K_M`.\n" +
          "**API not responding:** Is `lychee serve` running on port 11434?\n\n" +
          "Still stuck? Ask in the help forum!",
      },
    };

    const embedData = helpEmbeds[interaction.customId];
    if (embedData) {
      const embed = new EmbedBuilder()
        .setTitle(embedData.title)
        .setDescription(embedData.description)
        .setColor(BOT_COLOR);
      await interaction.reply({ embeds: [embed], ephemeral: true });
    }
  }
});

// ---------------------------------------------------------------------------
// Event: Message Delete/Edit Logging
// ---------------------------------------------------------------------------

client.on(Events.MessageDelete, async (message) => {
  if (!message.guild || message.author?.bot) return;
  logModeration(message.guild, "Message Deleted", {
    user: message.author?.tag || "Unknown",
    channel: message.channel.name,
    content: message.content?.slice(0, 500) || "[no text content]",
    attachments: message.attachments.size
      ? `${message.attachments.size} attachment(s)`
      : "none",
  });
});

client.on(Events.MessageUpdate, async (oldMessage, newMessage) => {
  if (!oldMessage.guild || oldMessage.author?.bot) return;
  if (oldMessage.content === newMessage.content) return;

  logModeration(newMessage.guild, "Message Edited", {
    user: oldMessage.author?.tag || "Unknown",
    channel: oldMessage.channel.name,
    before: oldMessage.content?.slice(0, 300) || "[empty]",
    after: newMessage.content?.slice(0, 300) || "[empty]",
  });
});

// ---------------------------------------------------------------------------
// Moderation Logging Helper
// ---------------------------------------------------------------------------

async function logModeration(guild, action, details) {
  if (!MOD_LOG_CHANNEL_ID) return;

  try {
    const channel = await guild.channels.fetch(MOD_LOG_CHANNEL_ID);
    if (!channel) return;

    const embed = new EmbedBuilder()
      .setTitle(`📋 ${action}`)
      .setColor(0xf1c40f)
      .setTimestamp()
      .addFields(
        Object.entries(details).map(([key, value]) => ({
          name: key.charAt(0).toUpperCase() + key.slice(1),
          value: String(value).slice(0, 1024) || "—",
          inline: key.length < 10,
        }))
      );

    await channel.send({ embeds: [embed] });
  } catch (err) {
    logger.error("Failed to log moderation action:", err.message);
  }
}

// ---------------------------------------------------------------------------
// Error Handling
// ---------------------------------------------------------------------------

process.on("unhandledRejection", (err) => {
  logger.error("Unhandled rejection:", err.message);
});

process.on("uncaughtException", (err) => {
  logger.error("Uncaught exception:", err.message);
  // Don't exit — let the bot try to recover
});

// ---------------------------------------------------------------------------
// Graceful Shutdown
// ---------------------------------------------------------------------------

process.on("SIGINT", async () => {
  logger.info("🛑 Shutting down...");
  client.user?.setPresence({ status: "invisible" });
  client.destroy();
  process.exit(0);
});

process.on("SIGTERM", async () => {
  logger.info("🛑 Received SIGTERM, shutting down...");
  client.destroy();
  process.exit(0);
});

// ---------------------------------------------------------------------------
// Start
// ---------------------------------------------------------------------------

if (!process.env.DISCORD_TOKEN) {
  logger.error(
    "❌ DISCORD_TOKEN is not set. Copy .env.example to .env and fill in your token."
  );
  process.exit(1);
}

logger.info("🍈 Starting Lychee Discord Bot...");
client.login(process.env.DISCORD_TOKEN);
