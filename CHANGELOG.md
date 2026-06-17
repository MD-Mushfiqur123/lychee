# Changelog

All notable changes to the Lychee project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [0.2.0-alpha] - 2026-06-17

### Added

- `--run` flag: pull + create + chat in one command
- Interactive REPL mode with `/model`, `/system`, `/clear` commands
- Built-in web dashboard at `localhost:11434`
- Shell completions (bash, zsh, fish, powershell)
- `lychee update` self-update command
- `lychee benchmark` model comparison
- One-liner install scripts (`curl | sh`, `irm | iex`)
- Docker multi-arch builds (amd64 + arm64)
- Desktop auto-start integration
- Pipeline export/import/share
- NSIS Windows installer
- Python SDK, JS/TS SDK, Ruby SDK
- Discord community bot
- GitHub issue/PR templates

### Changed

- Universal pull: auto-detect HuggingFace models (org/model format)
- `--list` and `--quant` flags for HF pulls
- Better error messages, 6 bug fixes

### Fixed

- Critical: Printf format mismatch in `cmd_generate` (runtime panic)
- Multiple unchecked errors and unused variables

## [0.1.1-alpha] - 2026-06-16

### Added

- Initial release with 24 CLI commands
- Anthropic API + OpenAI API compatibility
- Multi-model pipeline engine
- Structured output support
- Persistent memory
- 50 HF model catalog

[0.2.0-alpha]: https://github.com/lychee-org/lychee/compare/v0.1.1-alpha...v0.2.0-alpha
[0.1.1-alpha]: https://github.com/lychee-org/lychee/releases/tag/v0.1.1-alpha
