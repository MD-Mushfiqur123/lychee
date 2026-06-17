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
/** Configuration for connecting to a Lychee server. */
export interface LycheeConfig {
    /**
     * Base URL of the Lychee server.
     * @default "http://localhost:11434"
     */
    host?: string;
    /**
     * Request timeout in milliseconds.
     * @default 120_000 (2 minutes)
     */
    timeout?: number;
    /**
     * Optional fetch implementation (useful for Node <18, Deno, or custom agents).
     */
    fetch?: typeof fetch;
}
/** A chat message in OpenAI-compatible format. */
export interface Message {
    /** Role: "system", "user", "assistant", or "tool" */
    role: string;
    /** Message content (text) */
    content: string;
}
/** Options for `generate()` and `chat()`. */
export interface GenerateOptions {
    /** Stream tokens as they are produced. */
    stream?: boolean;
    /** Sampling temperature (0–2). Lower = more deterministic. */
    temperature?: number;
    /** Maximum number of tokens to generate. */
    maxTokens?: number;
    /** Nucleus sampling probability. */
    topP?: number;
    /** Stop sequences — generation stops when any is encountered. */
    stop?: string[];
}
/** Describes a model available on the server. */
export interface ModelInfo {
    name: string;
    modified_at: string;
    size: number;
    digest: string;
    details: {
        parent_model: string;
        format: string;
        family: string;
        families: string[];
        parameter_size: string;
        quantization_level: string;
    };
}
/**
 * JavaScript/TypeScript client for Lychee Server.
 *
 * Talks to Lychee's OpenAI-compatible `/v1` endpoints.
 */
export declare class LycheeClient {
    private host;
    private timeout;
    private fetchImpl;
    constructor(config?: LycheeConfig);
    /**
     * Generate a completion from a plain-text prompt.
     *
     * @param model  - Model name (e.g. `"gemma3"`, `"phi3"`)
     * @param prompt - The input prompt text
     * @param options - Optional generation parameters
     * @returns The generated text
     */
    generate(model: string, prompt: string, options?: GenerateOptions): Promise<string>;
    /**
     * Send a full conversation and get the assistant's reply.
     *
     * @param model    - Model name
     * @param messages - Conversation history
     * @param options  - Optional generation parameters
     * @returns The assistant's reply text
     */
    chat(model: string, messages: Message[], options?: GenerateOptions): Promise<string>;
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
    streamChat(model: string, messages: Message[]): AsyncGenerator<string, void, undefined>;
    /**
     * List all models available on the Lychee server.
     */
    listModels(): Promise<ModelInfo[]>;
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
    pullModel(name: string): Promise<void>;
    /**
     * Generate embeddings for one or more inputs.
     *
     * @param model - Embedding model name (e.g. `"nomic-embed-text"`)
     * @param input - Single string or array of strings to embed
     * @returns Array of embedding vectors (each a number array)
     */
    embeddings(model: string, input: string | string[]): Promise<number[][]>;
    private postJSON;
    private getJSON;
    private abortSignal;
}
//# sourceMappingURL=index.d.ts.map