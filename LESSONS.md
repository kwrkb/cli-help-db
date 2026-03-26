# LESSONS

## Phase 1-3: 初期実装 (2026-03-26)

### --help の exit code は信用できない
- 多くのCLIツールは `--help` で non-zero exit する（docker, jq 等）
- **ルール**: `CombinedOutput` の結果が空でなければ exit code は無視する。`helptree/runner.go` と同じパターンを踏襲

### collector のフォールバック順序は重要
- `--help` → `-h` → `man` の順で試行。`-h` だけ受け付けるツールや、`--help` が存在しないツールがある
- **ルール**: 出力が10文字未満なら「取得失敗」と見なして次のソースにフォールバックする

### scanner の重複排除は first-wins
- `$PATH` に同じディレクトリが複数回含まれたり、異なるディレクトリに同名バイナリが存在する
- **ルール**: シェルの挙動と一致させるため、最初に見つかったものを採用する（`map[string]bool` で管理）

### フックスクリプトの除外リストは静的DB版では不要
- 動的 `auto-help.sh` では除外リスト（git, ls 等）でスキップしていたが、静的DB版ではそもそもDBに存在しないコマンドは自動スキップされる
- **ルール**: ホワイトリスト方式なら除外リストは冗長。ただし npx/bunx のような副作用があるコマンドのスキップはフック側で維持する

### Go プロジェクトの外部依存は最小限に
- CLI フレームワーク（cobra等）を使わず stdlib `flag` + `FlagSet` で5サブコマンドを十分にルーティングできた
- **ルール**: サブコマンドが10未満でネスト不要なら stdlib で十分。依存追加は `yaml.v3` のような「stdlib にない機能」に限定する

## Phase 4: Polish — build/update 統合 (2026-03-26)

### 似た機能のコマンドは統合してフラグで分岐させる
- `build`（全取得）と `update`（差分取得）を別コマンドにしていたが、ユーザーが使い分ける動機が弱かった
- **ルール**: デフォルト動作を安全側（差分更新）にし、`--force` で全再取得にする。コマンドを2つに分けるより1つ + フラグの方がUXが良い

### --all と --dry-run は相性が良い
- `--all` は `$PATH` 全コマンドを対象にするため数百〜千単位になる。`--dry-run` で事前確認できないと怖くて使えない
- **ルール**: 大量データを処理するフラグを追加する際は、必ず `--dry-run`（プレビュー）をセットで提供する

### コマンド削除時は関連ファイルを漏れなく消す
- `update` コマンド削除時に `update.go` の削除だけでなく、`root.go` の switch case と usage テキストからも除去が必要だった
- **ルール**: コマンド削除チェックリスト — (1) 実装ファイル (2) ルーター/ディスパッチャの登録 (3) usage/helpテキスト (4) README (5) テストファイル

## Phase 5: Lazy loading (2026-03-26)

### フックの動的処理は pure bash で完結させる
- lazy loading は Go の collector を呼ぶ案もあったが、フックは自己完結すべき（Go バイナリが PATH にない環境でも動く必要がある）
- **ルール**: フックスクリプトから外部バイナリ（自作ツール含む）への依存を作らない。bash + coreutils で完結させる

### 同期フックのタイムアウトはバッチ処理より短くする
- `build` コマンドは 3秒/コマンドだが、PreToolUse hook は同期的で Claude Code をブロックする。同じタイムアウトでは遅すぎる
- **ルール**: 同期フックの外部コマンド実行は 2秒以下に。バッチ処理とフックでタイムアウト値を分ける

### opt-in フラグで既存挙動を守る
- lazy loading を `--lazy` フラグで opt-in にし、デフォルト（フラグなし）は従来の静的参照のみ。既存ユーザーの挙動を変えない
- **ルール**: 動的処理や副作用（DB書き込み等）を伴う機能追加は opt-in フラグにする。デフォルトは安全側を維持

### macOS の UserConfigDir は ~/.config ではない
- Go の `os.UserConfigDir()` は macOS で `~/Library/Application Support` を返す。`~/.config/` にファイルを置いても読み込まれない
- `config.example.yaml` のコメントに `~/.config/cli-help-db/config.yaml` と書いてあるが、macOS では正しくない
- **ルール**: クロスプラットフォームで config path を案内する際は `os.UserConfigDir()` の実際の戻り値を確認する。README や example にはプラットフォームごとのパスを明記する

