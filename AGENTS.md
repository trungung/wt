# AGENTS.md

This file provides guidelines for coding agents working on this Go-based git worktree management tool.

## Build, Lint, and Test Commands

### Essential Commands

- `make build` - Build the wt binary
- `make test` - Run all tests (`go test -v ./...`)
- `make lint` - Run all linters (Go + Markdown)
- `make lint-go` - Run golangci-lint
- `make lint-md` - Run markdownlint
- `make format-md` - Fix Markdown linting issues
- `make help` - Show all available make targets

### Running Tests

- Run all tests: `go test -v ./...`
- Run single test: `go test -v ./... -run TestIntegration`
- Run specific subtest: `go test -v ./test/... -run "TestIntegration/Default branch"`
- Test specific package: `go test -v ./internal/core/...`

## Code Style Guidelines

### Project Structure

- `cmd/wt/main.go` - CLI entry point (cobra commands)
- `internal/core/` - Core business logic
- `internal/git/` - Git operations wrapper
- `internal/config/` - Configuration handling
- `internal/ui/` - Interactive prompts (tap library)
- `test/` - Integration and concurrency tests

### Import Order

1. Standard library (blank line)
2. Third-party packages (blank line)
3. Internal packages (no blank line between internal packages)

Example:

```go
import (
    "fmt"
    "os"

    "github.com/spf13/cobra"

    "github.com/trungung/wt/internal/git"
)
```

### Naming Conventions

- **Packages**: lowercase, single word (core, config, git, ui)
- **Exported functions/types**: PascalCase (EnsureWorktree, Config)
- **Unexported functions/types**: camelCase (checkCollisions, run)
- **Variables**: camelCase (worktreePath, defaultBranch)
- **Constants**: PascalCase if exported

### Error Handling

- Wrap errors with context: `return fmt.Errorf("failed to load config: %w", err)`
- Use error types for complex scenarios (e.g., `RollbackError` in core/core.go:101)
- Use `errors.As` for type checking errors
- Always check and return errors, never ignore them
- Use named returns only when necessary for clarity

### Function Design

- Keep functions small and focused (single responsibility)
- Prefer returning errors over panicking
- Use closure callbacks for confirmations and interactions

### Testing Patterns

- Use table-driven tests for multiple scenarios
- Organize tests with `t.Run()` for subtests
- Setup temp directories with `os.MkdirTemp` and cleanup with `defer os.RemoveAll`
- For integration tests: build binary to temp dir, run it via exec.Command
- Helper functions: `runGit(t, dir, args...)` for git commands
- Test both success and failure paths
- Verify state changes, not just return values

### Configuration

- JSON-based config stored at repo root as `.wt.config.json`
- Use atomic writes: write to temp file, then rename
- Provide sensible defaults for optional fields
- Validate configuration and return clear error messages

### Git Operations

- Wrap git commands in internal/git package
- Use `git worktree list --porcelain` for parsing
- Always clean paths with `filepath.Clean()` and `filepath.EvalSymlinks`
- Handle both local and remote branch existence

### Concurrency Safety

- Use file locking via `git.AcquireLock(root, timeout)` before modifications
- Always defer unlock: `defer func() { _ = unlock() }()`
- Lock timeout: 5 seconds is standard

### CLI Design

- Use cobra for command structure
- Commands: `wt` (list), `wt <branch>` (ensure), `wt exec`, `wt remove`, `wt prune`, `wt init`, `wt health`
- Flags: use short flags with `StringVarP`, `BoolVarP`, etc.
- Return errors from RunE handlers, don't call os.Exit directly in handlers

### Comments

- Godoc comments for all exported types/functions (starting with the name)
- Package comments at top of files
- No inline comments for obvious code
- Use TODO sparingly, prefer creating issues

### Debug Logging

- Set `WT_DEBUG=1` environment variable to enable git command tracing
- Debug function in internal/git/debug.go logs command execution time

### Markdown Documentation

- Lint with `npx markdownlint-cli2 "**/*.md" "#node_modules"`
- Use `.markdownlint-cli2.jsonc` for configuration
- Disable MD013 (line length), MD033 (inline HTML), MD041 (first line heading)

### Performance

- Minimize git calls by caching results when appropriate
- Use `filepath.Clean()` to normalize paths before comparisons
- Avoid unnecessary file system operations

### Platform Compatibility

- Tested on macOS and Linux
- Use `filepath.EvalSymlinks()` for temp directories (important on macOS)
- Hardcoded permissions: 0755 for directories, 0644 for files

### CI/CD

- Go 1.25.5 is the target version
- GitHub Actions CI runs: lint-md, golangci-lint, build, test
- Ensure all tests pass before pushing

## Common Patterns

### Running a git command

```go
func run(dir string, args ...string) ([]byte, error) {
    cmd := exec.Command("git", args...)
    cmd.Dir = dir
    out, err := cmd.Output()
    return out, err
}
```

### Writing config atomically

```go
tempFile := configPath + ".tmp"
if err := os.WriteFile(tempFile, data, 0644); err != nil {
    return err
}
if err := os.Rename(tempFile, configPath); err != nil {
    _ = os.Remove(tempFile)
    return err
}
```

### Locking pattern

```go
unlock, err := git.AcquireLock(root, 5*time.Second)
if err != nil {
    return err
}
defer func() { _ = unlock() }()
// ... perform modifications
```
