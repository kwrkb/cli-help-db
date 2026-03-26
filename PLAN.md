# PLAN: cli-help-db

## Context

現在の `auto-help.sh` フックは毎回動的に `--help` を実行している。タイムアウトリスク・トークンコスト・セッション単位キャッシュの揮発性が課題。`cli-help-db` は静的ヘルプDBを事前生成し、フックからはファイルを読むだけにする。

## Design Decisions

- **CLI**: stdlib `flag` のみ（FlagSet per subcommand）
- **Config**: `gopkg.in/yaml.v3` のみ（外部依存1つ）
- **DB format**: `~/.claude/cli-help/{command}.txt`（プレーンテキスト）
- **helptree参考**: CombinedOutput + non-zero exit許容パターンを踏襲

## Phase 1: Scaffolding + scan + build [x] DONE

- [x] `go mod init github.com/kwrkb/cli-help-db` + yaml.v3
- [x] `internal/config/` — YAML読み込み、`~`展開、デフォルト値
- [x] `internal/scanner/` — `$PATH` 列挙、実行可能判定、重複排除
- [x] `internal/collector/` — `--help` → `-h` → `man` フォールバック、タイムアウト、行数制限、並列実行
- [x] `internal/db/` — `.txt` 書き込み・読み取り・一覧
- [x] `internal/cmd/` — root, scan, build サブコマンド
- [x] `main.go` — エントリポイント
- [x] テスト全PASS、実コマンド（curl, jq, docker）でbuild動作確認済

## Phase 2: update + list [x] DONE

- [x] `internal/cmd/list.go` — DB内コマンド一覧表示
- [x] `internal/cmd/update.go` — config whitelist と既存DB の差分で未取得分のみ収集
- [x] 動作確認済（up-to-date判定、PATHフィルタ）

## Phase 3: hook生成 [x] DONE

- [x] `internal/hook/hook.go` — 現行 `auto-help.sh` のコマンド抽出・除外ロジックを維持、DB参照版スクリプト生成
- [x] `internal/cmd/hook.go` — stdout出力
- [x] テストPASS、生成スクリプト確認済
- [x] 旧 `auto-help.sh` を削除、`settings.json` の hooks エントリも除去
  > cli-help-db の hook コマンドで新版を生成する運用に切り替え済

## Phase 4: Polish [ ]

- [ ] `--all` フラグ（ホワイトリスト無視でフルスキャン）
- [ ] `--dry-run`, `--verbose`, `--quiet`
- [ ] 実行サマリ改善
- [ ] `README.md`
- [ ] `.gitignore` 作成
- [ ] 設定ファイルのサンプル同梱

## Key Files to Reference

- `/home/yugosasaki/.claude/hooks/auto-help.sh` — 旧フック（削除済。cli-help-db hook で新版を生成する）
- `/home/yugosasaki/code/helptree/internal/runner/runner.go` — CombinedOutput パターン
