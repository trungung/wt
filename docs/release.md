## Minimal release plan (macOS + Linux) — GitHub Releases via GoReleaser

### Assumptions

- Repo is on GitHub: `github.com/<you>/wt`
- `wt` is a Go module and builds with `go build ./...`
- You want release artifacts for: `darwin/arm64`, `darwin/amd64`, `linux/arm64`, `linux/amd64`

---

## 1) Prep the repo (one-time)

1. Ensure module:
   ```sh
   go mod init github.com/<you>/wt
   go mod tidy
   ```
2. Make sure `wt --version` works (even a basic string is fine for now).
3. Add a license + README (GoReleaser/Homebrew will expect these later).

---

## 2) Add GoReleaser config (one-time)

Create `.goreleaser.yaml` at repo root (minimal):

```yaml
version: 2

before:
  hooks:
    - go mod tidy

builds:
  - id: wt
    main: ./cmd/wt # change to ./ if your main is at repo root
    binary: wt
    env:
      - CGO_ENABLED=0
    goos:
      - darwin
      - linux
    goarch:
      - amd64
      - arm64

archives:
  - formats: [tar.gz]
    name_template: "{{ .ProjectName }}_{{ .Version }}_{{ .Os }}_{{ .Arch }}"

checksum:
  name_template: "checksums.txt"

changelog:
  use: github
```

Notes:

- Set `main` correctly: `./` or `./cmd/wt` depending on your layout.
- `CGO_ENABLED=0` makes portable binaries (good default for CLIs).

---

## 3) Add GitHub Actions workflow (one-time)

Create `.github/workflows/release.yml`:

```yaml
name: release

on:
  push:
    tags:
      - "v*"

permissions:
  contents: write

jobs:
  goreleaser:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
        with:
          fetch-depth: 0

      - uses: actions/setup-go@v5
        with:
          go-version: "1.22"

      - uses: goreleaser/goreleaser-action@v6
        with:
          version: latest
          args: release --clean
        env:
          GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}
```

This publishes a GitHub Release automatically when you push a tag like `v0.1.0`.

---

## 4) Dry-run locally (recommended)

Install GoReleaser locally and test:

```sh
goreleaser release --snapshot --clean
```

This builds artifacts without publishing.

---

## 5) Cut a release

1. Commit everything:
   ```sh
   git add .
   git commit -m "chore: add goreleaser"
   git push
   ```
2. Tag and push:
   ```sh
   git tag v0.1.0
   git push origin v0.1.0
   ```

Result: GitHub Actions builds and uploads:

- `wt_0.1.0_darwin_amd64.tar.gz`
- `wt_0.1.0_darwin_arm64.tar.gz`
- `wt_0.1.0_linux_amd64.tar.gz`
- `wt_0.1.0_linux_arm64.tar.gz`
- `checksums.txt`

---

## 6) Tell users how to install (for now)

In README, provide the simplest manual install:

- Download from Releases, extract, move `wt` to a PATH directory:
  ```sh
  tar -xzf wt_0.1.0_darwin_arm64.tar.gz
  install -m 755 wt ~/bin/wt
  ```
