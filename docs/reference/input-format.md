# Markdown Input Format

casemd extracts inspection cases from a line-oriented subset of Markdown.
It does not render a complete Markdown document.

## Structure and validation

| Notation | Meaning | Requirement |
| --- | --- | --- |
| `# Title` | Document title | Optional; at most once, before the hierarchy |
| `## Major` | Major item | Must contain a medium item |
| `### Medium` | Medium item | Requires a major item and must contain a case |
| `#### Case` | Inspection case | Requires both parent levels, a step, and a checkpoint |
| `1. Action` | Ordered step | At least one nonempty step per case |
| `* [ ] Expected result` | Unchecked checkpoint | At least one nonempty checkpoint per case |
| `* [x] Expected result` | Checked checkpoint | Also accepted as a checkpoint |

Heading markers must start at the beginning of the line.
Empty headings and levels 5 and 6 produce diagnostics.
List lines may have surrounding whitespace.
Ordered steps use digits followed by a period and whitespace; checkpoints use an asterisk followed by whitespace, `[ ]` or `[x]`, and whitespace.
Other list markers, including `- [ ]`, are not recognized as checkpoints.

Run `casemd validate --input notes.md` to check a specification without producing artifacts.
Conversion also checks all inputs before generating output.

## Text handling

- Heading text is trimmed and the optional document title is not exported.
- Ordered-step numbers and surrounding whitespace are removed; steps keep their source order rather than their original numbering.
- Checkpoint lines retain their task-list markers, with surrounding whitespace removed.
- Ordinary paragraphs and nonmatching lines are not exported.
  Put information needed during execution in a recognized step or checkpoint.
- Each checkpoint becomes one output row with the case headings, all steps separated by line breaks, and blank execution-record columns.
- Matched backtick code spans lose their delimiters in CSV, XLSX, and Google Sheets.
  XLSX and Google Sheets also display the code text in a monospace font.
  Unmatched backtick delimiters remain literal, and a backslash immediately before a backtick escapes that backtick.

For example:

```markdown
## Setup
### Environment
#### Dependencies
Context in an ordinary paragraph is not exported.

3. Install required packages
7. Confirm default configurations
* [ ] Packages installed successfully
* [x] Defaults match specification
```

This produces two rows with the same two steps, without the numbers `3.` and `7.`.
The checkpoint cells contain `* [ ] Packages installed successfully` and `* [x] Defaults match specification` respectively.
The checked marker is part of the scenario text; it does not populate the `Result` column.

See the [scenario workflow guide](../guides/v1-scenario-workflow.md) for validation, workbook generation, and execution-record instructions.

<details>
<summary>日本語</summary>

# Markdown 入力形式

casemd は Markdown の行単位の部分集合から点検ケースを抽出します。
Markdown 文書全体を表示するものではありません。

## 構造と検証

| 記法 | 意味 | 条件 |
| --- | --- | --- |
| `# Title` | 文書タイトル | 任意。階層の前に最大1回 |
| `## Major` | 大項目 | 中項目が必要 |
| `### Medium` | 中項目 | 親の大項目と子のケースが必要 |
| `#### Case` | 点検ケース | 両親階層、手順、チェックポイントが必要 |
| `1. Action` | 番号付き手順 | 各ケースに空でない手順が1件以上必要 |
| `* [ ] Expected result` | 未チェックのチェックポイント | 各ケースに空でないチェックポイントが1件以上必要 |
| `* [x] Expected result` | チェック済みのチェックポイント | チェックポイントとして受け付ける |

見出し記号は行頭に置きます。
空の見出しとレベル5・6の見出しは診断対象です。
リスト行の前後には空白を置けます。
手順は数字、ピリオド、空白の順で記述し、チェックポイントはアスタリスク、空白、`[ ]` または `[x]`、空白の順で記述します。
`- [ ]` などの別のリスト記号はチェックポイントとして認識しません。

`casemd validate --input notes.md` で成果物を生成せず検証できます。
変換時にも出力生成前にすべての入力を検証します。

## 文字列の扱い

- 見出し本文の前後空白は除去し、文書タイトルは出力しません。
- 手順の番号と前後空白は除去し、元の番号ではなく記述順を保持します。
- チェックポイントはタスクリスト記号を保持し、行の前後空白を除去します。
- 通常の段落や認識対象外の行は出力しません。
  実施時に必要な情報は、認識対象の手順またはチェックポイントに含めてください。
- チェックポイントごとに1行を生成し、ケースの見出し、改行で区切った全手順、空の実施記録列を付けます。
- 対応するバッククォートで囲んだコードは、CSV、XLSX、Google Sheets ですべて区切り記号を除去します。
  XLSX と Google Sheets ではコード本文を等幅フォントで表示します。
  閉じていないバッククォートはそのまま残し、直前のバックスラッシュはバッククォートのエスケープとして扱います。

上記の例では、`3.` と `7.` を除去した同じ2件の手順を持つ、2行の出力を生成します。
チェックポイント列にはそれぞれ `* [ ] Packages installed successfully` と `* [x] Defaults match specification` が入ります。
チェック済み記号はシナリオ本文の一部であり、`Result` 列の値にはなりません。

検証、ブック生成、実施記録の手順は[シナリオワークフローガイド](../guides/v1-scenario-workflow.md)を参照してください。

</details>
