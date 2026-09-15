# ADR-0001: Manage Scenario Test Specifications in Markdown

- Status: Accepted
- Date: 2026-08-17

## Context

Scenario tests cross multiple operations and boundaries, and often require manual verification. Unit tests cannot express or maintain those cases alone. The team needs a format that people can read, review, edit, and track in version control while still providing enough structure for mechanical conversion.

## Decision drivers

- Keep scenario specifications in the same change-management process as source code.
- Make changes to scenarios, steps, and checkpoints easy to review.
- Avoid a dependency on a particular spreadsheet application for authoring.
- Balance human readability with machine-readable structure.
- Export specifications into formats useful during test execution.

## Decision

Scenario and use-case test specifications are authored in Markdown and managed with Git. Markdown is the source of truth; CSV, XLSX, and Google Sheets are one-way derived artifacts.

casemd interprets the Markdown structure and converts it into artifacts used during test execution. Execution records such as results, execution dates, testers, and notes are recorded in the generated artifacts. Reverse synchronization from artifacts to Markdown is not provided.

The existing Markdown structure is retained as the v1 compatibility contract. Unit tests continue to verify implementation units, while Markdown scenario specifications complement them.

## v1 format

- `#` is an optional document title.
- `##` is a major item.
- `###` is a medium item.
- `####` is one test case.
- Ordered lists are ordered execution steps.
- Task lists are checkpoints or expected results.

IDs, tags, preconditions, and priorities are not required by v1.

## Validation and execution model

`casemd validate` checks v1 conformance without generating artifacts. Validation returns exit code `0` on success and a non-zero code when diagnostics are present. Diagnostics include the source, line number, violation, and, when possible, a suggested correction.

Normal conversion uses the same validation rules and never generates artifacts from invalid input. v1 assumes that people execute the scenarios and record the results. Automated browser, API, or CLI operations, test-code generation, result ingestion, and scenario execution in CI are outside this ADR.

The XLSX spreadsheet is the first human-execution target. CSV remains a secondary exchange format, and Google Sheets UX improvements follow the XLSX evaluation.

## Spreadsheet UX criteria

- Long scenarios, multiple steps, and checkpoint lists remain readable in cells.
- Line breaks are preserved with wrapped, top-aligned text.
- Headers remain identifiable while scrolling.
- Initial widths and row presentation support both reading and result entry.
- Result, test date, tester, and notes fields are easy to enter.
- Representative scenarios with long and mixed-width text are visually verified.

## Consequences

Pull Requests can review scenario specifications alongside code, and the same Markdown can produce normalized output formats. The definition and execution-record responsibilities remain separate.

Markdown authoring rules must be defined and validated. Spreadsheet edits cannot be synchronized back to Markdown, and regenerating artifacts does not automatically preserve existing execution records. Large specifications need file-organization conventions. Future metadata and automated runners require separate decisions.

<details>
<summary>日本語</summary>

# ADR-0001: Markdown でシナリオテスト仕様を管理する

- 状態: Accepted
- 日付: 2026-08-17

## コンテキスト

複数の操作や境界を横断し、手動確認を含むシナリオは、ユニットテストだけでは表現と維持が困難です。人が読み書きでき、レビューでき、Gitで履歴を追跡でき、機械的にも変換できる形式が必要です。

## 決定要因

- ソースコードと同じ変更管理プロセスで仕様を扱えること。
- シナリオ、手順、チェックポイントの差分をレビューしやすいこと。
- 特定の表計算ソフトウェアに依存せず編集できること。
- 可読性と機械変換可能な構造を両立すること。
- テスト実施に必要な形式へ出力できること。

## 決定

シナリオおよびユースケースのテスト仕様はMarkdownで記述し、Gitで管理します。Markdownを正本とし、CSV、XLSX、Google Sheetsは一方向の派生成果物とします。

casemdはMarkdownの構造を解釈して成果物へ変換します。結果、実施日、担当者、備考などの実施記録は生成物で管理し、生成物からMarkdownへの逆同期は提供しません。

既存のMarkdown構造をv1の互換契約として維持します。ユニットテストは実装単位を検証し、Markdownのシナリオ仕様はそれを補完します。

## v1フォーマット

- `#` は任意のドキュメントタイトルです。
- `##` は大項目です。
- `###` は中項目です。
- `####` は一つのテストケースです。
- 番号付きリストは順序を持つ実施手順です。
- タスクリストはチェックポイントまたは期待結果です。

ID、タグ、事前条件、優先度はv1の必須項目ではありません。

## 検証と実行モデル

`casemd validate` は成果物を生成せずv1への適合を確認します。成功時は終了コード `0`、診断がある場合は非 `0` を返します。通常の変換も同じ検証ルールを使い、不正な入力から成果物を生成しません。

v1は人がシナリオを実行して結果を記録することを前提とします。ブラウザ、API、CLIの自動操作、テストコード生成、結果取り込み、CI上でのシナリオ実行はこのADRの対象外です。

人間向け実施体験の最初の対象はXLSXです。CSVは補助的な交換形式とし、Google SheetsのUX改善はXLSXの評価後に検討します。

## スプレッドシートUXの評価基準

- 長いシナリオ、複数手順、チェックポイントをセル内で読みやすく表示すること。
- 改行を保持し、折り返しと上揃えを使うこと。
- スクロール中も列の意味を確認できること。
- 手順の理解と結果入力の両方に適した初期幅と行表示にすること。
- 結果、実施日、担当者、備考を入力しやすいこと。
- 長文と全角半角混在文字を含む代表シナリオで表示確認すること。

## 結果

Pull Requestで仕様をコードとともにレビューできます。一方、Markdownの記述規則を定義・検証する必要があります。表計算ソフトウェアでの編集はMarkdownへ同期できず、仕様変更後の再生成で実施記録を自動的には引き継ぎません。大規模仕様の分割規約、メタデータ、自動実行基盤は別途決定します。

</details>
