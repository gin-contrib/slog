# AGENTS Guide for gin-slog Repository

## Build, Lint, and Test Commands

- **Build:** `go build ./...`
- **Format:** `go fmt ./...`
- **Lint:** `go vet ./...`
- **Test:** `go test ./...`
- **Run single test:** `go test -run ^TestName$ ./...`

## Code Style Guidelines

- **Imports:**
  - Standard library imports, then third-party, then local, each in separate blocks.
  - Use Go `gofmt` for all formatting.
- **Types:**
  - Use clear types (e.g., `slog.Level`, not `int`).
  - Type aliases (e.g., `type Fn func(*gin.Context, *slog.Logger) *slog.Logger`) are used for middleware options.
- **Naming:**
  - Exported types, funcs: `CamelCase`. Private: `camelCase`.
  - Packages should remain lower_snake_case.
- **Error Handling:**
  - Return errors when possible; panic only for programmer errors, not user errors.
  - Use Go 1.20+ error wrapping.
- **Logging/Middleware:**
  - Use Go 1.23+ `log/slog`; never log sensitive data (see options for hidden headers).
- **Formatting:**
  - Enforce via `go fmt ./...`.
- **Functions:**
  - Small, focused, no side-effects outside intended context.

## Additional Info

- See README.md for configurable logging options.
- No Cursor/Copilot rules are present as of June 2025.

Keep the repo clear, idiomatic, and production-grade.
