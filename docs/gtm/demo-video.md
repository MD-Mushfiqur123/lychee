# Lychee Product Demo Video: Storyboard & Script

This document details the storyboard and script for a 2-minute product showcase video demonstrating Lychee's capabilities.

---

## Part 1: Introduction (0:00 - 0:30)
- **Visual**: Screen capture showing Ollama CLI, then running a command to download and run Lychee.
- **Audio/Voiceover**: "Local language models are incredibly powerful, but serving them in production with Ollama often runs into concurrency bottlenecks and tool limits. Meet Lychee: a drop-in, 100% backward-compatible Ollama fork designed for speed, compatibility, and real-world agents."
- **Visual**: Type `lychee run llama3` to show compatibility with the standard CLI experience.

## Part 2: High Concurrency (0:30 - 1:00)
- **Visual**: Split-screen showing a concurrency tool sending requests to Ollama vs Lychee.
- **Audio/Voiceover**: "Under the hood, Lychee introduces HTTP/2 multiplexing and sync.Pool buffer management. Instead of stalling concurrent streams or causing garbage collection pauses, Lychee keeps streams fluid and cuts P99 stream latency by over 30%."
- **Visual**: Show the built-in observability dashboard at `http://localhost:11434/dashboard` with requests updating in real-time.

## Part 3: Native Agents & Tools (1:00 - 1:30)
- **Visual**: Terminal running `lychee agent "Search the web for the latest LLM news and summarize it"`.
- **Audio/Voiceover**: "Lychee goes beyond simple completion APIs by embedding a secure runtime for autonomous agents. With out-of-the-box tools for web search, page fetching, and terminal sandboxing, you can invoke agents directly from the command line."
- **Visual**: Show the agent successfully using `google` search tool and displaying the summarized output.

## Part 4: Conclusion (1:30 - 2:00)
- **Visual**: Swapping the base URL in a Python script using the Anthropic Claude SDK to `http://localhost:11434/anthropic/v1`.
- **Audio/Voiceover**: "Best of all, Lychee is fully compatible with OpenAI and Anthropic SDKs. Swap your backend URL and target local models in seconds. Accelerate your LLM infrastructure today with Lychee."
- **Visual**: Show Github repo URL (https://github.com/MD-Mushfiqur123/lychee) and a call-to-action to star the project.
