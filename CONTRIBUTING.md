# Contributing to wt

Thank you for your interest in contributing to `wt`!

## Development Setup

1. **Prerequisites**
   - Go 1.25 or higher
   - Git installed on your system

2. **Clone and build**

   ```bash
   git clone https://github.com/trungung/wt.git
   cd wt
   go build ./cmd/wt
   ```

3. **Run tests**

   ```bash
   make test
   # or
   go test -v ./...
   ```

4. **Run linting**

   ```bash
   make lint
   # or run specific linters
   make lint-go  # golangci-lint
   make lint-md  # markdownlint
   ```

## Making Changes

1. **Create a branch**

   ```bash
   git checkout -b feature/your-feature-name
   ```

2. **Make your changes**
   - Follow code style in `AGENTS.md`
   - Write tests for new functionality
   - Update documentation if needed

3. **Run pre-commit checks**

   ```bash
   make ci-check  # runs lint, build, and test
   ```

## Pull Requests

Before submitting a PR:

- [ ] Run `make ci-check` and ensure it passes
- [ ] Write tests for new functionality
- [ ] Update relevant documentation (`docs/user/`)
- [ ] Update `CHANGELOG.md` for user-facing changes
- [ ] Include `wt health` output if relevant to change

## Questions or Issues?

- Open a [GitHub Issue](https://github.com/trungung/wt/issues) for bugs or feature requests
- Check existing issues before creating new ones
- See [AGENTS.md](AGENTS.md) for coding guidelines

## Testing

Run specific tests:

```bash
# Run all tests
go test -v ./...

# Run specific package
go test -v ./internal/core/...

# Run specific test
go test -v ./... -run TestIntegration
```

## Release Process

See [Release Process](docs/developer/release-process.md) for details on how releases are made.
