# 🍒 lychee-js

> JavaScript/TypeScript SDK for [Lychee](https://github.com/MD-Mushfiqur123/lychee) — the Universal Local LLM Runtime & Orchestration Layer.

[![npm version](https://img.shields.io/npm/v/lychee-js?style=for-the-badge&color=yellow)](https://www.npmjs.com/package/lychee-js)
[![license](https://img.shields.io/badge/license-MIT-blue?style=for-the-badge)](https://github.com/MD-Mushfiqur123/lychee/blob/main/LICENSE)

---

## Installation

```bash
npm install lychee-js
# or
yarn add lychee-js
# or
pnpm add lychee-js
```

> Requires **Node.js >= 18** (uses native `fetch`).

---

## Quick Start

```ts
import { LycheeClient } from "lychee-js";

const lychee = new LycheeClient({ host: "http://localhost:11434" });

// Generate a completion
const answer = await lychee.generate("gemma3", "What is 2 + 2?");
console.log(answer); // "4"

// Chat with history
const reply = await lychee.chat("gemma3", [
  { role: "system", content: "You are a helpful pirate." },
  { role: "user", content: "Where is the treasure?" },
]);

// Stream tokens
for await (const token of lychee.streamChat("gemma3", [
  { role: "user", content: "Tell me a joke." },
])) {
  process.stdout.write(token);
}

// List models
const models = await lychee.listModels();
console.log(models.map((m) => m.name));

// Pull a model from HuggingFace
await lychee.pullModel("microsoft/Phi-3-mini-4k-instruct-gguf");
```

---

## API Reference

### `new LycheeClient(config?)`

| Option | Type | Default | Description |
|:---|:---|:---|:---|
| `host` | `string` | `"http://localhost:11434"` | Lychee server base URL |
| `timeout` | `number` | `120000` | Request timeout in ms |
| `fetch` | `typeof fetch` | `globalThis.fetch` | Custom fetch implementation |

### Methods

#### `generate(model, prompt, options?)` → `Promise<string>`

Single-turn text generation from a plain prompt.

```ts
const text = await lychee.generate("gemma3", "Explain quantum computing in one sentence.");
```

#### `chat(model, messages, options?)` → `Promise<string>`

Multi-turn conversation.

```ts
const reply = await lychee.chat("gemma3", [
  { role: "system", content: "You are a coding assistant." },
  { role: "user", content: "Write a function to reverse a linked list." },
]);
```

#### `streamChat(model, messages)` → `AsyncGenerator<string>`

Stream tokens as they're produced by the model.

```ts
for await (const chunk of lychee.streamChat("gemma3", [
  { role: "user", content: "Write a haiku about programming." },
])) {
  process.stdout.write(chunk);
}
```

#### `listModels()` → `Promise<ModelInfo[]>`

List all downloaded models.

#### `pullModel(name)` → `Promise<void>`

Pull a model from HuggingFace.

```ts
await lychee.pullModel("bartowski/Meta-Llama-3.1-8B-Instruct-GGUF");
```

#### `embeddings(model, input)` → `Promise<number[][]>`

Generate embeddings.

```ts
const vecs = await lychee.embeddings("nomic-embed-text", "Hello, world!");
console.log(vecs[0].length); // e.g. 768
```

### Types

| Interface | Description |
|:---|:---|
| `LycheeConfig` | Client configuration |
| `Message` | `{ role: string; content: string }` |
| `GenerateOptions` | `{ stream?, temperature?, maxTokens?, topP?, stop? }` |
| `ModelInfo` | Model metadata (name, size, digest, family, etc.) |

---

## Environment

Works in any runtime that supports `fetch`:

| Runtime | Status |
|:---|:---:|
| **Node.js >= 18** | ✅ native |
| **Deno** | ✅ native |
| **Bun** | ✅ native |
| **Browser** | ✅ (if CORS configured on server) |
| **Node.js 16–17** | ⚠️ polyfill: `npm install node-fetch`, pass via `config.fetch` |

---

## Related

- [Lychee Server](https://github.com/MD-Mushfiqur123/lychee) — the runtime
- [Lychee Python SDK](https://pypi.org/project/lychee-python/) — `pip install lychee-python`
- [Lychee Desktop](https://github.com/MD-Mushfiqur123/lychee-desktop) — native Windows GUI
- [Lychee Documentation](https://md-mushfiqur123.github.io/lychee-docs/)

---

## License

MIT — see [LICENSE](https://github.com/MD-Mushfiqur123/lychee/blob/main/LICENSE).
