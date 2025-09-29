# casemd

[![CI](https://github.com/9renpoto/casemd/actions/workflows/ci.yml/badge.svg)](https://github.com/9renpoto/casemd/actions/workflows/ci.yml)
[![codecov](https://codecov.io/gh/9renpoto/casemd/graph/badge.svg?token=D63wbdaCah)](https://codecov.io/gh/9renpoto/casemd)

CLI tool for converting structured Markdown inspection checklists into spreadsheet-friendly CSV files.

## Requirements

- Go 1.22+
- Node.js 20+ (for Lefthook-managed spell checking)
- Docker (for secretlint via container)

Using the provided devcontainer guarantees all dependencies are available. The workflow draws inspiration from [`ryuta46/eval-spec-maker`](https://github.com/ryuta46/eval-spec-maker), which popularized the inspection-sheet markdown format this tool consumes.
<!-- cspell:word ryuta -->

## Quick Start

```sh
# Run the CLI help
go run ./cmd/casemd --help

# Convert a Markdown inspection sheet into a CSV file
go run ./cmd/casemd --input notes.md --output build/notes.csv
```

The generated CSV contains predefined columns (Major Item, Medium Item, Minor Item, Validation Steps, Checkpoints, Result, Test Date, Tester, Notes) populated from the Markdown hierarchy and list content.

## Input Format

Markdown files should express each inspection case with nested headings for the hierarchy and lists for the execution details:

- `#` Heading — Optional document title (`Category` in the legacy format).
- `##` Heading — Major Item; starts a new block of related checks.
- `###` Heading — Medium Item inside the current major item.
- `####` Heading — Minor Item that becomes a single CSV row.
- Numbered list (`1.`) — Ordered validation steps captured verbatim in the `Validation Steps` column (line breaks preserved).
- Task list (`* [ ]`) — Checkpoints collected in the `Checkpoints` column (line breaks preserved, `[ ]` or `[x]` kept).

Extended example:

```markdown
# Inspection Sheet

## Setup
### Environment
#### Dependencies

1. Install required packages
2. Confirm default configurations
* [ ] Packages installed successfully
* [ ] Defaults match specification

#### Environment variables

1. Validate required environment variables are set
* [ ] Variables align with deployment checklist

### Configuration
#### CLI defaults

1. Inspect generated CSV path
* [ ] Output file lands in build/
* [ ] Delimiter is comma

## Execution
### Workflow
#### CLI run

1. Run casemd with sample.md
* [ ] Exit code is 0
* [ ] CSV file exists

#### Post-run cleanup

1. Remove temporary files from build/
* [ ] No leftover artifacts

### Validation
#### Error handling

1. Run casemd without --input
* [ ] CLI prints actionable error
* [ ] Exit code is 1
```

### Input → Output Mapping

| Markdown Element | CSV Column | Notes |
| --- | --- | --- |
| `##` Major Item heading | Major Item | Repeated for each `####` descendant; blank rows in the preview mimic merged cells. |
| `###` Medium Item heading | Medium Item | Repeated for each minor item inside the medium item; duplicate cells appear blank in the preview to mimic merged headings. |
| `####` Minor Item heading | Minor Item | Identifies the granular check represented by the row. |
| Ordered list under the minor item | Validation Steps | Joined with newlines, preserving list order. |
| Task list under the minor item | Checkpoints | Joined with newlines, retaining `[ ]` / `[x]` markers. |
| (No Markdown source) | Result / Test Date / Tester / Notes | Columns left blank for test execution; teams fill them in after loading the CSV. |

## Output Preview

Using the Markdown example above, the CLI produces a CSV that spreadsheet tools render as an inspection table. When a major or medium item covers multiple checks, the CSV contains the repeated heading value; the preview below intentionally leaves duplicate cells blank to hint at the grouping you see after opening the file in spreadsheet software.

| Major Item | Medium Item | Minor Item | Validation Steps | Checkpoints | Result | Test Date | Tester | Notes |
| --- | --- | --- | --- | --- | --- | --- | --- | --- |
| Setup | Environment | Dependencies | Install required packages<br>Confirm default configurations | Packages installed successfully<br>Defaults match specification |  |  |  |  |
|  |  | Environment variables | Validate required environment variables are set | Variables align with deployment checklist |  |  |  |  |
|  | Configuration | CLI defaults | Inspect generated CSV path | Output file lands in build/<br>Delimiter is comma |  |  |  |  |
| Execution | Workflow | CLI run | Run casemd with sample.md | Exit code is 0<br>CSV file exists |  |  |  |  |
|  |  | Post-run cleanup | Remove temporary files from build/ | No leftover artifacts |  |  |  |  |
|  | Validation | Error handling | Run casemd without --input | CLI prints actionable error<br>Exit code is 1 |  |  |  |  |

## Development Workflow

```sh
# Install git hooks
lefthook install

# Execute the local checks before opening a PR
lefthook run pre-commit

# Build and test
go build ./cmd/casemd
go test ./...
```

Keep documentation up to date as the clean-architecture layers evolve. Application orchestration lives in `internal/app`, interface adapters reside under `internal/interfaces`, and domain parsing logic sits in `internal/core`.
