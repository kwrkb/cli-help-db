# cli-help-db

CLI の --help 出力を静的DBに事前キャッシュし、Claude Code の PreToolUse hook で即時参照するツール。

## Commands

```bash
go build ./...       # ビルド
go test ./...        # テスト実行
go vet ./...         # 静的解析（CIと同等）
go run . scan        # $PATH のコマンド一覧
go run . build       # ヘルプDB構築（増分）
go run . list        # DB内コマンド一覧
go run . hook --lazy # hook スクリプト生成
```

## Architecture

```
main.go → internal/cmd/root.go (サブコマンドルーター)
  ├── scan.go  → scanner/   ($PATH スキャン)
  ├── build.go → collector/ (並列 --help 収集) → db/ (ファイルDB書き込み)
  ├── list.go  → db/       (DB一覧)
  └── hook.go  → hook/     (bash スクリプト生成、通常 + lazy モード)
```

- 外部依存: `gopkg.in/yaml.v3` のみ
- DB形式: `~/.claude/cli-help/{command}.txt`（プレーンテキスト）
- hook テンプレートは Go の `fmt.Fprintf` 経由で生成される。`%%` は Go のエスケープで、出力は bash の `%`（最短一致）になる

## Testing

- collector は Executor インターフェースでモック化。実コマンド実行なし
- scanner は Windows 用テスト (`scanner_windows_test.go`) あり
- CI: ubuntu-latest + windows-latest
