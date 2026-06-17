# Contributing to Lychee

> [!NOTE]
> Lychee is an open-source fork of Ollama (licensed under MIT). To understand the additions and differences of Lychee compared to the Ollama baseline, please refer to [CHANGELOG.md](./CHANGELOG.md).

Thank you for your interest in contributing to Lychee! Here are a few guidelines to help get you started.

## Set up

See the [development documentation](./docs/development.md) for full instructions on building and running Lychee.

### Building from Source

To build the project:
```bash
go build ./...
```

### Running Tests

To run the unit test suite:
```bash
go test ./cmd/... -v
```

### Ideal issues

* [Bugs](https://github.com/MD-Mushfiqur123/lychee/issues?q=is%3Aissue+is%3Aopen+label%3Abug): issues where Lychee stops working or where it results in an unexpected error.
* [Performance](https://github.com/MD-Mushfiqur123/lychee/issues?q=is%3Aissue+is%3Aopen+label%3Aperformance): issues to make Lychee faster at model inference, downloading or uploading.
* [Security](https://github.com/MD-Mushfiqur123/lychee/blob/main/SECURITY.md): issues that could lead to a security vulnerability. As mentioned in [SECURITY.md](https://github.com/MD-Mushfiqur123/lychee/blob/main/SECURITY.md), please do not disclose security vulnerabilities publicly.

### Issues that are harder to review

* New features: new features (e.g. API fields, environment variables) add surface area to Lychee and make it harder to maintain in the long run as they cannot be removed without potentially breaking users in the future.
* Refactoring: large code improvements are important, but can be harder or take longer to review and merge.
* Documentation: small updates to fill in or correct missing documentation are helpful, however large documentation additions can be hard to maintain over time.

### Issues that may not be accepted

* Changes that break backwards compatibility in Lychee's API (including the OpenAI-compatible API)
* Changes that add significant friction to the user experience
* Changes that create a large future maintenance burden for maintainers and contributors

## Proposing a (non-trivial) change

> By "non-trivial", we mean a change that is not a bug fix or small
> documentation update. If you are unsure, please open an issue to discuss it.

Before opening a non-trivial Pull Request, please open an issue to discuss the change and
get feedback from the maintainers. This helps us understand the context of the
change and how it fits into Lychee's roadmap and prevents us from duplicating
work or you from spending time on a change that we may not be able to accept.

Tips for proposals:

* Explain the problem you are trying to solve, not what you are trying to do.
* Explain why the change is important.
* Explain how the change will be used.
* Explain how the change will be tested.

Additionally, for bonus points: Provide draft documentation you would expect to
see if the changes were accepted.

## Pull requests

**Commit messages**

The title should look like:

    <package>: <short description>

The package is the most affected Go package. If the change does not affect Go
code, then use the directory name instead. Changes to a single well-known
file in the root directory may use the file name.

The short description should start with a lowercase letter and be a
continuation of the sentence:

      "This changes Lychee to..."

Examples:

      llm/backend/mlx: support the llama architecture
      CONTRIBUTING: provide clarity on good commit messages, and bad

Bad Examples:

      feat: add more emoji
      fix: was not using famous web framework
      chore: generify code

**Tests**

Please include tests. Strive to test behavior, not implementation.

**New dependencies**

Dependencies should be added sparingly. If you are adding a new dependency,
please explain why it is necessary and what other ways you attempted that
did not work without it.

## Differentiators & Roadmap Focus

When contributing new features or improvements to Lychee, keep in mind our core pillars of differentiation compared to standard runners:
1. **Universal API Gateway**: Native translation layers for Anthropic Messages, OpenAI Completions, and OpenAI Responses.
2. **Native Agent Mode**: Stateful agent orchestration within the sandbox environment (`x/agent`).
3. **Prompt Caching**: Intelligent KV cache reuse for low TTFT.
4. **Model Composer**: Sequential execution chaining with template variables for multi-step prompts.
5. **Conversation Memory**: Persistent session tracking using JSON and SQLite with search, pagination, and auto-titling APIs.
6. **Model Router**: Load balancing (round-robin, random, least-loaded, weighted) across multiple instances with circuit breakers and health checks.
7. **Structured Output**: JSON schema-conforming validation and error-correction retry logic with streaming SSE updates and temperature escalation.
8. **Multi-Language SDKs**: Official integration client libraries (Python, JavaScript, Rust), IDE integrations (VSCode), and advanced CLI pipelines.
9. **Embedded Web Dashboard**: Embedded React-based SPA console served at `/dashboard/` for interactive chat and configuration.
10. **Model Comparison & Benchmarking**: CLI tools for benchmarking (`compare`, `scan`, `generate-client`).
11. **Rich Terminal Dashboard**: Live terminal-based performance monitoring (`stats --tui`).

Features aligning with these pillars are highly encouraged.

## Translations

We welcome community translations of the Lychee README and documentation! If you'd like to help make Lychee accessible in your language:

1. **Read the English README** ([README.md](./README.md)) to understand the full project scope.
2. **Create a new file** named `README.xx.md` (e.g., `README.fr.md` for French, `README.de.md` for German) in the repository root.
3. **Translate the content** — keep the structure, code examples, and links intact. You may adapt phrasing for natural readability in your language.
4. **Add your language** to the [🌍 Translations](./README.md#-translations) section of the main README.
5. **Open a pull request** — mark it with the `documentation` and `translation` labels.

### Guidelines

- Preserve all code blocks, links, and badge URLs exactly as they are.
- Translate descriptive text and headings, but keep technical terms (API names, commands, file paths) in English.
- If a language already has a translation, you are welcome to review and improve it.
- All translations are maintained by the community — if you notice an outdated translation, please help update it!

Questions? Open a [discussion](https://github.com/MD-Mushfiqur123/lychee/discussions) or reach out in our community channels.

---

## Need help?

If you need help with anything, please open an issue or start a discussion on our GitHub repository.
