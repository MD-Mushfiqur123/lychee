/**
 * deploy-commands.js
 *
 * Standalone script to register slash commands with Discord.
 * Run once after initial setup or when commands change:
 *
 *   node deploy-commands.js
 */

require("dotenv").config();
const { REST, Routes } = require("discord.js");
const { SlashCommandBuilder } = require("discord.js");

// These must match the slashCommands array in index.js
const commands = [
  new SlashCommandBuilder()
    .setName("lychee-status")
    .setDescription("Show Lychee server status, uptime, and loaded models"),
  new SlashCommandBuilder()
    .setName("lychee-models")
    .setDescription("List locally installed Lychee models")
    .addStringOption((opt) =>
      opt
        .setName("query")
        .setDescription("Filter models by name (optional)")
        .setRequired(false)
    ),
  new SlashCommandBuilder()
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
  new SlashCommandBuilder()
    .setName("poll")
    .setDescription("Create a community poll")
    .addStringOption((opt) =>
      opt.setName("question").setDescription("The poll question").setRequired(true)
    )
    .addStringOption((opt) =>
      opt.setName("option1").setDescription("Option 1").setRequired(true)
    )
    .addStringOption((opt) =>
      opt.setName("option2").setDescription("Option 2").setRequired(true)
    )
    .addStringOption((opt) =>
      opt.setName("option3").setDescription("Option 3 (optional)").setRequired(false)
    )
    .addStringOption((opt) =>
      opt.setName("option4").setDescription("Option 4 (optional)").setRequired(false)
    )
    .addStringOption((opt) =>
      opt.setName("option5").setDescription("Option 5 (optional)").setRequired(false)
    ),
  new SlashCommandBuilder()
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
];

(async () => {
  const token = process.env.DISCORD_TOKEN;
  const clientId = process.env.DISCORD_CLIENT_ID;
  const guildId = process.env.DISCORD_GUILD_ID;

  if (!token || !clientId) {
    console.error(
      "❌ DISCORD_TOKEN and DISCORD_CLIENT_ID must be set in .env"
    );
    process.exit(1);
  }

  const rest = new REST({ version: "10" }).setToken(token);

  try {
    console.log(`📝 Registering ${commands.length} slash commands...`);

    if (guildId) {
      // Guild-specific (instant, great for testing)
      await rest.put(Routes.applicationGuildCommands(clientId, guildId), {
        body: commands,
      });
      console.log(`✅ Registered ${commands.length} guild commands instantly.`);
    } else {
      // Global (takes up to 1 hour to propagate)
      await rest.put(Routes.applicationCommands(clientId), {
        body: commands,
      });
      console.log(
        `✅ Registered ${commands.length} global commands (may take up to 1 hour to appear).`
      );
    }
  } catch (err) {
    console.error("❌ Failed to register commands:", err);
    process.exit(1);
  }
})();
