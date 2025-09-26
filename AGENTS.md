# Repository Guidelines

## Project Structure & Module Organization
Runtime code lives under `cmd/casemd/` for the CLI entrypoint, `internal/app/` for use-case orchestration, `internal/core/` for domain rules (e.g., Markdown parsing), and `internal/interfaces/` for adapters such as the CLI layer. Add new adapters under `internal/interfaces/<name>` and keep fixtures beside packages in `testdata/`. Markdown docs and policy files remain in the repository root.

## Build, Test, and Development Commands
Open the project inside the devcontainer to pick up Go 1.22, Node.js 20, and the spell/secret scanners. Run `lefthook install` once after cloning, then `lefthook run pre-commit` to execute all mandatory checks. Use `go build ./cmd/casemd` to verify the CLI compiles, `go test ./... -cover` for local feedback, `go vet ./...` for static analysis, and `npx cspell "**/*.md"` to validate documentation. Mirror the secret scan outside the devcontainer with `docker run --rm -v "$(pwd):/work" --workdir /work secretlint/secretlint:latest secretlint .`. CI stores coverage artifacts and uploads them to Codecov; add a repository secret named `CODECOV_TOKEN` before running the workflow.

## Coding Style & Naming Conventions
Always format Go code with `gofmt`; the pre-commit hook rewrites staged `.go` files automatically. Keep packages small and focused, favor dependency injection to protect clean-architecture boundaries, and avoid cross-layer imports from inner to outer layers. Use snake_case for package names (`markdownparser`) and PascalCase for exported types, while CLI flags remain kebab-case (e.g., `--input`). Markdown files use one-sentence-per-line formatting and American English spellings to satisfy `cspell`.

## Testing Guidelines
Co-locate tests with their packages (`internal/core/parser/headings_test.go`). Prefer table-driven cases for parsing and conversion logic, assert on observable CSV output, and extend coverage for each bug fix. Capture filesystem behavior behind interfaces or use temporary directories in tests. Document intentional gaps in PR descriptions and tag TODOs in code with owner initials.

## Commit & Pull Request Guidelines
Write imperative, present-tense commit subjects under 72 characters (`Add heading parser`). Squash micro-fixes before review and keep unrelated changes out of the same PR. Each pull request should describe the motivation, outline architectural impacts, and link tracking issues. Attach terminal output or screenshots when behavior changes. Confirm `lefthook run pre-commit` succeeds locally to avoid CI churn.

## Security & Configuration Tips
Do not commit credentials; secretlint runs in CI and locally to catch tokens. Update `.devcontainer/Dockerfile` whenever toolchain versions shift so contributors get identical environments, and mirror those bumps in GitHub Actions to keep build parity.
