# Release Process

This document describes the process for releasing new versions of `wt`.

## Overview

`wt` uses GoReleaser for automated releases. The release process is triggered by pushing git tags to GitHub. GitHub Actions builds binaries for multiple platforms and creates a GitHub release automatically.

## Prerequisites

- Write access to `github.com/trungung/wt`
- GoReleaser configured (`.goreleaser.yaml`)
- GitHub Actions workflow configured (`.github/workflows/release.yml`)

## Pre-Release Checklist

Before creating a release, verify:

### 1. Version Update

Update `version` variable in `cmd/wt/main.go`:

```go
var version = "0.0.1"  // Update this
```

### 2. Documentation

- [ ] README.md is up-to-date
- [ ] CHANGELOG.md has new version entry with date
- [ ] User documentation is current (`docs/user/`)
- [ ] AI documentation is updated (`docs/ai/llms.txt`)
- [ ] No broken documentation links

### 3. Configuration Files

- [ ] `.goreleaser.yaml` is correct
- [ ] `.github/workflows/release.yml` is correct

### 4. Testing

Run comprehensive tests:

```bash
# Run all tests
go test ./...

# Build locally
go build ./cmd/wt

# Test version output
./wt --version  # Should match version in code

# Test health check
./wt health

# Smoke test commands
./wt init --yes
./wt test-branch
./wt remove test-branch
```

### 5. Clean Working Directory

```bash
git status  # Should be clean
```

## Release Steps

### Step 1: Commit Changes

Commit all release-related changes:

```bash
git add .
git commit -m "chore: prepare v0.0.1 release

- Update version to 0.0.1
- Add comprehensive user documentation
- Add LICENSE and CHANGELOG
- Update README with new structure"
```

### Step 2: Push Changes

```bash
git push origin feat/docs-ai1
```

### Step 3: Create and Push Tag

Create annotated tag (recommended):

```bash
git tag -a v0.0.1 -m "Release v0.0.1

## Added
- Initial release with full feature set
- Comprehensive user documentation
- MIT license
- AI-friendly documentation (llms.txt)

## Changed
- Branch sanitization is now strict

## Security
- File locking for concurrent safety
- Dirty worktree protection"
```

Push tag to GitHub:

```bash
git push origin v0.0.1
```

### Step 4: Wait for GitHub Actions

GitHub Actions automatically:

1. Detects tag push
2. Runs GoReleaser workflow
3. Builds binaries for:
   - darwin/arm64 (macOS Apple Silicon)
   - darwin/amd64 (macOS Intel)
   - linux/arm64
   - linux/amd64
4. Creates tar.gz archives (or .zip for Windows)
5. Generates checksums.txt
6. Creates GitHub release
7. Uploads artifacts

Monitor progress: `https://github.com/trungung/wt/actions`

**Typical build time:** 2-5 minutes

### Step 5: Verify Release

After workflow completes:

1. Visit releases page: `https://github.com/trungung/wt/releases`
2. Verify release notes are correct
3. Download and test one binary:

```bash
# Download for your platform
curl -L https://github.com/trungung/wt/releases/download/v0.0.1/wt_Darwin_arm64.tar.gz -o wt.tar.gz
tar xzf wt.tar.gz
./wt --version  # Should show "0.0.1"
```

## Post-Release

### Update Documentation (if needed)

If release went smoothly, update any references in docs to point to new version.

### Announce Release

- Update README "Version" section
- Post on social media (optional)
- Create GitHub Discussions announcement (optional)

## Semantic Versioning

Follow [SemVer 2.0.0](https://semver.org/spec/v2.0.0.html):

- **MAJOR**: Incompatible API changes (unlikely for CLI)
- **MINOR**: New features, backwards-compatible
- **PATCH**: Bug fixes, backwards-compatible

**Examples:**

- `0.0.1` → `0.0.2`: Bug fixes
- `0.0.2` → `0.1.0`: New features
- `0.1.0` → `1.0.0`: Major breaking changes

## Dry-Run Release (Testing)

Test GoReleaser locally before actual release:

```bash
# Install GoReleaser
brew install goreleaser

# Dry-run (no GitHub push, just build)
goreleaser release --snapshot --clean
```

This creates binaries locally in `dist/` directory for testing.

## Troubleshooting

### GitHub Actions Fails

Check workflow logs:

1. Go to Actions tab
2. Click on failed workflow run
3. Check error message

**Common issues:**

- Build failure: Fix code, create new commit, delete old tag, create new tag
- GoReleaser config error: Fix `.goreleaser.yaml`

### Need to Update Release

If release needs changes:

1. Fix code/docs
2. Commit and push
3. Delete old tag: `git tag -d v0.0.1`
4. Delete remote tag: `git push origin :refs/tags/v0.0.1`
5. Create new tag and push

### Wrong Version in Binary

If version output doesn't match tag:

1. Check `cmd/wt/main.go` `version` variable
2. Ensure it matches tag exactly (no "v" prefix in code)
3. Re-tag and release

## Automated Release Workflow

The `.github/workflows/release.yml` workflow:

- Triggers on: `push tags: 'v*'`
- Runs GoReleaser: `release --clean`
- Configured in: `.goreleaser.yaml`

**Platform targets:**

- macOS: arm64, amd64
- Linux: arm64, amd64

**Archive formats:**

- tar.gz (macOS, Linux)
- zip (Windows, if enabled)

## Next Steps

After v0.0.1 release:

1. Monitor user feedback via GitHub Issues
2. Collect feature requests
3. Plan v0.0.2 or v0.1.0 based on feedback
4. Update CHANGELOG with `[Unreleased]` section

## See Also

- [GoReleaser Documentation](https://goreleaser.com/)
- [CHANGELOG](../../CHANGELOG.md) - Version history
- [GitHub Actions](https://github.com/trungung/wt/actions) - CI/CD runs
