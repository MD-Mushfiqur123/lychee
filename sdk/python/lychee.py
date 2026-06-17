"""
lychee — Python client for Lychee, the universal LLM runtime.

Minimal single-file SDK that talks to a local Lychee server at http://localhost:11434.

Examples:

    from lychee import LycheeClient

    client = LycheeClient()

    # Chat
    resp = client.chat("gemma3", [{"role": "user", "content": "Hello!"}])
    print(resp["message"]["content"])

    # Streaming generate
    for chunk in client.generate("gemma3", "Write a poem.", stream=True):
        print(chunk.get("response", ""), end="", flush=True)

    # List installed models
    for m in client.list_models():
        print(m["name"])

    # Pull a model from HuggingFace
    for progress in client.pull_model("bartowski/Meta-Llama-3.1-8B-Instruct-GGUF"):
        print(progress.get("status"), progress.get("completed", 0), progress.get("total", 0))

    # Get embeddings
    vec = client.embeddings("nomic-embed-text", "Hello world")
    print(len(vec))  # e.g. 768
"""

from __future__ import annotations

from typing import Any, Dict, Generator, Iterator, List, Optional, Union

import requests


__all__ = ["LycheeClient", "LycheeError"]
__version__ = "0.2.0"


class LycheeError(Exception):
    """Raised when the Lychee server returns an error or is unreachable."""


