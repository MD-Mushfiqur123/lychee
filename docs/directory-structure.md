# `~/.lychee/` Directory Structure

This document describes every directory and file under `~/.lychee/` — the data directory for the Lychee chat server. Understanding this layout helps with backups, disk management, and debugging.

## Overview

```
~/.lychee/
├── config.yaml          # User configuration
├── models/              # Downloaded AI models
│   ├── blobs/           # GGUF model files (content-addressed)
│   └── manifests/       # Model metadata & references
├── logs/                # Server runtime logs
├── history              # Chat conversation history
└── huggingface/         # Hugging Face Hub cache
    └── hub/             # Downloaded HF models & datasets
```

---

## `config.yaml`

**Purpose:** The single source of truth for all Lychee server configuration. Created on first run if absent.

**Typical contents:**
- `server.host` / `server.port` — bind address and port
- `model` — default model to load on startup
- `models` — list of models with their paths, quantization settings, context sizes
- `system_prompt` — default system prompt for new conversations
- `api_keys` — optional API keys for external services (if any)

**Cleanup:** Don't delete this file unless you want a factory reset. Back it up before upgrading or migrating.

**Symlink tip:**
```bash
# Keep config in a synced/version-controlled location
mv ~/.lychee/config.yaml ~/dotfiles/lychee/config.yaml
ln -s ~/dotfiles/lychee/config.yaml ~/.lychee/config.yaml
```

---

## `models/`

**Purpose:** Stores all AI models that Lychee can load for inference. This is typically the largest directory by far.

```
models/
├── blobs/
│   ├── sha256-abc123...   # Model weights (GGUF format)
│   └── sha256-def456...
└── manifests/
    └── TheBloke/
        └── Llama-3-8B-Instruct-GGUF/
            └── latest     # JSON manifest referencing blobs
```

### `models/blobs/`

**Purpose:** The actual model weight files in GGUF format, stored by their SHA256 content hash. This is a **content-addressed store** — identical files deduplicate naturally.

| Detail | Value |
|--------|-------|
| Format | GGUF (`.gguf` extension) |
| Naming | `sha256-<64-char-hex>` |
| Typical size | 4–70 GB per file, depending on model |

**Cleanup:**
- Delete individual blobs that are no longer referenced by any manifest.
- To reclaim space: remove the model via `lychee model rm <name>`, which removes the manifest reference and optionally the blob.
- Safe manual cleanup: run `lychee model list` first, then remove blobs for models no longer listed.

**Symlink tip:**
```bash
# Store model blobs on a large external drive to save SSD space
mv ~/.lychee/models/blobs /mnt/bigdisk/lychee-blobs
ln -s /mnt/bigdisk/lychee-blobs ~/.lychee/models/blobs
```

### `models/manifests/`

**Purpose:** JSON metadata files that describe a model — which blob(s) it uses, parameter sizes, chat templates, tokenizer info, etc. The structure mirrors Hugging Face's `org/repo` hierarchy.

**Typical structure:**
```json
{
  "name": "Llama-3-8B-Instruct-Q4_K_M",
  "blobs": ["sha256-abc123..."],
  "parameters": "8B",
  "quantization": "Q4_K_M",
  "chat_template": "...",
  "tokenizer": { ... }
}
```

**Cleanup:**
- Removing a manifest entry effectively "uninstalls" a model from Lychee's knowledge.
- The underlying blob in `blobs/` is **not** auto-deleted unless you use the CLI tool to uninstall.

**Symlink tip:**
```bash
# Share manifests across machines (useful for homelab clusters)
mv ~/.lychee/models/manifests /shared/nfs/lychee-manifests
ln -s /shared/nfs/lychee-manifests ~/.lychee/models/manifests
```

---

## `logs/`

**Purpose:** Runtime logs from the Lychee server — request handling, model loading, errors, and performance metrics.

**Typical files:**
```
logs/
├── server.log        # Main server log (rotated)
├── server.log.1      # Older rotated log
├── error.log         # Errors and warnings
└── access.log        # API request access log (if enabled)
```

**Log levels** are controlled via `config.yaml` → `logging.level` (debug, info, warn, error).

