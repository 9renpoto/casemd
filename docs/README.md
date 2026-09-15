# Documentation

This directory contains the project's durable design decisions and shared terminology.

## Structure

- [`adr/`](adr/) contains Architecture Decision Records.
- [`glossary.md`](glossary.md) defines terms used across the product and its documentation.
- `guides/` contains task-focused documentation for people using casemd.
- The repository README is the project entry point and links here for detailed documentation.
- This directory is the source for the published documentation site.

ADR files use a four-digit sequence and a kebab-case decision title: `NNNN-decision-title.md`. English is the primary document language. Japanese translations are kept in a `<details>` block at the end of the same document.

Issue and Pull Request descriptions follow the same convention: English is visible by default, and the Japanese translation is placed in a `<details>` block.

When a decision changes, update the ADR status and add a new ADR for a material replacement. Do not silently rewrite an accepted decision to describe a different architecture.

<details>
<summary>日本語</summary>

# ドキュメント

このディレクトリには、プロジェクトの永続的な設計判断と共通用語を収録します。

## 構成

- [`adr/`](adr/) にArchitecture Decision Recordを置きます。
- [`glossary.md`](glossary.md) にプロダクトと文書で使う用語を定義します。
- `guides/` に casemd を使う人向けのタスク中心の文書を置きます。
- リポジトリの README はプロジェクトの入口とし、詳細な文書はこのディレクトリへリンクします。
- このディレクトリを公開ドキュメントサイトのソースとします。

ADRは4桁の連番とkebab-caseの決定タイトルを使います。英語を標準表示の本文とし、日本語訳は同じ文書の末尾にある `<details>` ブロックへ置きます。

IssueとPull Requestの説明も同じ規約とし、英語を標準表示、日本語訳を `<details>` ブロックに置きます。

決定を変更する場合はADRの状態を更新し、重要な置き換えには新しいADRを追加します。異なるアーキテクチャを説明するためにAcceptedの決定を黙って書き換えません。

</details>
