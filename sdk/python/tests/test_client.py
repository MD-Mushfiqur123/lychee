"""Tests for the Lychee Python SDK."""

import json
import unittest
from unittest.mock import MagicMock, patch
from io import BytesIO

from lychee import Lychee, LycheeError


def make_response(data: dict | list, status: int = 200) -> MagicMock:
    """Helper to create a mock HTTP response."""
    body = json.dumps(data).encode("utf-8")
    mock = MagicMock()
    mock.read.return_value = body
    mock.status = status
    mock.__iter__ = lambda self: iter([body])
    return mock


class TestLycheeInit(unittest.TestCase):
    def test_default_url(self):
        client = Lychee()
        self.assertEqual(client.base_url, "http://localhost:11434")

    def test_custom_url(self):
        client = Lychee("http://192.168.1.10:11434")
        self.assertEqual(client.base_url, "http://192.168.1.10:11434")

    def test_trailing_slash_stripped(self):
        client = Lychee("http://localhost:11434/")
        self.assertEqual(client.base_url, "http://localhost:11434")

    def test_repr(self):
        client = Lychee()
        self.assertIn("localhost:11434", repr(client))


class TestLycheeChat(unittest.TestCase):
    def setUp(self):
        self.client = Lychee()

    @patch("urllib.request.urlopen")
    def test_chat_simple(self, mock_urlopen):
        mock_urlopen.return_value = make_response({
            "message": {"role": "assistant", "content": "Hello!"},
            "done": True,
        })
        result = self.client.chat("gemma3", "Hi")
        self.assertEqual(result["message"]["content"], "Hello!")

    @patch("urllib.request.urlopen")
    def test_chat_with_system(self, mock_urlopen):
        mock_urlopen.return_value = make_response({
            "message": {"role": "assistant", "content": "Bonjour!"},
            "done": True,
        })
        result = self.client.chat("gemma3", "Hi", system="Reply in French.")
        self.assertIsNotNone(result)

    @patch("urllib.request.urlopen")
    def test_chat_with_history(self, mock_urlopen):
        mock_urlopen.return_value = make_response({
            "message": {"role": "assistant", "content": "It's 4."},
            "done": True,
        })
        history = [
            {"role": "user", "content": "What is 2+2?"},
            {"role": "assistant", "content": "4"},
        ]
        result = self.client.chat("gemma3", "Are you sure?", history=history)
        self.assertIn("message", result)


class TestLycheeGenerate(unittest.TestCase):
    def setUp(self):
        self.client = Lychee()

    @patch("urllib.request.urlopen")
    def test_generate(self, mock_urlopen):
        mock_urlopen.return_value = make_response({
            "response": "The sky is blue.",
            "done": True,
        })
        result = self.client.generate("gemma3", "Why is the sky blue?")
        self.assertIn("response", result)


class TestLycheeCompose(unittest.TestCase):
    def setUp(self):
        self.client = Lychee()

    @patch("urllib.request.urlopen")
    def test_compose_basic(self, mock_urlopen):
        mock_urlopen.return_value = make_response({
            "output": "Final composed output",
            "results": [
                {"model": "gemma3", "output": "Step 1 output"},
                {"model": "llama3.2", "output": "Final composed output"},
            ],
        })
        result = self.client.compose(
            input="Hello",
            steps=[
                {"model": "gemma3", "prompt": "Translate: {{input}}"},
                {"model": "llama3.2", "prompt": "Summarize: {{step[0].output}}"},
            ],
        )
        self.assertEqual(result["output"], "Final composed output")
        self.assertEqual(len(result["results"]), 2)

    @patch("urllib.request.urlopen")
    def test_compose_with_options(self, mock_urlopen):
        mock_urlopen.return_value = make_response({"output": "ok", "results": []})
        result = self.client.compose(
            input="test",
            steps=[{
                "model": "gemma3",
                "prompt": "{{input}}",
                "timeout_sec": 30,
                "fallback_model": "llama3.2",
            }],
        )
        self.assertIsNotNone(result)


class TestLycheeMessages(unittest.TestCase):
    def setUp(self):
        self.client = Lychee()

    @patch("urllib.request.urlopen")
    def test_messages_basic(self, mock_urlopen):
        mock_urlopen.return_value = make_response({
            "content": [{"type": "text", "text": "Hello there!"}],
            "model": "gemma3",
            "role": "assistant",
        })
        result = self.client.messages(
            "gemma3",
            [{"role": "user", "content": "Hello!"}],
        )
        self.assertEqual(result["content"][0]["text"], "Hello there!")


class TestLycheeModels(unittest.TestCase):
    def setUp(self):
        self.client = Lychee()

    @patch("urllib.request.urlopen")
    def test_list_models(self, mock_urlopen):
        mock_urlopen.return_value = make_response({
            "models": [
                {"name": "gemma3:latest", "size": 1234567890},
                {"name": "llama3.2:latest", "size": 2345678901},
            ]
        })
        models = self.client.list_models()
        self.assertEqual(len(models), 2)
        self.assertEqual(models[0]["name"], "gemma3:latest")

    @patch("urllib.request.urlopen")
    def test_list_models_empty(self, mock_urlopen):
        mock_urlopen.return_value = make_response({"models": []})
        models = self.client.list_models()
        self.assertEqual(models, [])


class TestLycheeErrors(unittest.TestCase):
    def setUp(self):
        self.client = Lychee()

    @patch("urllib.request.urlopen")
    def test_http_error_raises_lychee_error(self, mock_urlopen):
        import urllib.error
        mock_urlopen.side_effect = urllib.error.HTTPError(
            url="http://localhost:11434/api/chat",
            code=500,
            msg="Internal Server Error",
            hdrs=None,
            fp=BytesIO(b'{"error": "model not found"}'),
        )
        with self.assertRaises(LycheeError):
            self.client.chat("nonexistent-model", "Hi")


if __name__ == "__main__":
    unittest.main()
