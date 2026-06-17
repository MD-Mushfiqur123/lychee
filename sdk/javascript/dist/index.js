"use strict";
/**
 * Lychee JavaScript/TypeScript SDK
 *
 * Universal LLM Runtime client — talk to any local model through Lychee Server.
 *
 * @example
 * ```ts
 * import { LycheeClient } from "lychee-js";
 *
 * const lychee = new LycheeClient({ host: "http://localhost:11434" });
 *
 * // Simple generate
 * const answer = await lychee.generate("gemma3", "What is the meaning of life?");
 * console.log(answer);
 *
 * // Chat with messages
 * const reply = await lychee.chat("gemma3", [
 *   { role: "system", content: "You are a helpful pirate." },
 *   { role: "user", content: "Where is the treasure?" },
 * ]);
 * console.log(reply);
 *
 * // Streaming
 * for await (const chunk of lychee.streamChat("gemma3", [
 *   { role: "user", content: "Tell me a short story." }
 * ])) {
 *   process.stdout.write(chunk);
 * }
 * ```
 *
 * @packageDocumentation
 */
Object.defineProperty(exports, "__esModule", { value: true });
exports.LycheeClient = void 0;
// ---------------------------------------------------------------------------
// Client
// ---------------------------------------------------------------------------
/**
 * JavaScript/TypeScript client for Lychee Server.
 *
 * Talks to Lychee's OpenAI-compatible `/v1` endpoints.
 */
class LycheeClient {
    host;
    timeout;
    fetchImpl;
    constructor(config) {
        this.host = config?.host ?? "http://localhost:11434";
        this.timeout = config?.timeout ?? 120_000;
        this.fetchImpl = config?.fetch ?? globalThis.fetch;
    }
    // -- Public API -----------------------------------------------------------
    /**
     * Generate a completion from a plain-text prompt.
     *
     * @param model  - Model name (e.g. `"gemma3"`, `"phi3"`)
     * @param prompt - The input prompt text
     * @param options - Optional generation parameters
     * @returns The generated text
     */
    async generate(model, prompt, options) {
        const res = await this.postJSON("/v1/chat/completions", {
            model,
            messages: [{ role: "user", content: prompt }],
            temperature: options?.temperature,
            top_p: options?.topP,
            max_tokens: options?.maxTokens,
            stop: options?.stop,
            stream: false,
        }, { signal: this.abortSignal() });
        return res.choices[0]?.message?.content ?? "";
    }
    /**
     * Send a full conversation and get the assistant's reply.
     *
     * @param model    - Model name
     * @param messages - Conversation history
     * @param options  - Optional generation parameters
     * @returns The assistant's reply text
     */
    async chat(model, messages, options) {
        const res = await this.postJSON("/v1/chat/completions", {
            model,
            messages,
            temperature: options?.temperature,
            top_p: options?.topP,
            max_tokens: options?.maxTokens,
            stop: options?.stop,
            stream: false,
        }, { signal: this.abortSignal() });
        return res.choices[0]?.message?.content ?? "";
    }
    /**
     * Stream a chat conversation token-by-token.
     *
     * @param model    - Model name
     * @param messages - Conversation history
     * @returns An async generator yielding content deltas as they arrive.
     *
     * @example
     * ```ts
     * for await (const token of lychee.streamChat("gemma3", msgs)) {
     *   process.stdout.write(token);
     * }
     * ```
     */
    async *streamChat(model, messages) {
        const response = await this.fetchImpl(`${this.host}/v1/chat/completions`, {
            method: "POST",
            headers: { "Content-Type": "application/json" },
            body: JSON.stringify({ model, messages, stream: true }),
            signal: this.abortSignal(),
        });
        if (!response.ok) {
            const text = await response.text();
            throw new Error(`Lychee streamChat failed (${response.status}): ${text}`);
        }
        const reader = response.body?.getReader();
        if (!reader)
            throw new Error("No readable stream in response");
        const decoder = new TextDecoder();
        let buffer = "";
        try {
            while (true) {
                const { done, value } = await reader.read();
                if (done)
                    break;
                buffer += decoder.decode(value, { stream: true });
                // Process complete SSE lines
                const lines = buffer.split("\n");
                buffer = lines.pop() ?? ""; // keep incomplete line in buffer
                for (const line of lines) {
                    const trimmed = line.trim();
                    if (!trimmed.startsWith("data: "))
                        continue;
                    const dataStr = trimmed.slice(6);
                    if (dataStr === "[DONE]")
                        return;
                    try {
                        const chunk = JSON.parse(dataStr);
                        const content = chunk.choices[0]?.delta?.content;
                        if (content)
                            yield content;
                    }
                    catch {
                        // Skip malformed JSON lines
                    }
                }
            }
        }
        finally {
            reader.releaseLock();
        }
    }
    /**
     * List all models available on the Lychee server.
     */
    async listModels() {
        const res = await this.getJSON("/api/tags");
        return res.models ?? [];
    }
    /**
     * Pull a model from HuggingFace (or any source Lychee supports).
     *
     * @param name - Model identifier (e.g. `"bartowski/Meta-Llama-3.1-8B-Instruct-GGUF"`)
     *
     * @example
     * ```ts
     * await lychee.pullModel("microsoft/Phi-3-mini-4k-instruct-gguf");
     * ```
     */
    async pullModel(name) {
        await this.postJSON("/api/pull", { name });
    }
    /**
     * Generate embeddings for one or more inputs.
     *
     * @param model - Embedding model name (e.g. `"nomic-embed-text"`)
     * @param input - Single string or array of strings to embed
     * @returns Array of embedding vectors (each a number array)
     */
    async embeddings(model, input) {
        const res = await this.postJSON("/api/embeddings", {
            model,
            input: Array.isArray(input) ? input : [input],
        }, { signal: this.abortSignal() });
        return res.data.map((d) => d.embedding);
    }
    // -- Internal helpers -----------------------------------------------------
    async postJSON(path, body, init) {
        const url = `${this.host}${path}`;
        const response = await this.fetchImpl(url, {
            method: "POST",
            headers: { "Content-Type": "application/json" },
            body: JSON.stringify(body),
            signal: init?.signal,
        });
        if (!response.ok) {
            const text = await response.text();
            throw new Error(`Lychee API error ${response.status} (${path}): ${text}`);
        }
        return (await response.json());
    }
    async getJSON(path) {
        const url = `${this.host}${path}`;
        const response = await this.fetchImpl(url);
        if (!response.ok) {
            const text = await response.text();
            throw new Error(`Lychee API error ${response.status} (${path}): ${text}`);
        }
        return (await response.json());
    }
    abortSignal() {
        return AbortSignal.timeout(this.timeout);
    }
}
exports.LycheeClient = LycheeClient;
//# sourceMappingURL=index.js.map