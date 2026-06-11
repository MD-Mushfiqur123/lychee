/**
 * lychee-js: Official JavaScript/TypeScript SDK for the Lychee local LLM runtime.
 *
 * @example
 * ```js
 * import { Lychee } from 'lychee-js';
 *
 * const client = new Lychee(); // defaults to http://localhost:11434
 *
 * // Chat
 * const response = await client.chat('gemma3', 'Hello!');
 * console.log(response.message.content);
 *
 * // Streaming
 * for await (const chunk of client.chat('gemma3', 'Tell me a story', { stream: true })) {
 *   process.stdout.write(chunk.message?.content ?? '');
 * }
 *
 * // Model Composer pipeline
 * const result = await client.compose({
 *   input: 'Explain quantum computing',
 *   steps: [
 *     { model: 'gemma3', prompt: 'Explain simply: {{input}}' },
 *     { model: 'llama3.2', prompt: 'Summarize in 1 line: {{step[0].output}}' },
 *   ],
 * });
 * console.log(result.output);
 * ```
 */

'use strict';

class LycheeError extends Error {
  constructor(message, statusCode) {
    super(message);
    this.name = 'LycheeError';
    this.statusCode = statusCode;
  }
}

class Lychee {
  /**
   * @param {string} [baseUrl='http://localhost:11434'] - Lychee server URL
   */
  constructor(baseUrl = 'http://localhost:11434') {
    this.baseUrl = baseUrl.replace(/\/$/, '');
  }

  // ────────────────────────────────────────────────────────────────────────────
  // Internal helpers
  // ────────────────────────────────────────────────────────────────────────────

  async _post(path, payload, { stream = false } = {}) {
    const url = `${this.baseUrl}${path}`;
    const response = await fetch(url, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(payload),
    });

    if (!response.ok) {
      const body = await response.text();
      throw new LycheeError(`HTTP ${response.status}: ${body}`, response.status);
    }

    if (stream) {
      return this._readNDJSONStream(response.body);
    }

    return response.json();
  }

  async _get(path) {
    const url = `${this.baseUrl}${path}`;
    const response = await fetch(url);
    if (!response.ok) {
      throw new LycheeError(`HTTP ${response.status}`, response.status);
    }
    return response.json();
  }

  async *_readNDJSONStream(readableStream) {
    const reader = readableStream.getReader();
    const decoder = new TextDecoder();
    let buffer = '';

    try {
      while (true) {
        const { done, value } = await reader.read();
        if (done) break;
        buffer += decoder.decode(value, { stream: true });
        const lines = buffer.split('\n');
        buffer = lines.pop() ?? '';
        for (const line of lines) {
          const trimmed = line.trim();
          if (trimmed) {
            try {
              yield JSON.parse(trimmed);
            } catch {
              // skip malformed lines
            }
          }
        }
      }
      if (buffer.trim()) {
        try { yield JSON.parse(buffer.trim()); } catch { /* skip */ }
      }
    } finally {
      reader.releaseLock();
    }
  }

  // ────────────────────────────────────────────────────────────────────────────
  // Chat API
  // ────────────────────────────────────────────────────────────────────────────

  /**
   * Send a chat message to a local model.
   *
   * @param {string} model - Model name (e.g. "gemma3")
   * @param {string} message - User message
   * @param {object} [options]
   * @param {string} [options.system] - System prompt
   * @param {Array} [options.history] - Previous messages
   * @param {boolean} [options.stream=false] - Enable streaming
   * @param {object} [options.inferenceOptions] - Model options (temperature etc.)
   * @returns {Promise<object> | AsyncGenerator<object>}
   */
  async chat(model, message, { system, history = [], stream = false, inferenceOptions } = {}) {
    const messages = [];
    if (system) messages.push({ role: 'system', content: system });
    messages.push(...history);
    messages.push({ role: 'user', content: message });

    const payload = { model, messages, stream };
    if (inferenceOptions) payload.options = inferenceOptions;

    return this._post('/api/chat', payload, { stream });
  }

  // ────────────────────────────────────────────────────────────────────────────
  // Generate API
  // ────────────────────────────────────────────────────────────────────────────

  /**
   * Generate text from a local model (single-turn, no history).
   *
   * @param {string} model
   * @param {string} prompt
   * @param {object} [options]
   */
  async generate(model, prompt, { stream = false, inferenceOptions } = {}) {
    const payload = { model, prompt, stream };
    if (inferenceOptions) payload.options = inferenceOptions;
    return this._post('/api/generate', payload, { stream });
  }

  // ────────────────────────────────────────────────────────────────────────────
  // Model Composer
  // ────────────────────────────────────────────────────────────────────────────

  /**
   * Execute a multi-model composition pipeline.
   *
   * @param {object} params
   * @param {string} params.input - Initial input text
   * @param {Array<{model: string, prompt: string, timeout_sec?: number, fallback_model?: string, parallel?: Array, options?: object}>} params.steps
   * @param {boolean} [params.stream=false]
   *
   * @example
   * const result = await client.compose({
   *   input: 'Review this code: function add(a,b){return a+b}',
   *   steps: [
   *     { model: 'gemma3', prompt: 'Find issues in: {{input}}', timeout_sec: 30 },
   *     { model: 'phi3', prompt: 'Suggest improvements for: {{step[0].output}}' },
   *   ],
   * });
   * console.log(result.output);
   */
  async compose({ input, steps, stream = false }) {
    return this._post('/api/compose', { input, steps, stream }, { stream });
  }

  // ────────────────────────────────────────────────────────────────────────────
  // Anthropic-compatible Messages API
  // ────────────────────────────────────────────────────────────────────────────

  /**
   * Anthropic Messages API compatible endpoint.
   *
   * @param {string} model
   * @param {Array<{role: string, content: string}>} messages
   * @param {object} [options]
   * @param {number} [options.maxTokens=1024]
   * @param {string} [options.system]
   * @param {boolean} [options.stream=false]
   */
  async messages(model, messages, { maxTokens = 1024, system, stream = false } = {}) {
    const payload = { model, messages, max_tokens: maxTokens, stream };
    if (system) payload.system = system;
    return this._post('/v1/messages', payload, { stream });
  }

  // ────────────────────────────────────────────────────────────────────────────
  // OpenAI-compatible API
  // ────────────────────────────────────────────────────────────────────────────

  /**
   * OpenAI Chat Completions compatible endpoint.
   *
   * @param {string} model
   * @param {Array} messages
   * @param {object} [options]
   */
  async chatCompletions(model, messages, { stream = false, ...rest } = {}) {
    return this._post('/v1/chat/completions', { model, messages, stream, ...rest }, { stream });
  }

  // ────────────────────────────────────────────────────────────────────────────
  // Model Management
  // ────────────────────────────────────────────────────────────────────────────

  /** List all locally available models. */
  async listModels() {
    const resp = await this._get('/api/tags');
    return resp.models ?? [];
  }

  /**
   * Pull a model from the registry.
   * @param {string} model
   * @param {object} [options]
   * @param {boolean} [options.stream=true]
   */
  async pull(model, { stream = true } = {}) {
    return this._post('/api/pull', { model, stream }, { stream });
  }

  /** Show information about a model. */
  async show(model) {
    return this._post('/api/show', { model });
  }

  /** Check if the Lychee server is running. */
  async isRunning() {
    try {
      await this._get('/');
      return true;
    } catch {
      return false;
    }
  }
}

module.exports = { Lychee, LycheeError };
