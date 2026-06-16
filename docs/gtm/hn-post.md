# Show HN: Lychee â€“ A performance-optimized Ollama fork with HTTP/2, Anthropic compatibility, and built-in agents

Hi HN,

We've been working on Lychee (https://github.com/MD-Mushfiqur123/lychee), a fork of Ollama designed for production environments requiring high-concurrency streaming, cross-provider compatibility, and extensible agentic tools.

While Ollama is fantastic for local development, we ran into severe scaling bottlenecks and protocol limitations when trying to deploy it as a multi-user backend:
1. **HTTP/1.1 Head-of-Line Blocking**: Ollama serializes concurrent streamed chat responses because it lacks HTTP/2 multiplexing. Under heavy load, response streams stall.
2. **GC Memory Spikes**: Every request allocates massive JSON encoding buffers, triggering Go garbage collection pauses that halt active inference loops.
3. **Lack of Native Agent Tooling**: Running agents requires wrappers that expose raw shell or search capabilities.
4. **Provider Lock-in**: Developers writing against OpenAI or Anthropic SDKs have to rewrite their chat integration code.

Here's how we solved these in Lychee:
- **HTTP/2 Cleartext (h2c) Streaming**: Built full HTTP/2 stream multiplexing into the API server. Concurrent streams run side-by-side on a single TCP connection.
- **sync.Pool Buffer Reuse**: Recycled JSON encoding buffers using a thread-safe `sync.Pool`, reducing garbage collector latency by over 90% under high request concurrency.
- **Embedded Agent Runtime**: Native sandbox capability directly in the engine, providing access to web search (`google`), fetching (`curl`), and a secure shell executor (`shell`).
- **Anthropic Message API & OpenAI Compatibility**: Exposes a direct `/anthropic/v1/messages` endpoint so you can drop Lychee into your existing Claude-based applications without modifying code.
- **Embedded Web Dashboard**: Real-time observability UI showing inference requests, system stats, VRAM allocation, and active model execution.

The codebase compiles to a single static binary and maintains 100% backward compatibility with Ollama's Modelfile, CLI commands, and libraries.

We'd love to hear your feedback, bug reports, and suggestions for future optimization!

Github: https://github.com/MD-Mushfiqur123/lychee
