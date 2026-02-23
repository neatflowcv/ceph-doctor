# Repository Guidelines

## Project Structure & Module Organization

This repository is currently minimal and organized around a Go module
defined in `go.mod`. The expected layout is:

- `cmd/`: CLI entrypoints (currently empty, intended for `main` packages).
- `docs/adr/`: architecture decisions and technical choices. File names
  should be English kebab-case (e.g., `cli-library-kong.md`) and content
  should be written in Korean.

If you introduce new packages, keep them in top-level folders named for
their responsibility (e.g., `internal/`, `pkg/`) and keep `cmd/` for
binaries only.

## Build, Test, and Development Commands

Use standard Go tooling:

- `go build ./...`: compile all packages in the module.
- `go run ./cmd/<app>`: run a specific CLI (once a `cmd/<app>` exists).
- `go test ./...`: run all tests (there are no tests yet, so this will be
  a no-op until tests are added).
- `golangci-lint run --fix`: after changing any Go files, run this to lint
  and auto-fix.
- `bunx -bun markdownlint-cli2 --fix "docs/**/*.md"`: after changing any
  Markdown files, run this to lint and auto-fix.

## Coding Style & Naming Conventions

Follow standard Go conventions:

- Format with `gofmt` (tabs for indentation, standard Go style).
- Package names should be short, lowercase, and without underscores.
- File names should be lowercase and use hyphens only when required by
  documentation rules (see `docs/` rule above).
- Avoid global variables in CLI wiring; prefer passing configuration
  structs explicitly.

## Testing Guidelines

No testing framework is configured yet. When adding tests:

- Use Go’s built-in `testing` package.
- Name test files `*_test.go` and test functions `TestXxx`.
- Keep tests close to the code they validate.
- In tests, prefer `t.Context()` over `context.Background()`.

## Commit & Pull Request Guidelines

Commit Conventions:

- Use `<type>: 한글 설명` for commit messages.
- Use `docs: 한글 설명` for ADR changes, and commit each ADR separately.
- Split commits logically by change intent or scope.
- Use `feat` for functional changes, and `chore` for tooling/config/maintenance.

PRs should include a concise description, rationale, and any relevant ADR
references in `docs/adr/`.

## Documentation & ADRs

Record significant architectural choices in `docs/adr/`.
Keep ADRs short and actionable; include context, decision, and consequences.
