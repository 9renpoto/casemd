# ADR-0001: Markdown でシナリオテスト仕様を管理する / Manage Scenario Test Specifications in Markdown

- 状態: Accepted / Status: Accepted
- 日付: 2026-08-17 / Date: 2026-08-17

## コンテキスト / Context

casemd は、`eval-spec-maker` が示したアプローチを再実装し、テスト仕様の管理に関する課題を解決することを目指します。
casemd aims to solve test specification management problems by reimplementing the approach demonstrated by `eval-spec-maker`.

ユニットテストは個々の実装単位の振る舞いを継続的に検証するために有効です。
Unit tests are effective for continuously verifying the behavior of individual implementation units.

一方、複数の操作や境界を横断するシナリオ、業務上のユースケース、手動確認を含むテストケースは、ユニットテストだけでは表現および維持が困難です。
However, scenarios spanning multiple operations or boundaries, business use cases, and test cases involving manual verification are difficult to express and maintain through unit tests alone.

これらのテストケースには、人が読み書きしやすく、レビューでき、変更履歴を追跡できる管理形式が必要です。
These test cases need a management format that people can read and write, review, and track through version history.

## 決定要因 / Decision Drivers

- テストケースをソースコードと同じ変更管理プロセスで扱えること。
  Test cases can use the same change management process as source code.
- シナリオ、手順、確認項目の差分をレビューしやすいこと。
  Changes to scenarios, steps, and checkpoints are easy to review.
- 特定の表計算ソフトウェアへ依存せず編集できること。
  Authors can edit specifications without depending on a particular spreadsheet application.
- 人が読める形式と機械的に変換できる構造を両立すること。
  The format balances human readability with a structure that can be converted mechanically.
- テスト実施時に必要な形式へ出力できること。
  Specifications can be exported into formats needed during test execution.

## 決定 / Decision

シナリオおよびユースケースのテスト仕様は Markdown で記述し、Git で管理します。
Scenario and use-case test specifications will be authored in Markdown and managed with Git.

casemd は Markdown の構造を解釈し、テスト実施で利用する成果物へ変換します。
casemd will interpret the Markdown structure and convert it into artifacts used during test execution.

Markdown をテスト定義の正本とし、CSV、XLSX、Google Sheets は Markdown から生成する一方向の派生成果物とします。
Markdown is the source of truth for test definitions, while CSV, XLSX, and Google Sheets are one-way derived artifacts generated from Markdown.

テスト結果、実施日、担当者、備考などの実施記録は生成した成果物上で管理します。
Execution records such as results, execution dates, testers, and notes are managed in the generated artifacts.

生成した成果物の変更を Markdown へ逆同期する機能は提供しません。
The product will not provide reverse synchronization from generated artifacts to Markdown.

既存の Markdown 構造を v1 フォーマットとして維持し、互換契約にします。
The existing Markdown structure is retained as the v1 format and treated as a compatibility contract.

ユニットテストは引き続き実装単位の自動検証を担い、Markdown テスト仕様はユニットテストを補完します。
Unit tests will continue to provide automated verification at the implementation-unit level, while Markdown test specifications complement them.

## v1 フォーマット / v1 Format

- `#` は任意の文書タイトルを表します。
  `#` represents an optional document title.
- `##` は大項目を表します。
  `##` represents a major item.
- `###` は中項目を表します。
  `###` represents a medium item.
- `####` は一つのテストケースを表します。
  `####` represents one test case.
- 番号付きリストは順序を持つ実施手順を表します。
  Ordered lists represent ordered execution steps.
- タスクリストは確認項目または期待結果を表します。
  Task lists represent checkpoints or expected results.

ID、タグ、事前条件、優先度などのメタデータは v1 の必須項目に含めません。
Metadata such as IDs, tags, preconditions, and priorities is not required by v1.

## 検証 / Validation

成果物を生成せずに v1 フォーマットへの適合を確認する `casemd validate` コマンドを提供します。
The product provides a `casemd validate` command that checks conformance with the v1 format without generating artifacts.

検証に成功した場合は終了コード `0`、違反がある場合は非 `0` を返します。
Validation returns exit code `0` on success and a non-zero exit code when violations are present.

診断にはファイル名、行番号、違反内容、可能な場合は修正案を含めます。
Diagnostics include the file name, line number, violation, and a suggested correction when possible.

`validate` は複数の `--input` を受け付け、CSV または XLSX を生成しません。
`validate` accepts multiple `--input` values and does not generate CSV or XLSX output.

通常の変換処理も同じ検証ルールを使用し、不正な入力から成果物を生成しません。
Normal conversion uses the same validation rules and does not generate artifacts from invalid input.

## 実行モデル / Execution Model

v1 のシナリオテストは人間が実行することを前提とし、casemd はテスト担当者の理解、操作、記録の体験を最大化します。
Scenario tests in v1 are executed by people, and casemd prioritizes the tester experience of understanding, performing, and recording tests.

