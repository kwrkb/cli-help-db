# VISION: cli-help-db

## Problem

Claude Code sometimes misuses CLI tool options (wrong flags, deprecated syntax, nonexistent arguments). Injecting accurate `--help` output via `PreToolUse` hook's `additionalContext` significantly improves tool-use accuracy — but fetching help dynamically on every invocation introduces timeout risk and token cost.

## Solution

A Go CLI tool that **pre-builds a static help database** from commands on `$PATH`. The database is a directory of plain-text files (one per command), designed to be looked up instantly by a shell hook at Claude Code runtime.

## Design Principles

### Whitelist-First
The default mode operates on an explicit whitelist of commands defined in the config file. Users control exactly which tools get indexed.

A full-scan mode (`--all`) is available but opt-in — scanning every binary on `$PATH` without a whitelist is noisy and slow.

### Minimal Dependencies
Standard library only where possible. No frameworks. External dependencies are added only when they provide clear, justified value (e.g., YAML/TOML parsing).

### Cross-Platform
Works on Linux, macOS, and Windows. `$PATH` scanning and command execution adapt to the host OS.

## Out of Scope

- **LLM-based help summarization** — potential future enhancement, not in initial scope
- **GUI / TUI** — no Bubble Tea or similar; this is a plain CLI tool
- **Package manager distribution** — no Homebrew formula, winget manifest, etc. for now
