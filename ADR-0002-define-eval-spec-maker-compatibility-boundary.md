# ADR-0002: eval-spec-maker との互換境界を定義する / Define the eval-spec-maker Compatibility Boundary

- 状態: Accepted / Status: Accepted
- 日付: 2026-09-15 / Date: 2026-09-15

## コンテキスト / Context

casemd の Markdown 入力形式は、`ryuta46/eval-spec-maker` が公開した点検表の記法から着想を得ています。
casemd's Markdown input format is inspired by the inspection-sheet notation published by `ryuta46/eval-spec-maker`.

実際に eval-spec-maker をビルドして同一入力を変換すると、見出しの意味、シート名、列見出し、Markdown の解釈、XLSX のレイアウトに差分がありました。
Building eval-spec-maker and converting the same inputs revealed differences in heading semantics, sheet names, column headers, Markdown interpretation, and XLSX layout.

eval-spec-maker は `#` 見出しごとにシートを作成します。
eval-spec-maker creates a worksheet for each `#` heading.

casemd v1 は `#` を任意の文書タイトルとして扱い、入力ファイルごとにシートを作成します。
casemd v1 treats `#` as an optional document title and creates one worksheet for each input file.

eval-spec-maker は日本語固定の列見出しと独自のレイアウトを生成します。
eval-spec-maker generates fixed Japanese column headers and its own layout.

casemd は人間による実施記録を優先した独自の列見出しと XLSX レイアウトを生成します。
casemd generates its own column headers and XLSX layout with an emphasis on human execution records.

## 決定 / Decision

casemd v1 は eval-spec-maker と完全な入出力互換を提供しません。
casemd v1 does not provide full input or output compatibility with eval-spec-maker.

両者に共通する `##`、`###`、`####`、番号付きリスト、タスクリストの構造記法だけを互換の基盤とします。
Only the shared structural notation of `##`, `###`, `####`, ordered lists, and task lists forms the compatibility foundation.

casemd では `#` を文書タイトルとして最大 1 回だけ許可します。
casemd permits `#` once at most as a document title.

複数シートを生成する場合は `--input` を複数回指定します。
Use repeated `--input` options to generate multiple worksheets.

シート名は文書タイトルではなく入力ファイル名から導出します。
Worksheet names are derived from input file names rather than document titles.

列見出しと XLSX のスタイルは casemd の人間向け実施記録の契約として維持します。
Column headers and XLSX styles remain part of casemd's human-focused execution-record contract.

番号付きリストとタスクリストの各行は、リスト記号の後ろをそのまま保持します。
Each ordered-list or task-list line preserves the text after its list marker.

タスクリストの各行は、個別の実施記録列を持つ XLSX 行へ展開します。
Each task-list line expands into an XLSX row with its own execution-record columns.

インラインコード、引用、全角文字、半角文字は rich Markdown として再解釈しません。
Inline code, block quotes, full-width characters, and half-width characters are not reinterpreted as rich Markdown.

eval-spec-maker の Java または Gradle 環境は casemd のビルド、テスト、CI の依存関係にしません。
The Java or Gradle environment for eval-spec-maker is not a dependency of casemd builds, tests, or CI.

## 結果 / Consequences

eval-spec-maker の既存 Markdown を casemd へ移行する際は、複数の `#` 見出しを入力ファイルへ分割する必要があります。
Migrating existing eval-spec-maker Markdown to casemd requires splitting multiple `#` headings into input files.

casemd はリスト行に含まれる全角半角混在の文字列と改行を、生成した XLSX で読みやすく保持します。
casemd keeps mixed full-width and half-width text and line breaks in generated XLSX files for readability.

eval-spec-maker の出力との差分は不具合ではなく、ここで定義した v1 契約です。
Differences from eval-spec-maker output are not defects when they follow this v1 contract.

完全互換、見出しの多重カテゴリ、列見出しのローカライズは v1 の対象外です。
Full compatibility, repeated heading categories, and column-header localization are out of scope for v1.