**Cleanup:**
- Safe to delete older rotated logs (`.log.1`, `.log.2`, etc.).
- Active log (`server.log`) can be truncated with `> ~/.lychee/logs/server.log` while the server is stopped.
- Set up log rotation via `logrotate` on Linux or a scheduled task on Windows/macOS.

**Symlink tip:**
```bash
# Centralize logs for monitoring
rm -rf ~/.lychee/logs
ln -s /var/log/lychee ~/.lychee/logs
```

---

## `history`

**Purpose:** A single file storing all chat conversation history — messages, timestamps, model responses, and context metadata.

**Format:** JSON Lines (one JSON object per conversation/message block) or SQLite, depending on the Lychee version. Check your version's documentation for the exact schema.

**Size characteristics:**
- Grows linearly with usage — typically a few MB per year of regular use.
- Long conversations with large contexts increase growth rate.

**Cleanup:**
- Delete to start fresh with no chat history.
- Append-only — no partial cleanup without manual JSON editing.
- Backup before deleting if you may want to reference old conversations.

**Symlink tip:**
```bash
# Keep history backed up in a synced folder
mv ~/.lychee/history ~/Documents/lychee-backups/history
ln -s ~/Documents/lychee-backups/history ~/.lychee/history
```

---

## `huggingface/`

**Purpose:** A standard Hugging Face Hub cache, used when Lychee downloads models or tokenizers directly from Hugging Face. This is a subset of the full `HF_HOME` structure.

```
huggingface/
└── hub/
    ├── models--meta-llama--Llama-3-8B-Instruct/
    │   ├── blobs/          # Raw file blobs (SHA256-addressed)
    │   ├── refs/           # Branch/tag pointers → blobs
    │   └── snapshots/      # Immutable version snapshots
    └── models--TheBloke--Mistral-7B-Instruct-GGUF/
        └── ...
```

### `huggingface/hub/`

**Purpose:** The standard Hugging Face `huggingface_hub` cache directory. Contains downloaded models, tokenizers, and any other HF artifacts.

**Key subdirectories per model:**
| Path | Purpose |
|------|---------|
| `blobs/` | Raw files keyed by SHA256 |
| `refs/main` | Points to latest commit |
| `snapshots/<hash>/` | Symlinks to blobs, forming a complete snapshot |

**Cleanup:**
- Use `huggingface-cli delete-cache` to interactively clean unused models.
- Safe to delete entirely if Lychee uses only local GGUF files (not HF-sourced models).
- The `blobs/` inside each model directory deduplicate with `models/blobs/` — if Lychee imports HF models into its own `models/` store, the HF cache may be redundant.

**Symlink tip:**
```bash
# Point to the global HF cache to share downloads across tools
rm -rf ~/.lychee/huggingface
ln -s ~/.cache/huggingface ~/.lychee/huggingface
# Or set environment variable before starting Lychee:
# export HF_HOME=~/.cache/huggingface
```

---

## Disk Space Summary

Sorted by typical size contribution:

| Directory | Typical Size | Growth Pattern |
|-----------|-------------|----------------|
| `models/blobs/` | 10–200+ GB | Per model download |
| `huggingface/hub/` | 10–100+ GB | Per HF model download |
| `logs/` | 10–500 MB | ~1–10 MB/day |
| `history` | 1–50 MB | ~1 MB/month |
| `config.yaml` | <5 KB | Static |
| `models/manifests/` | <1 MB | ~1 KB per model |

## Quick Commands

```bash
# Check total size
du -sh ~/.lychee/

# Check per-directory sizes
du -sh ~/.lychee/*/

# Find largest model blobs
ls -lhS ~/.lychee/models/blobs/ | head -10

# Clean old logs
rm ~/.lychee/logs/*.log.[0-9]*

# Clean HF cache (interactive)
huggingface-cli delete-cache
```

## Relocation Tips

To move the entire `~/.lychee/` directory to another location:

```bash
# Option 1: Symlink the whole thing
mv ~/.lychee /mnt/data/lychee
ln -s /mnt/data/lychee ~/.lychee

# Option 2: Set environment variable (if supported)
export LYCHEE_HOME=/mnt/data/lychee
```

Always stop the Lychee server before moving or modifying files under `~/.lychee/`.
