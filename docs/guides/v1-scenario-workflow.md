# Run a v1 Scenario Workflow

Use this guide to validate a v1 scenario specification, generate an XLSX workbook, verify its presentation, and record the execution result.

## Prerequisites

- Go 1.26 or later is available when running casemd from a checkout.
- An Excel-compatible spreadsheet application is available for the visual verification.

## Validate the specification

Validate the representative scenario before generating an artifact.

```sh
go run ./cmd/casemd validate --input testdata/human-first-scenario.md
```

The command exits with code `0` and produces no CSV, XLSX, or Google Spreadsheet when the specification is valid.

Use the invalid compatibility fixture to see the failure behavior.

```sh
go run ./cmd/casemd validate --input testdata/eval-spec-maker-multiple-categories.md
```

An invalid specification exits with a non-zero code and writes diagnostics to standard error.

Each diagnostic identifies the source, line, rule, message, and a suggested correction when one is available.

Correct every diagnostic and run validation again before converting the specification.

## Generate the execution workbook

Create an XLSX workbook from the valid representative scenario.

```sh
go run ./cmd/casemd \
  --input testdata/human-first-scenario.md \
  --spreadsheet-output build/human-first-scenario.xlsx
```

The converter applies the same validation rules as `casemd validate`.

It does not produce a partial CSV, XLSX, or Google Spreadsheet when an input is invalid.

Each checkpoint becomes an XLSX row, while the ordered validation steps remain in order on every related checkpoint row.

## Verify the workbook visually

Open `build/human-first-scenario.xlsx` in an Excel-compatible spreadsheet application.

Use the following checklist before handing the workbook to a tester.

- [ ] Long scenario text is wrapped and readable without covering adjacent cells.
- [ ] Multiple validation steps and checkpoint lists preserve their line breaks and are top-aligned.
- [ ] The header row remains visible while scrolling and filters are available for every data column.
- [ ] The test-definition columns are visually distinct from `Result`, `Test Date`, `Tester`, and `Notes`.
- [ ] The execution-record columns can be edited without changing the scenario definition.

Record any presentation defect as a product issue with the input file, spreadsheet application, and visible result.

## Record execution results

During manual execution, enter results in the generated workbook's `Result`, `Test Date`, `Tester`, and `Notes` columns.

Keep the Markdown file as the source of truth for the scenario definition.

Do not copy results, execution dates, testers, or notes back into Markdown.

When the Markdown definition changes, generate a new workbook and transfer execution records deliberately when they are still applicable.

## Distinguish the Web UI checklist

`scenarios/web-ui.md` is a separate Japanese checklist for browser-visible behavior.

After a merge to `main`, CI exports it as `web-ui-scenarios.xlsx` for manual Web UI testing.

Use `testdata/human-first-scenario.md` for this guide's general XLSX presentation verification.

## v1 boundary

Automated scenario execution, reverse synchronization, v1 metadata extensions, and Google Sheets-specific UX are outside the v1 scope.

Read [ADR-0001](../adr/0001-manage-scenario-test-specifications-in-markdown.md) for the scenario-workflow decision and [ADR-0002](../adr/0002-define-eval-spec-maker-compatibility-boundary.md) for the eval-spec-maker compatibility boundary.

<details>
<summary>日本語</summary>

# v1 シナリオワークフローを実行する

このガイドでは、v1 シナリオ仕様の検証、XLSX workbook の生成、表示確認、実施結果の記録を行います。

## 前提条件

- チェックアウトから casemd を実行する場合は Go 1.26 以降を利用できます。
- 表示確認には Excel 互換の表計算アプリケーションを利用できます。

## 仕様を検証する

成果物を生成する前に代表シナリオを検証します。

```sh
go run ./cmd/casemd validate --input testdata/human-first-scenario.md
```

仕様が正しい場合、コマンドは終了コード `0` で終了し、CSV、XLSX、Google Spreadsheet を生成しません。

不正な場合の挙動は、互換性確認用の不正 fixture で確認できます。

```sh
go run ./cmd/casemd validate --input testdata/eval-spec-maker-multiple-categories.md
```

仕様が不正な場合、コマンドは非 `0` の終了コードで終了し、標準エラー出力へ診断を表示します。

各診断には source、行番号、ルール、メッセージ、可能な場合は修正案が含まれます。

仕様を変換する前にすべての診断を修正し、再度検証してください。

## 実施用 workbook を生成する

正しい代表シナリオから XLSX workbook を生成します。

```sh
go run ./cmd/casemd \
  --input testdata/human-first-scenario.md \
  --spreadsheet-output build/human-first-scenario.xlsx
```

変換処理は `casemd validate` と同じ検証規則を適用します。

入力が不正な場合、部分的な CSV、XLSX、Google Spreadsheet は生成しません。

各チェックポイントは XLSX の 1 行になり、番号付きの検証手順は関連する各チェックポイント行で順序を保ちます。

## workbook を表示確認する

Excel 互換の表計算アプリケーションで `build/human-first-scenario.xlsx` を開きます。

workbook をテスト担当者へ渡す前に、次の項目を確認します。

- [ ] 長いシナリオの文字列が折り返され、隣のセルに重ならず読める。
- [ ] 複数の検証手順とチェックポイントのリストが改行を保持し、上揃えで表示される。
- [ ] スクロール中もヘッダー行を確認でき、すべてのデータ列でフィルターを利用できる。
- [ ] テスト定義列と `Result`、`Test Date`、`Tester`、`Notes` が視覚的に区別できる。
- [ ] シナリオ定義を変更せずに実施記録列を編集できる。

表示上の問題を見つけた場合は、入力ファイル、表計算アプリケーション、表示結果を含めてプロダクト Issue として記録します。

## 実施結果を記録する

手動で実施する際は、生成された workbook の `Result`、`Test Date`、`Tester`、`Notes` 列に結果を記録します。

シナリオ定義の正本は Markdown ファイルに保ちます。

結果、実施日、担当者、備考を Markdown へ書き戻してはいけません。

Markdown の定義が変更された場合は新しい workbook を生成し、適用できる実施記録だけを意図的に転記します。

## Web UI チェックリストと区別する

`scenarios/web-ui.md` はブラウザーに表示される動作を対象とする、独立した日本語チェックリストです。

`main` へのマージ後、CI はこれを手動 Web UI テスト用の `web-ui-scenarios.xlsx` として出力します。

このガイドの一般的な XLSX 表示確認には `testdata/human-first-scenario.md` を使います。

## v1 の境界

自動シナリオ実行、逆同期、v1 メタデータ拡張、Google Sheets 固有の UX は v1 の対象外です。

シナリオワークフローの決定は [ADR-0001](../adr/0001-manage-scenario-test-specifications-in-markdown.md)、eval-spec-maker 互換性の境界は [ADR-0002](../adr/0002-define-eval-spec-maker-compatibility-boundary.md) を参照してください。

</details>
