# cli-help-db

Pre-build a static `--help` database for CLI commands, so [Claude Code](https://docs.anthropic.com/en/docs/claude-code) can look up accurate flag/option references instantly via a `PreToolUse` hook.

## Why

Claude Code sometimes misuses CLI options — wrong flags, deprecated syntax, nonexistent arguments. Injecting `--help` output through `additionalContext` fixes this, but fetching it dynamically every time risks timeouts and wastes tokens. This tool pre-collects help text into plain `.txt` files for instant, zero-cost lookup.

## Install

```bash
go install github.com/kwrkb/cli-help-db@latest
```

## Quick Start

**1. Create a config file**

```bash
mkdir -p ~/.config/cli-help-db
cat > ~/.config/cli-help-db/config.yaml << 'EOF'
commands:
  - curl
  - jq
  - docker
  - kubectl
  - terraform
  - gh
  - aws
EOF
```

**2. Build the database**

```bash
cli-help-db build
```

This collects `--help` output for each command and saves it to `~/.claude/cli-help/` (one `.txt` file per command).

**3. Generate and install the hook**

```bash
cli-help-db hook > ~/.claude/hooks/auto-help.sh
chmod +x ~/.claude/hooks/auto-help.sh
```

Then add to `~/.claude/settings.json`:

```json
{
  "hooks": {
    "PreToolUse": [
      {
        "matcher": "Bash",
        "hooks": [
          {
            "type": "command",
            "command": "bash ~/.claude/hooks/auto-help.sh"
          }
        ]
      }
    ]
  }
}
```

## Commands

| Command | Description |
|---------|-------------|
| `scan` | List all executable commands on `$PATH` |
| `build` | Collect `--help` and build the database |
| `list` | Show commands stored in the database |
| `hook` | Generate `auto-help.sh` hook script |

### `build` flags

| Flag | Description |
|------|-------------|
| `--force` | Re-collect all commands, ignoring existing database entries |
| `--all` | Scan all `$PATH` commands instead of config whitelist |
| `--dry-run` | Show target commands without actually collecting |
| `-config <path>` | Override config file path |

By default, `build` only collects help for commands not already in the database (incremental). Use `--force` to re-collect everything. Flags can be combined (e.g., `build --all --dry-run`).

## Config

`~/.config/cli-help-db/config.yaml` (see [`config.example.yaml`](config.example.yaml) for a full example):

```yaml
commands:        # Whitelist of commands to index
  - curl
  - jq
  - docker
output_dir: ~/.claude/cli-help   # Default
line_limit: 60                   # Max lines per command
timeout: 3s                      # Per-command timeout
parallelism: 8                   # Concurrent workers
```

## How It Works

1. **Build phase** (`cli-help-db build`): For each whitelisted command (or all `$PATH` commands with `--all`), tries `--help`, falls back to `-h`, then `man`. Output is trimmed to the configured line limit and saved as `~/.claude/cli-help/{command}.txt`. Already-collected commands are skipped unless `--force` is used.

2. **Runtime phase** (hook): When Claude Code invokes the Bash tool, the hook extracts the command name and looks up `{command}.txt` in the database. If found, it injects the help text via `additionalContext` — no subprocess execution, no timeout risk.

## License

MIT

---

# cli-help-db (日本語)

CLIコマンドの `--help` 出力を事前収集し、[Claude Code](https://docs.anthropic.com/en/docs/claude-code) が `PreToolUse` フック経由で即座に参照できる静的ヘルプデータベースを生成するGoツール。

## なぜ必要か

Claude Code は CLI ツールのオプションを誤ることがある（存在しないフラグ、非推奨の構文など）。`additionalContext` で `--help` 情報を注入すれば精度が上がるが、毎回動的に実行するとタイムアウトリスクとトークンコストがかかる。このツールはヘルプテキストをプレーン `.txt` ファイルとして事前収集し、実行時コストゼロで参照可能にする。

## インストール

```bash
go install github.com/kwrkb/cli-help-db@latest
```

## クイックスタート

**1. 設定ファイルを作成**

```bash
mkdir -p ~/.config/cli-help-db
cat > ~/.config/cli-help-db/config.yaml << 'EOF'
commands:
  - curl
  - jq
  - docker
  - kubectl
  - terraform
  - gh
  - aws
EOF
```

**2. データベースをビルド**

```bash
cli-help-db build
```

各コマンドの `--help` 出力を `~/.claude/cli-help/` に保存する（1コマンド1ファイル）。

**3. フックを生成・インストール**

```bash
cli-help-db hook > ~/.claude/hooks/auto-help.sh
chmod +x ~/.claude/hooks/auto-help.sh
```

`~/.claude/settings.json` に追加:

```json
{
  "hooks": {
    "PreToolUse": [
      {
        "matcher": "Bash",
        "hooks": [
          {
            "type": "command",
            "command": "bash ~/.claude/hooks/auto-help.sh"
          }
        ]
      }
    ]
  }
}
```

## コマンド一覧

| コマンド | 説明 |
|---------|------|
| `scan` | `$PATH` 上の実行可能コマンドを一覧表示 |
| `build` | `--help` を収集してデータベースを構築 |
| `list` | データベース内のコマンド一覧を表示 |
| `hook` | `auto-help.sh` フックスクリプトを生成 |

### `build` のフラグ

| フラグ | 説明 |
|-------|------|
| `--force` | 既存DBを無視して全コマンド再取得 |
| `--all` | ホワイトリスト無視で `$PATH` 全コマンドを対象にビルド |
| `--dry-run` | 実際に取得せず対象コマンド一覧を表示 |
| `-config <path>` | 設定ファイルのパスを指定 |

デフォルトでは未取得のコマンドのみ収集する（差分更新）。`--force` で全再取得。フラグは併用可能（例: `build --all --dry-run`）。

## 設定

`~/.config/cli-help-db/config.yaml`（完全な例は [`config.example.yaml`](config.example.yaml) を参照）:

```yaml
commands:        # 対象コマンドのホワイトリスト
  - curl
  - jq
  - docker
output_dir: ~/.claude/cli-help   # デフォルト出力先
line_limit: 60                   # 1コマンドあたりの最大行数
timeout: 3s                      # コマンド実行タイムアウト
parallelism: 8                   # 並列ワーカー数
```

## 動作の仕組み

1. **ビルド時** (`cli-help-db build`): ホワイトリストの各コマンド（`--all` 指定時は `$PATH` 全コマンド）に対して `--help` → `-h` → `man` の順で試行。出力を行数制限でトリミングし、`~/.claude/cli-help/{command}.txt` として保存。`--force` 指定がない場合、既存のコマンドはスキップする。

2. **実行時** (フック): Claude Code が Bash ツールを呼ぶと、フックがコマンド名を抽出し `{command}.txt` を参照。ファイルがあれば `additionalContext` としてヘルプテキストを注入する。サブプロセス実行なし、タイムアウトリスクなし。

## ライセンス

MIT