class LycheeClient:
    """
    Minimal Python client for the Lychee local LLM runtime.

    Parameters:
        host: Base URL of the Lychee server (default ``http://localhost:11434``).
        timeout: Request timeout in seconds.
    """

    def __init__(self, host: str = "http://localhost:11434", timeout: float = 120.0):
        self.host = host.rstrip("/")
        self.timeout = timeout
        self._session = requests.Session()
        self._session.headers.update({"Content-Type": "application/json"})

    # ── helpers ──────────────────────────────────────────────────────────────

    def _post(
        self, path: str, payload: Dict[str, Any], stream: bool = False
    ) -> Any:
        url = f"{self.host}{path}"
        resp = self._session.post(url, json=payload, stream=stream, timeout=self.timeout)
        self._raise_for_status(resp)

        if stream:
            return self._iter_ndjson(resp)
        return resp.json()

    def _get(self, path: str) -> Any:
        url = f"{self.host}{path}"
        resp = self._session.get(url, timeout=self.timeout)
        self._raise_for_status(resp)
        return resp.json()

    @staticmethod
    def _raise_for_status(resp: requests.Response) -> None:
        if not resp.ok:
            try:
                body = resp.text[:500]
            except Exception:
                body = "<unreadable>"
            raise LycheeError(f"HTTP {resp.status_code}: {body}")

    @staticmethod
    def _iter_ndjson(resp: requests.Response) -> Generator[Dict, None, None]:
        """Yield JSON objects from newline-delimited JSON (streaming)."""
        for line in resp.iter_lines(decode_unicode=True):
            if line:
                import json
                yield json.loads(line)

    # ── generate ─────────────────────────────────────────────────────────────

    def generate(
        self, model: str, prompt: str, *, stream: bool = False, **options
    ) -> Union[Dict, Generator[Dict, None, None]]:
        """
        Generate text from a model (single-turn, no chat history).

        Args:
            model: Model name, e.g. ``"gemma3"``, ``"llama3.2"``.
            prompt: The raw prompt string.
            stream: If *True*, returns a generator yielding chunks line-by-line.
            **options: Additional inference options forwarded as ``"options"``.

        Returns:
            Response dict, or a generator of chunk dicts when stream=True.
        """
        payload: Dict[str, Any] = {"model": model, "prompt": prompt, "stream": stream}
        if options:
            payload["options"] = options
        return self._post("/api/generate", payload, stream=stream)

    # ── chat ─────────────────────────────────────────────────────────────────

    def chat(
        self,
        model: str,
        messages: List[Dict[str, Any]],
        *,
        stream: bool = False,
        **options,
    ) -> Union[Dict, Generator[Dict, None, None]]:
        """
        Chat with a model using a message list.

        Args:
            model: Model name.
            messages: List of message dicts, each with ``"role"`` and ``"content"``.
            stream: If *True*, returns a generator yielding chunks line-by-line.
            **options: Additional inference options.

        Returns:
            Chat response dict, or a generator of chunk dicts when stream=True.

        Example::

            resp = client.chat("gemma3", [
                {"role": "system", "content": "You are a helpful assistant."},
                {"role": "user", "content": "What is the capital of France?"},
            ])
            print(resp["message"]["content"])
        """
        payload: Dict[str, Any] = {"model": model, "messages": messages, "stream": stream}
        if options:
            payload["options"] = options
        return self._post("/api/chat", payload, stream=stream)

    # ── list_models ──────────────────────────────────────────────────────────

    def list_models(self) -> List[Dict]:
        """
        List all locally installed models.

        Returns:
            A list of model dicts, each containing ``name``, ``size``, ``modified_at``, etc.
        """
        resp = self._get("/api/tags")
        return resp.get("models", [])

    # ── pull_model ───────────────────────────────────────────────────────────

    def pull_model(self, name: str) -> Generator[Dict, None, None]:
        """
        Pull a model from the HuggingFace registry.

        Yields progress dicts with keys like ``"status"``, ``"completed"``, ``"total"``.

        Args:
            name: Model identifier, e.g. ``"bartowski/Meta-Llama-3.1-8B-Instruct-GGUF"``.
        """
        yield from self._post("/api/pull", {"model": name, "stream": True}, stream=True)

    # ── embeddings ───────────────────────────────────────────────────────────

    def embeddings(self, model: str, input: str) -> List[float]:
        """
        Get embeddings for a string.

        Args:
            model: Embedding model name, e.g. ``"nomic-embed-text"``.
            input: The input text to embed.

        Returns:
            A list of floats representing the embedding vector.
            When the server returns multiple inputs, only the first vector is returned.
        """
        resp = self._post("/api/embed", {"model": model, "input": input})
        embeds: List[List[float]] = resp.get("embeddings", [])
        if not embeds:
            raise LycheeError("No embeddings returned from server")
        return embeds[0]

    # ── embeddings_batch ─────────────────────────────────────────────────────

    def embeddings_batch(self, model: str, inputs: List[str]) -> List[List[float]]:
        """
        Get embeddings for multiple input strings in one request.

        Args:
            model: Embedding model name.
            inputs: List of input strings.

        Returns:
            A list of embedding vectors (one per input string).
        """
        resp = self._post("/api/embed", {"model": model, "input": inputs})
        return resp.get("embeddings", [])

    # ── model info / management ──────────────────────────────────────────────

    def show_model(self, model: str) -> Dict:
        """Return model details: license, template, parameters, etc."""
        return self._post("/api/show", {"model": model})

    def delete_model(self, model: str) -> None:
        """Delete a local model by name."""
        url = f"{self.host}/api/delete"
        resp = self._session.delete(
            url, json={"model": model}, timeout=self.timeout
        )
        self._raise_for_status(resp)

    def list_running(self) -> List[Dict]:
        """List models currently loaded in memory."""
        return self._get("/api/ps").get("models", [])

    # ── compose (multi-model DAG) ────────────────────────────────────────────

    def compose(
        self,
        input: str,
        steps: List[Dict[str, Any]],
        *,
        stream: bool = False,
    ) -> Union[Dict, Generator[Dict, None, None]]:
        """
        Execute a multi-model composition pipeline (DAG).

        Args:
            input: Initial input string.
            steps: List of step dicts. Each step must have ``"model"`` and ``"prompt"``.
            stream: If *True*, yields SSE event dicts.

        Returns:
            Composition result dict where ``result["output"]`` is the final output.
        """
        return self._post(
            "/api/compose",
            {"input": input, "steps": steps, "stream": stream},
            stream=stream,
        )

    # ── structured output ────────────────────────────────────────────────────

    def structured(
        self,
        model: str,
        prompt: str,
        schema: Union[Dict, List, str],
        *,
        max_retries: int = 3,
        **options,
    ) -> Dict:
        """
        Generate JSON conforming to a schema with automatic retry on failure.

        Args:
            model: Model name.
            prompt: The generation prompt.
            schema: JSON Schema dict, list, or string.
            max_retries: Maximum validation retries.
            **options: Additional inference options.
        """
        payload: Dict[str, Any] = {
            "model": model,
            "prompt": prompt,
            "schema": schema,
            "max_retries": max_retries,
        }
        if options:
            payload["options"] = options
        return self._post("/api/structured", payload)

    # ── convenience ──────────────────────────────────────────────────────────

    def version(self) -> Dict:
        """Return the server version info."""
        return self._get("/api/version")

    def health(self) -> bool:
        """Return *True* if the server is reachable."""
        try:
            self._get("/")
            return True
        except Exception:
            return False

    def close(self) -> None:
        """Close the underlying HTTP session."""
        self._session.close()

    def __repr__(self) -> str:
        return f"LycheeClient(host={self.host!r})"

    def __enter__(self) -> "LycheeClient":
        return self

    def __exit__(self, *args: Any) -> None:
        self.close()