生成する成果物は、実施手順、確認項目、結果、実施日、担当者、備考を人が確認および入力しやすい構造にします。
Generated artifacts make execution steps, checkpoints, results, execution dates, testers, and notes easy for people to review and enter.

ブラウザ、API、CLI の自動操作、テストコード生成、自動実行結果の取り込み、CI 上でのシナリオ実行は v1 の対象外です。
Automated browser, API, or CLI operations, test code generation, automated result ingestion, and scenario execution in CI are outside the scope of v1.

自動実行が必要になった場合は、実行基盤ごとの要件を別 ADR で決定します。
If automated execution becomes necessary, its runner-specific requirements will be decided in a separate ADR.

人間向け実施体験の最初の検証対象は XLSX スプレッドシートとします。
The first validation target for the human execution experience is the XLSX spreadsheet.

CSV はデータ交換用の補助形式として維持し、Google Sheets の UX 改善は XLSX で得た知見を基に後続で検討します。
CSV remains a secondary data exchange format, while Google Sheets UX improvements are considered later using lessons learned from XLSX.

## Spreadsheet UX の評価基準 / Spreadsheet UX Evaluation Criteria

- 長いシナリオ、複数の実施手順、箇条書きの確認項目をセル内で読みやすく表示します。
  Long scenarios, multiple execution steps, and bulleted checkpoints remain readable within cells.
- 改行を保持し、テキストの折り返しと上揃えを使用します。
  Line breaks are preserved, with wrapped and top-aligned text.
- 見出しを識別しやすくし、スクロール中も列の意味を確認できるようにします。
  Headers remain distinguishable and column meanings remain visible while scrolling.
- 列幅と行の表示を、手順の理解と結果入力の両方に適した初期状態にします。
  Initial column widths and row presentation support both understanding steps and entering results.
- 結果、実施日、担当者、備考を入力しやすくします。
  Results, execution dates, testers, and notes are easy to enter.
- 長文と箇条書きを含む代表的なシナリオで、人による表示確認を行います。
  Human visual verification uses a representative scenario containing long text and bullet lists.

## 想定する対象 / Intended Scope

- 複数ステップからなる利用シナリオ。
  User scenarios consisting of multiple steps.
- 業務またはプロダクトのユースケースに基づくテストケース。
  Test cases based on business or product use cases.
- 人による確認項目を含む受け入れおよび回帰テスト。
  Acceptance and regression tests containing human verification checkpoints.
- 人が手順を実行し、結果と証跡を記録するテスト運用。
  Test operations in which people perform steps and record results and evidence.

## 結果 / Consequences

### 利点 / Benefits

- Pull Request 上でテスト仕様をレビューできます。
  Test specifications can be reviewed in Pull Requests.
- ソースコードとテスト仕様の変更履歴を関連付けられます。
  Source code and test specification histories can be related.
- Markdown を正規化された出力形式へ再利用できます。
  Markdown can be reused to produce normalized output formats.
- テスト定義と実施記録の責任を分離できます。
  Responsibilities for test definitions and execution records are separated.
- 既存の入力ファイルと変換処理の互換性を維持できます。
  Compatibility with existing input files and conversion behavior can be maintained.
- CI で成果物を生成せずにテスト仕様を検証できます。
  CI can validate test specifications without generating artifacts.
- 自動実行基盤の都合より、テスト担当者の可読性と操作性を優先できます。
  Tester readability and usability can take priority over automation-runner constraints.
- XLSX で人間向け UX を小さく検証してから、他の Spreadsheet サービスへ知見を展開できます。
  Human-focused UX can be validated in XLSX before applying the lessons to other spreadsheet services.

### トレードオフ / Trade-offs

- 機械変換のために Markdown の記述規則を定義し、検証する必要があります。
  Markdown authoring rules must be defined and validated for mechanical conversion.
- 表計算ソフトウェア上の自由な編集を、そのまま Markdown へ反映することはできません。
  Free-form edits made in spreadsheet applications cannot automatically be reflected in Markdown.
- テスト定義を変更した後に成果物を再生成する場合、既存の実施記録を自動的には引き継ぎません。
  Regenerating artifacts after changing test definitions does not automatically preserve existing execution records.
- 大規模なテスト仕様では、ファイル分割と参照方法の規約が必要になります。
  Large test specifications require conventions for file organization and references.
- v1 に含まれないメタデータは、互換性を考慮した後続の拡張として設計する必要があります。
  Metadata not included in v1 must be designed as a follow-up extension with compatibility in mind.
- 変換処理と独立コマンドで診断結果が一致するよう、検証ロジックを共有する必要があります。
  Validation logic must be shared so conversion and the standalone command produce consistent diagnostics.
- 自動テストランナーとの直接連携を必要とする利用者には、v1 だけでは対応できません。
  v1 does not serve users who require direct integration with automated test runners.
