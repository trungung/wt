.PHONY: lint lint-go lint-md format-md test ci-check build release-check release-dry-run help

# Default target
all: lint build test

help: ## Display this help
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-15s\033[0m %s\n", $$1, $$2}'

lint: lint-go lint-md ## Run all linters

lint-go: ## Run Go linter
	golangci-lint run

lint-md: ## Run Markdown linter
	npx markdownlint-cli2 "**/*.md" "#node_modules"

format-md: ## Fix Markdown linting issues automatically
	npx markdownlint-cli2 --fix "**/*.md" "#node_modules"

test: ## Run tests
	go test -v ./...

ci-check: ## Run all CI checks (lint, build, test)
	@echo "Running CI checks..."
	$(MAKE) lint
	$(MAKE) build
	$(MAKE) test
	@echo "CI checks passed!"

build: ## Build the project
	go build -v ./cmd/wt

release-check: ## Simulate release pipeline to verify cross-platform builds
	goreleaser release --snapshot --clean

release-dry-run: ## Dry-run GoReleaser release locally
	goreleaser release --snapshot --clean
