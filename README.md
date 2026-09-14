# casemd

[![CI](https://github.com/9renpoto/casemd/actions/workflows/ci.yml/badge.svg)](https://github.com/9renpoto/casemd/actions/workflows/ci.yml)
[![codecov](https://codecov.io/gh/9renpoto/casemd/graph/badge.svg?token=D63wbdaCah)](https://codecov.io/gh/9renpoto/casemd)

CLI tool for converting structured Markdown inspection checklists into CSV files, multi-sheet spreadsheets, and Google Spreadsheets.

## Requirements

- Go 1.26+
- typos-cli

Install Go using the toolchain manager or package distribution appropriate for your development environment.
The provided devcontainer is also available with the supporting tools preinstalled.
The workflow draws inspiration from [`ryuta46/eval-spec-maker`](https://github.com/ryuta46/eval-spec-maker), which popularized the inspection-sheet markdown format this tool consumes.

## Installation

Install the latest released version with Homebrew:

```sh
brew install 9renpoto/tap/casemd
```

Verify the installed version:

```sh
casemd --version
```

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

## Quick Start

The repository ships with `notes.md` and `follow-up.md`, which replicate the extended example below so you can exercise the CLI immediately.

```sh
# Run the CLI help
go run ./cmd/casemd --help

# Convert Markdown inspection sheets into both CSV and XLSX outputs
go run ./cmd/casemd --input notes.md --csv-output build/notes.csv --spreadsheet-output build/notes.xlsx

# Create a Google Spreadsheet (requires GOOGLE_SHEETS_ACCESS_TOKEN with an OAuth token)
GOOGLE_SHEETS_ACCESS_TOKEN=ya29.example-token go run ./cmd/casemd --input notes.md --google-spreadsheet-title "Inspection Sheet Export"

# Generate only one of the output formats
go run ./cmd/casemd --input notes.md --csv-output build/notes.csv
go run ./cmd/casemd --input notes.md --input follow-up.md --spreadsheet-output build/all-notes.xlsx
```

The generated spreadsheet contains predefined columns (Major Item, Medium Item, Minor Item, Validation Steps, Checkpoints, Result, Test Date, Tester, Notes) populated from the Markdown hierarchy and list content.
Each Markdown file becomes its own sheet inside the workbook.
Passing `--google-spreadsheet-title` uploads the same structure to Google Sheets using the bearer token exposed through `GOOGLE_SHEETS_ACCESS_TOKEN`.

## Input Format

Markdown files should express each inspection case with nested headings for the hierarchy and lists for the execution details:

- `#` Heading — Optional document title (`Category` in the legacy format).
- `##` Heading — Major Item; starts a new block of related checks.
- `###` Heading — Medium Item inside the current major item.
- `####` Heading — Minor Item that becomes a single spreadsheet row.
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

1. Inspect generated spreadsheet path
* [ ] Output file lands in build/
* [ ] Delimiter is comma

## Execution
### Workflow
#### CLI run

1. Run casemd with sample.md
* [ ] Exit code is 0
* [ ] Spreadsheet file exists

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

| Markdown Element | Spreadsheet Column | Notes |
| --- | --- | --- |
| `##` Major Item heading | Major Item | Repeated for each `####` descendant; blank rows in the preview mimic merged cells. |
| `###` Medium Item heading | Medium Item | Repeated for each minor item inside the medium item; duplicate cells appear blank in the preview to mimic merged headings. |
| `####` Minor Item heading | Minor Item | Identifies the granular check represented by the row. |
| Ordered list under the minor item | Validation Steps | Joined with newlines, preserving list order. |
| Task list under the minor item | Checkpoints | Joined with newlines, retaining `[ ]` / `[x]` markers. |
| (No Markdown source) | Result / Test Date / Tester / Notes | Columns left blank for test execution; teams fill them in after loading the spreadsheet. |

## Output Preview

Using the Markdown example above, the CLI produces a spreadsheet that spreadsheet tools render as an inspection table. When a major or medium item covers multiple checks, the workbook contains the repeated heading value; the preview below intentionally leaves duplicate cells blank to hint at the grouping you see after opening the file in spreadsheet software.

| Major Item | Medium Item | Minor Item | Validation Steps | Checkpoints | Result | Test Date | Tester | Notes |
| --- | --- | --- | --- | --- | --- | --- | --- | --- |
| Setup | Environment | Dependencies | Install required packages<br>Confirm default configurations | Packages installed successfully<br>Defaults match specification |  |  |  |  |
|  |  | Environment variables | Validate required environment variables are set | Variables align with deployment checklist |  |  |  |  |
|  | Configuration | CLI defaults | Inspect generated spreadsheet path | Output file lands in build/<br>Delimiter is comma |  |  |  |  |
| Execution | Workflow | CLI run | Run casemd with sample.md | Exit code is 0<br>Spreadsheet file exists |  |  |  |  |
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

# Start the web UI
go run ./cmd/casemd serve

# In another terminal, verify the health endpoint
curl --fail http://localhost:3000/healthz
```

Keep documentation up to date as the clean-architecture layers evolve. Application orchestration lives in `internal/app`, interface adapters reside under `internal/interfaces`, and domain parsing logic sits in `internal/core`.
