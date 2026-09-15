# casemd

[![CI](https://github.com/9renpoto/casemd/actions/workflows/ci.yml/badge.svg)](https://github.com/9renpoto/casemd/actions/workflows/ci.yml)
[![codecov](https://codecov.io/gh/9renpoto/casemd/graph/badge.svg?token=D63wbdaCah)](https://codecov.io/gh/9renpoto/casemd)

Convert structured Markdown inspection checklists into CSV files, Excel workbooks, and Google Spreadsheets.

The Markdown files are the source of truth for inspection cases.
Each output is a generated artifact that can be shared with a testing or operations team.

<details>
<summary>日本語</summary>

構造化した Markdown の点検チェックリストを、CSV ファイル、Excel ブック、Google スプレッドシートに変換します。

点検ケースの正本は Markdown ファイルです。
各出力は生成物として、テストチームや運用チームと共有できます。

</details>

## Install

### Homebrew

Install the latest released version with Homebrew:

```sh
brew install 9renpoto/tap/casemd
```

Verify the installed version:

```sh
casemd --version
```

## Web UI scenario tests

The manual scenario checklist for the Web UI is maintained in [`scenarios/web-ui.md`](scenarios/web-ui.md) in Japanese.
It is limited to browser-visible behavior such as page loading, preview conversion, error display, keyboard interaction, and responsive layout.
The sample cases intentionally mix full-width Japanese characters with half-width Latin characters, numbers, symbols, and long text for layout dogfooding.

After a change is merged into `main`, CI converts this Markdown file to `web-ui-scenarios.xlsx` and publishes it as the `web-ui-scenarios` workflow artifact.
Execution results, test dates, testers, and notes can be recorded in the generated workbook.

## Releases

The `Bump version` workflow runs every Tuesday and can also be dispatched manually for the intended SemVer increment.
If no GitHub Release exists yet, a `minor` bump bootstraps `v0.1.0`.
It creates a draft pull request from the `release` branch containing the generated CHANGELOG entry.

Review the draft pull request's version, release notes, labels, and milestone, then merge it after CI succeeds.
After the merge, make sure the GitHub Actions secret `HOMEBREW_TAP_GITHUB_TOKEN` has `contents: write` permission for `9renpoto/homebrew-tap`, and tag the merged commit with its CHANGELOG version:

```sh
git checkout main
git pull --ff-only
git tag --annotate vX.Y.Z --message "Release vX.Y.Z"
git push origin vX.Y.Z
```

Pushing the tag triggers the release workflow, which creates a GitHub Release, uploads Darwin and Linux archives with `checksums.txt`, and updates the Homebrew tap.

After it completes, verify the published install path in a clean environment:

```sh
brew install 9renpoto/tap/casemd
casemd --version
```

### From source

### Requirements

- Go 1.26 or later.
- An OAuth access token with the `https://www.googleapis.com/auth/spreadsheets` scope when creating Google Spreadsheets.

Install the CLI with Go:

```sh
go install github.com/9renpoto/casemd/cmd/casemd@latest
```

You can also run it directly from a checkout with `go run ./cmd/casemd`.

<details>
<summary>日本語</summary>

### 必要条件

- Go 1.26 以降。
- Google スプレッドシートを作成する場合は、`https://www.googleapis.com/auth/spreadsheets` スコープを持つ OAuth アクセストークン。

Go で CLI をインストールします。

```sh
go install github.com/9renpoto/casemd/cmd/casemd@latest
```

チェックアウトしたリポジトリから `go run ./cmd/casemd` で直接実行することもできます。

</details>

## Usage

Show the available flags:

```sh
casemd --help
```

Convert one Markdown file to CSV and XLSX:

```sh
casemd \\
  --input notes.md \\
  --csv-output build/notes.csv \\
  --spreadsheet-output build/notes.xlsx
```

Pass `--input` more than once to combine files into one workbook.
Each input file becomes a separate worksheet in the workbook.

```sh
casemd \\
  --input notes.md \\
  --input follow-up.md \\
  --spreadsheet-output build/all-notes.xlsx
```

Create a Google Spreadsheet by providing an access token:

```sh
GOOGLE_SHEETS_ACCESS_TOKEN=ya29.example-token \\
  casemd \\
  --input notes.md \\
  --google-spreadsheet-title "Inspection Sheet Export"
```

Start the local preview web UI on port 3000:

```sh
casemd serve
```

Set `CASEMD_WEB_ADDR` or pass an address as the second argument to change the bind address.
The web UI exposes `GET /healthz` and `POST /api/preview` in addition to the browser interface.

<details>
<summary>日本語</summary>

利用できるフラグを表示します。

```sh
casemd --help
```

Markdown ファイルを CSV と XLSX に変換します。

```sh
casemd \\
  --input notes.md \\
  --csv-output build/notes.csv \\
  --spreadsheet-output build/notes.xlsx
```

`--input` を複数回指定すると、ファイルを 1 つのブックにまとめられます。
入力ファイルごとにブック内のワークシートが 1 枚作成されます。

アクセストークンを指定して Google スプレッドシートを作成します。

```sh
GOOGLE_SHEETS_ACCESS_TOKEN=ya29.example-token \\
  casemd \\
  --input notes.md \\
  --google-spreadsheet-title "Inspection Sheet Export"
```

ポート 3000 でローカルのプレビュー Web UI を起動します。

```sh
casemd serve
```

バインドアドレスを変更するには `CASEMD_WEB_ADDR` を設定するか、第 2 引数にアドレスを指定します。
ブラウザー画面に加えて、`GET /healthz` と `POST /api/preview` を提供します。

</details>

## Input format

Use headings to define the inspection hierarchy and lists to define execution details:

| Markdown element | Meaning | Output column |
| --- | --- | --- |
| `#` heading | Optional document title | — |
| `##` heading | Major item | Major Item |
| `###` heading | Medium item | Medium Item |
| `####` heading | Individual inspection case | Minor Item |
| Ordered list such as `1.` | Validation steps | Validation Steps |
| Task list such as `* [ ]` | Checkpoints | Checkpoints |

The generated table also contains blank `Result`, `Test Date`, `Tester`, and `Notes` columns for execution records.
Each checkpoint becomes a separate row so its execution record is independent.
Validation steps retain their order and line breaks on every checkpoint row.
Inline code written with backticks, such as `` `user@example.com` ``, is exported without its delimiters and shown in a monospace font in Excel and Google Sheets.
An unmatched backtick delimiter remains literal so no scenario text is silently lost.

Example:

```markdown
# Inspection Sheet

## Setup
### Environment
#### Dependencies

1. Install required packages
2. Confirm default configurations
* [ ] Packages installed successfully
* [ ] Defaults match specification
```

The repository includes [`notes.md`](notes.md) and [`follow-up.md`](follow-up.md) as working examples.

<details>
<summary>日本語</summary>

見出しで点検階層を定義し、リストで実行内容を定義します。

| Markdown 要素 | 意味 | 出力列 |
| --- | --- | --- |
| `#` 見出し | 任意のドキュメントタイトル | — |
| `##` 見出し | 大項目 | Major Item |
| `###` 見出し | 中項目 | Medium Item |
| `####` 見出し | 個別の点検ケース | Minor Item |
| `1.` などの番号付きリスト | 検証手順 | Validation Steps |
| `* [ ]` などのタスクリスト | チェックポイント | Checkpoints |

生成される表には、実行記録用に `Result`、`Test Date`、`Tester`、`Notes` の空列も含まれます。
チェックポイントは 1 件ずつ行へ展開されるため、個別に実施記録を残せます。
検証手順は各チェックポイント行で順序と改行を保持します。
`` `user@example.com` `` のようなバッククォートによるインラインコードは、区切り記号を除いて Excel と Google スプレッドシートでは等幅フォントで表示します。
閉じていないバッククォートは、シナリオ本文を失わないようリテラルのまま出力します。

リポジトリには実行例として [`notes.md`](notes.md) と [`follow-up.md`](follow-up.md) が含まれています。

</details>

## eval-spec-maker compatibility

casemd is inspired by [`ryuta46/eval-spec-maker`](https://github.com/ryuta46/eval-spec-maker), but it is not a drop-in compatible replacement.

| Behavior | casemd v1 | eval-spec-maker |
| --- | --- | --- |
| `#` heading | Optional document title, at most once | Worksheet category, repeatable |
| Worksheet unit | Each input file | Each `#` heading |
| Column headers | casemd execution-record contract | Fixed Japanese labels |
| Markdown list content | Preserve text after the list marker | Interpret Markdown tokens in the list content |
| XLSX layout | Human-focused casemd layout | eval-spec-maker layout |

Use repeated `--input` options when a casemd workbook needs multiple worksheets.
Read [`ADR-0002`](docs/adr/0002-define-eval-spec-maker-compatibility-boundary.md) before migrating existing eval-spec-maker specifications.

Project design decisions and terminology are documented in [`docs/`](docs/), with ADRs under [`docs/adr/`](docs/adr/).

## API

The preview server provides the following HTTP endpoints when `casemd serve` is running:

| Method | Path | Description |
| --- | --- | --- |
| `GET` | `/` | Open the Markdown-to-CSV preview UI. |
| `GET` | `/healthz` | Return `{"status":"ok"}` when the server is ready. |
| `POST` | `/api/preview` | Convert a JSON body containing `name` and `markdown` into CSV. |

Example request:

```sh
curl --fail \\
  -H 'Content-Type: application/json' \\
  -d '{"name":"notes.md","markdown":"## Setup\\n### Environment\\n#### Dependencies"}' \\
  http://localhost:3000/api/preview
```

## Development

Install the Git hooks and run the repository checks:

```sh
lefthook install
lefthook run pre-commit
```

Run focused checks during development:

```sh
go build ./cmd/casemd
go test ./... -cover
go vet ./...
typos
```

The project follows clean-architecture boundaries.
Application orchestration lives in `internal/app`, domain parsing lives in `internal/core`, and interface adapters live in `internal/interfaces`.

## Contributing

Bug reports and pull requests are welcome.
Please keep changes focused, add or update tests for behavior changes, run `lefthook run pre-commit`, and describe architectural impact in the pull request.

## Security

Please report security issues according to [`SECURITY.md`](SECURITY.md).

## Acknowledgments

The Markdown inspection-sheet format is inspired by [`ryuta46/eval-spec-maker`](https://github.com/ryuta46/eval-spec-maker).

## License

[MIT](LICENSE) © 2026 9renpoto

<details>
<summary>日本語</summary>

## API

`casemd serve` の起動中は、次の HTTP エンドポイントを利用できます。

| メソッド | パス | 説明 |
| --- | --- | --- |
| `GET` | `/` | Markdown から CSV へのプレビュー UI を開きます。 |
| `GET` | `/healthz` | サーバーが準備できている場合に `{"status":"ok"}` を返します。 |
| `POST` | `/api/preview` | `name` と `markdown` を含む JSON を CSV に変換します。 |

## 開発

Git フックをインストールして、リポジトリのチェックを実行します。

```sh
lefthook install
lefthook run pre-commit
```

## コントリビューション

バグ報告とプルリクエストを歓迎します。
変更は小さく保ち、動作を変更する場合はテストを追加または更新し、`lefthook run pre-commit` を実行してください。

## セキュリティ

セキュリティ上の問題は [`SECURITY.md`](SECURITY.md) に従って報告してください。

## 謝辞

Markdown の点検シート形式は [`ryuta46/eval-spec-maker`](https://github.com/ryuta46/eval-spec-maker) に着想を得ています。

## ライセンス

[MIT](LICENSE) © 2026 9renpoto

</details>
