# Glossary

## Source of truth

The authoritative representation of a test definition. In casemd, this is the Markdown file managed in Git.

## Derived artifact

An output generated from the source of truth. CSV, XLSX, and Google Sheets are derived artifacts and are not synchronized back to Markdown.

## Checkpoint

An individual task-list item representing a result to verify. Each checkpoint becomes its own spreadsheet row so execution records remain independent.

## Execution record

Information captured while performing a checkpoint: result, test date, tester, and notes.

## Rich Markdown

Markdown-like inline notation that carries presentation semantics into an output format. casemd currently renders matched backtick code spans as monospace text in XLSX and Google Sheets; unmatched delimiters remain literal.

<details>
<summary>日本語</summary>

# 用語集

## 正本 / Source of truth

テスト定義の正式な表現です。casemdではGitで管理するMarkdownファイルを指します。

## 派生成果物 / Derived artifact

正本から生成される出力です。CSV、XLSX、Google Sheetsは派生成果物であり、Markdownへ逆同期しません。

## チェックポイント / Checkpoint

確認する結果を表す個別のタスクリスト項目です。実施記録を独立させるため、各チェックポイントはスプレッドシートの1行になります。

## 実施記録 / Execution record

チェックポイントの実行時に記録する結果、実施日、担当者、備考です。

## Rich Markdown

出力形式へ表示上の意味を渡すMarkdownに近いインライン記法です。casemdは現在、対応するバッククォートのコードスパンをXLSXとGoogle Sheetsで等幅テキストとして表示し、閉じていない区切り記号はリテラルのまま出力します。

</details>
