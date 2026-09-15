# ADR-0002: Define the eval-spec-maker Compatibility Boundary

- Status: Accepted
- Date: 2026-09-15

## Context

casemd's Markdown input is inspired by the inspection-sheet notation published by [`ryuta46/eval-spec-maker`](https://github.com/ryuta46/eval-spec-maker). Building the reference product and converting the same inputs revealed differences in heading semantics, worksheet names, column headers, Markdown interpretation, and XLSX layout.

eval-spec-maker creates a worksheet for each `#` heading, while casemd treats `#` as an optional document title and creates one worksheet per input file. The products also use different labels and layouts.

## Decision

casemd v1 does not provide full input or output compatibility with eval-spec-maker. The compatibility foundation is limited to the shared structure of `##`, `###`, `####`, ordered lists, and task lists.

- `#` is allowed at most once as the document title.
- Multiple worksheets are created by repeating `--input`.
- Worksheet names come from input file names.
- casemd's column headers and XLSX styles are its human execution-record contract.
- Text after ordered-list and task-list markers is preserved as authored.
- Each task-list line expands to an XLSX row with independent execution-record columns.
- Inline code is rendered as spreadsheet rich text, while block quotes and mixed-width characters are preserved as authored text.
- eval-spec-maker's Java and Gradle environment is not a casemd build, test, or CI dependency.

## Consequences

Migrating existing eval-spec-maker Markdown may require splitting repeated `#` sections into input files. Differences from eval-spec-maker output are not defects when they follow this v1 contract. Full compatibility, repeated heading categories, and localized column headers are out of scope for v1.

<details>
<summary>日本語</summary>

# ADR-0002: eval-spec-makerとの互換境界を定義する

- 状態: Accepted
- 日付: 2026-09-15

## コンテキスト

casemdのMarkdown入力は、[`ryuta46/eval-spec-maker`](https://github.com/ryuta46/eval-spec-maker) が公開した点検表の記法に着想を得ています。参考プロダクトをビルドして同じ入力を変換した結果、見出しの意味、シート名、列見出し、Markdownの解釈、XLSXレイアウトに差分がありました。

eval-spec-makerは `#` 見出しごとにシートを作成しますが、casemdは `#` を任意のドキュメントタイトルとして扱い、入力ファイルごとに1枚のシートを作成します。列名とレイアウトも異なります。

## 決定

casemd v1はeval-spec-makerとの完全な入出力互換を提供しません。互換の基盤は、共通する `##`、`###`、`####`、番号付きリスト、タスクリストの構造に限定します。

- `#` はドキュメントタイトルとして最大1回だけ許可します。
- 複数シートは `--input` を複数回指定して作成します。
- シート名は入力ファイル名から導出します。
- casemdの列見出しとXLSXスタイルを人間向け実施記録の契約とします。
- 番号付きリストとタスクリストの記号以降の文字列は入力どおり保持します。
- タスクリストの各行は、実施記録列を持つXLSXの個別行へ展開します。
- インラインコードはスプレッドシートのリッチテキストとして表示し、引用と全角文字・半角文字は入力どおり保持します。
- eval-spec-makerのJavaやGradle環境はcasemdのビルド、テスト、CIの依存関係にしません。

## 結果

既存のeval-spec-maker Markdownを移行する場合、複数の `#` セクションを入力ファイルへ分割することがあります。このv1契約に従う出力差分は不具合ではありません。完全互換、見出しの多重カテゴリ、列見出しのローカライズはv1の対象外です。

</details>
