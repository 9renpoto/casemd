# casemd

[![CI](https://github.com/9renpoto/casemd/actions/workflows/ci.yml/badge.svg)](https://github.com/9renpoto/casemd/actions/workflows/ci.yml)
[![codecov](https://codecov.io/gh/9renpoto/casemd/graph/badge.svg?token=D63wbdaCah)](https://codecov.io/gh/9renpoto/casemd)

CLI tool for converting Markdown content into spreadsheet-friendly formats. The initial milestone extracts top-level Markdown headings and writes them to a CSV file, paving the way for full spreadsheet creation.

## Requirements

- Go 1.22+
- Node.js 20+ (for Lefthook-managed spell checking)
- Docker (for secretlint via container)

Using the provided devcontainer guarantees all dependencies are available.

## Quick Start

```sh
# Run the CLI help
go run ./cmd/casemd --help

# Convert Markdown headings into a CSV file
go run ./cmd/casemd --input notes.md --output build/notes.csv
```

The generated CSV contains a header row (`Heading`) followed by each top-level heading discovered in the source Markdown.

## Output Preview

The CLI produces a tabular CSV that spreadsheet tools render as a table with distinct columns for inspection planning. When a major or medium item covers multiple checks (as in the reference screenshot), the CSV repeats the value on each relevant row; the table below leaves those cells blank on subsequent rows to hint at the visual grouping you get after opening the CSV in spreadsheet software.

| Major Item | Medium Item | Minor Item | Validation Steps | Checkpoints | Result | Test Date | Tester | Notes |
| --- | --- | --- | --- | --- | --- | --- | --- | --- |
| Setup | Environment | Dependencies | Install required packages | Packages installed successfully | Pass | 2024-03-15 | BRG | Verified with devcontainer |
|  |  | Environment variables | Validate required environment variables are set | Variables align with deployment checklists | Pass | 2024-03-15 | BRG | Documented in `.env.example` |
|  | Configuration | CLI defaults | Confirm default output path and delimiter | Defaults match specification | Pass | 2024-03-15 | BRG | No overrides applied |
| Execution | Workflow | CLI run | Execute `casemd --input sample.md --output sample.csv` | CLI exits with status 0 | Pass | 2024-03-16 | BRG | Sample markdown stored in `testdata/` |
|  |  | Post-run cleanup | Remove temporary files from `build/` | No leftover artifacts | Pass | 2024-03-16 | BRG | Cleaned via `make clean` |
|  | Validation | Error handling | Provide missing input file | CLI reports friendly error | Pass | 2024-03-17 | BRG | Exit code 1 verified |
| Verification | Output | CSV content | Open generated CSV and confirm columns | Columns match specification | Pass | 2024-03-16 | BRG | Ready for reviewer |

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