### coreutils の `timeout` は macOS にデフォルトで存在しない
- lazy hook が `timeout 2 "$BASE_CMD" --help` を使っていたが、macOS の標準シェル環境には GNU `timeout` がない。エラーメッセージが help 出力として誤保存される
- **ルール**: `timeout` を使う場合は `command -v timeout` で存在チェックし、なければ直接実行にフォールバックする。クロスプラットフォームのシェルスクリプトでは GNU coreutils 固有コマンドに注意

### シェル展開スキップは `$(...)` も含める
- `$VAR` と バッククォートのスキップはあったが、`$(cmd)` 形式のコマンド置換がすり抜けていた（レビューで指摘）
- **ルール**: シェル変数・展開のスキップパターンは `$'*'`、`` ` ``、`$('*'` の3種を必ずカバーする

## Phase 6: Windows 対応 (2026-03-26)

### [Bug] Windows の Scan() はファイル名に拡張子を含む
- **状況**: `cli-help-db build` が「none of the configured commands were found on $PATH」で全滅
- **原因**: `Scan()` が `curl.exe` を返すが、config のホワイトリストは `curl`。`Filter()` が完全一致で比較していたためマッチしない
- **ルール**: Windows では `Scan()` の結果を拡張子なしでもインデックスする。`runtime.GOOS == "windows"` ガードで分岐し、他 OS に影響を与えない

### [Config] Windows の isExecutable は権限ビットではなく拡張子で判定する
- **状況**: 既存テストが `0755` の権限ビットで「実行可能」を判定していたが、Windows では無意味
- **原因**: Windows は NTFS ACL で権限管理し、Unix の `mode&0111` は常に true を返す。実行可能判定は PATHEXT 拡張子（`.exe`, `.cmd`, `.bat` 等）で行う必要がある
- **ルール**: テストで実行可能ファイルを作る場合は OS に応じた命名にする（`execName()` ヘルパーパターン）。Windows 固有テストは `//go:build windows` で分離する

### [Process] WSL で作ったツールは Windows ネイティブで必ず検証する
- **状況**: WSL 上で全フェーズ完了・テスト PASS だったが、Windows ネイティブでは `scanner_test.go` が FAIL、`Filter()` が機能しなかった
- **原因**: WSL は Linux 環境であり、Windows 固有のパス解決・拡張子・権限モデルの差異が見えない
- **ルール**: クロスプラットフォーム対応を謳うなら CI に全ターゲット OS を含める。最低限 `ubuntu-latest` + `windows-latest` の matrix を設定する

### [Config] OS 固有テストファイルにはビルドタグを忘れない
- **状況**: `scanner_windows_test.go` に `//go:build windows` を付け忘れ、Linux CI で `0644` のファイルが実行可能と判定されず FAIL する
- **原因**: Go のファイル名規約 `_windows_test.go` はビルドタグの代替にならない（テストでは無視される）
- **ルール**: `_windows_test.go` / `_linux_test.go` には必ず先頭行に `//go:build` タグを付与する

### [Design] テストは内部ロジック再実装ではなく公開 API を叩く
- **状況**: `TestFilter_WindowsExtensionStripping` が `Filter()` を呼ばず、内部のマップ構築ロジックを再実装していた（PR レビューで指摘）
- **原因**: `Filter()` が内部で `Scan()` → `$PATH` を参照するため直接テストしづらいと判断し、ロジックをコピーした
- **ルール**: `t.Setenv("PATH", dir)` で環境を差し替えて公開関数を直接テストする。内部ロジックの再実装はリファクタ耐性がない

### [Tool] Git Bash は gh api のパスを勝手にファイルパスに変換する
- **状況**: `gh api /repos/owner/repo/...` が `C:/Program Files/Git/repos/...` に変換されてエラー
- **原因**: MSYS2 (Git Bash) が `/` 始まりの引数を Windows ファイルパスに自動変換する
- **ルール**: Git Bash で `gh api` を使う場合は先頭の `/` を省略する（`repos/owner/repo/...`）
