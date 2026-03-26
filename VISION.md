# VISION: cli-help-db

## Problem

Claude Code sometimes misuses CLI tool options (wrong flags, deprecated syntax, nonexistent arguments). Injecting accurate `--help` output via `PreToolUse` hook's `additionalContext` significantly improves tool-use accuracy — but fetching help dynamically on every invocation introduces timeout risk and token cost.

## Solution

A Go CLI tool that **pre-builds a static help database** from commands on `$PATH`. The database is a directory of plain-text files (one per command), designed to be looked up instantly by a shell hook at Claude Code runtime.

## Core Commands

### `scan`
Enumerate executable files on `$PATH` and display them. Useful for discovering available commands before building the database.

### `build`
Collect help text for specified commands and save as one file per command. Incremental by default (skips existing entries); `--force` for full re-collection.

- **Help source fallback order**: `--help` → `-h` → `man`
- **Line limit**: Trim output to a configurable maximum (default: 60 lines)
- **Timeout**: Per-command execution timeout (default: 3 seconds)
- **Parallelism**: Concurrent collection with bounded goroutines
- **`--all`**: Scan all `$PATH` commands instead of config whitelist
- **`--dry-run`**: Preview target commands without collecting

### `list`
Display commands currently stored in the database.

### `hook`
Generate an `auto-help.sh` hook script that Claude Code can use to look up help text from the database at runtime via `additionalContext`.

- **`--lazy`**: Enable on-demand collection — automatically fetch and cache `--help` for unknown commands on first use

## Design Principles

### Whitelist-First
The default mode operates on an explicit whitelist of commands defined in the config file. Users control exactly which tools get indexed.

A full-scan mode (`--all`) is available but opt-in — scanning every binary on `$PATH` without a whitelist is noisy and slow.

### Configuration
- **Config file**: `~/.config/cli-help-db/config.yaml`
- **Output directory**: `~/.claude/cli-help/` (default, configurable)
- Config specifies: command whitelist, line limit, timeout, output path

### Minimal Dependencies
Standard library only where possible. No frameworks. External dependencies are added only when they provide clear, justified value (e.g., YAML/TOML parsing).

### Cross-Platform
Works on Linux, macOS, and Windows. `$PATH` scanning and command execution adapt to the host OS.

## Out of Scope

- **LLM-based help summarization** — potential future enhancement, not in initial scope
- **GUI / TUI** — no Bubble Tea or similar; this is a plain CLI tool
- **Package manager distribution** — no Homebrew formula, winget manifest, etc. for now
