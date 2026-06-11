"""
lychee-python: Official Python SDK for the Lychee local LLM runtime.

Install:
    pip install lychee-python

Usage:
    from lychee import Lychee

    client = Lychee()  # defaults to http://localhost:11434

    # Chat
    response = client.chat("gemma3", "Hello!")

    # Compose pipeline
    result = client.compose(
        input="Translate and summarize: Hello world",
        steps=[
            {"model": "gemma3", "prompt": "Translate to French: {{input}}"},
            {"model": "llama3.2", "prompt": "Summarize: {{step[0].output}}"},
        ]
    )
"""

from __future__ import annotations

import json
from typing import Any, Generator, Iterator
from urllib import request, error as urllib_error


class LycheeError(Exception):
    """Raised when the Lychee server returns an error."""
    pass


class Lychee:
    """
    Python client for the Lychee local LLM runtime.

    Examples:
        >>> client = Lychee()
        >>> response = client.chat("gemma3", "What is 2 + 2?")
        >>> print(response["message"]["content"])

        >>> # Streaming
        >>> for chunk in client.chat("gemma3", "Tell me a story", stream=True):
        ...     print(chunk["message"]["content"], end="", flush=True)

        >>> # Model Composer pipeline
        >>> result = client.compose(
        ...     input="Quantum computing explained",
        ...     steps=[
        ...         {"model": "gemma3", "prompt": "Explain simply: {{input}}"},
        ...         {"model": "llama3.2", "prompt": "Summarize in 1 line: {{step[0].output}}"},
        ...     ]
        ... )
        >>> print(result["output"])
    """

    def __init__(self, base_url: str = "http://localhost:11434"):
        self.base_url = base_url.rstrip("/")

    # ──────────────────────────────────────────────────────────────────────────
    # Internal helpers
    # ──────────────────────────────────────────────────────────────────────────

    def _post(self, path: str, payload: dict[str, Any], stream: bool = False) -> Any:
        url = f"{self.base_url}{path}"
        data = json.dumps(payload).encode("utf-8")
        req = request.Request(
            url,
            data=data,
            headers={"Content-Type": "application/json"},
            method="POST",
        )
        try:
            resp = request.urlopen(req)
        except urllib_error.HTTPError as e:
            body = e.read().decode("utf-8", errors="replace")
            raise LycheeError(f"HTTP {e.code}: {body}") from e

        if stream:
            return self._iter_ndjson(resp)
        else:
            raw = resp.read().decode("utf-8")
            return json.loads(raw)

    def _iter_ndjson(self, resp) -> Generator[dict, None, None]:
        for line in resp:
            line = line.decode("utf-8").strip()
            if line:
                yield json.loads(line)

    def _get(self, path: str) -> Any:
        url = f"{self.base_url}{path}"
        req = request.Request(url, method="GET")
        try:
            resp = request.urlopen(req)
        except urllib_error.HTTPError as e:
            body = e.read().decode("utf-8", errors="replace")
            raise LycheeError(f"HTTP {e.code}: {body}") from e
        return json.loads(resp.read().decode("utf-8"))

    # ──────────────────────────────────────────────────────────────────────────
    # Chat API
    # ──────────────────────────────────────────────────────────────────────────

    def chat(
        self,
        model: str,
        message: str,
        *,
        system: str | None = None,
        history: list[dict] | None = None,
        stream: bool = False,
        options: dict[str, Any] | None = None,
    ) -> dict | Iterator[dict]:
        """
        Send a chat message to a local model.

        Args:
            model: Model name (e.g. "gemma3", "llama3.2")
            message: User message text
            system: Optional system prompt
            history: Previous messages for multi-turn conversation
            stream: If True, returns an iterator of streamed chunks
            options: Model inference options (temperature, top_p, etc.)

        Returns:
            Response dict or iterator of chunks if stream=True
        """
        messages = []
        if system:
            messages.append({"role": "system", "content": system})
        if history:
            messages.extend(history)
        messages.append({"role": "user", "content": message})

        payload: dict[str, Any] = {"model": model, "messages": messages, "stream": stream}
        if options:
            payload["options"] = options

        return self._post("/api/chat", payload, stream=stream)

    # ──────────────────────────────────────────────────────────────────────────
    # Generate API
    # ──────────────────────────────────────────────────────────────────────────

    def generate(
        self,
        model: str,
        prompt: str,
        *,
        stream: bool = False,
        options: dict[str, Any] | None = None,
    ) -> dict | Iterator[dict]:
        """
        Generate text from a local model (single-turn, no chat history).

        Args:
            model: Model name
            prompt: Raw prompt string
            stream: If True, returns an iterator of streamed chunks
            options: Model inference options
        """
        payload: dict[str, Any] = {"model": model, "prompt": prompt, "stream": stream}
        if options:
            payload["options"] = options
        return self._post("/api/generate", payload, stream=stream)

    # ──────────────────────────────────────────────────────────────────────────
    # Model Composer
    # ──────────────────────────────────────────────────────────────────────────

    def compose(
        self,
        input: str,
        steps: list[dict[str, Any]],
        *,
        stream: bool = False,
    ) -> dict | Iterator[dict]:
        """
        Execute a multi-model composition pipeline.

        Each step can reference previous outputs via template variables:
        - ``{{input}}`` — the original input
        - ``{{step[N].output}}`` — output of step N
        - ``{{step[N].parallel[M].output}}`` — parallel branch output

        Args:
            input: Initial input string
            steps: List of step dicts. Each step must have 'model' and 'prompt'.
                   Optional: 'timeout_sec', 'fallback_model', 'parallel', 'options'
            stream: If True, yields SSE events as dicts

        Example::

            result = client.compose(
                input="Analyze this code: def foo(): pass",
                steps=[
                    {
                        "model": "gemma3",
                        "prompt": "Find bugs in: {{input}}",
                        "timeout_sec": 30,
                        "fallback_model": "llama3.2",
                    },
                    {
                        "model": "phi3",
                        "prompt": "Suggest fixes for: {{step[0].output}}",
                    },
                ]
            )
            print(result["output"])
        """
        payload: dict[str, Any] = {"input": input, "steps": steps, "stream": stream}
        return self._post("/api/compose", payload, stream=stream)

    # ──────────────────────────────────────────────────────────────────────────
    # Anthropic-compatible API
    # ──────────────────────────────────────────────────────────────────────────

    def messages(
        self,
        model: str,
        messages: list[dict],
        *,
        max_tokens: int = 1024,
        system: str | None = None,
        stream: bool = False,
    ) -> dict | Iterator[dict]:
        """
        Anthropic Messages API compatible endpoint.

        Drop-in replacement for the official anthropic-sdk-python:

        Example::

            from lychee import Lychee
            client = Lychee()
            response = client.messages("gemma3", [
                {"role": "user", "content": "Hello!"}
            ])
            print(response["content"][0]["text"])
        """
        payload: dict[str, Any] = {
            "model": model,
            "messages": messages,
            "max_tokens": max_tokens,
            "stream": stream,
        }
        if system:
            payload["system"] = system
        return self._post("/v1/messages", payload, stream=stream)

    # ──────────────────────────────────────────────────────────────────────────
    # Model Management
    # ──────────────────────────────────────────────────────────────────────────

    def list_models(self) -> list[dict]:
        """Return a list of all locally available models."""
        resp = self._get("/api/tags")
        return resp.get("models", [])

    def pull(self, model: str, *, stream: bool = True) -> dict | Iterator[dict]:
        """Pull a model from the Lychee registry."""
        return self._post("/api/pull", {"model": model, "stream": stream}, stream=stream)

    def show(self, model: str) -> dict:
        """Show information about a local model."""
        return self._post("/api/show", {"model": model})

    def delete(self, model: str) -> None:
        """Delete a local model."""
        url = f"{self.base_url}/api/delete"
        data = json.dumps({"model": model}).encode("utf-8")
        req = request.Request(
            url, data=data,
            headers={"Content-Type": "application/json"},
            method="DELETE",
        )
        try:
            request.urlopen(req)
        except urllib_error.HTTPError as e:
            raise LycheeError(f"HTTP {e.code}") from e

    # ──────────────────────────────────────────────────────────────────────────
    # Convenience helpers
    # ──────────────────────────────────────────────────────────────────────────

    def is_running(self) -> bool:
        """Check if the Lychee server is running."""
        try:
            self._get("/")
            return True
        except Exception:
            return False

    def __repr__(self) -> str:
        return f"Lychee(base_url={self.base_url!r})"
