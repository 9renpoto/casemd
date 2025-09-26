# casemd

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
