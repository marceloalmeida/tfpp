# CLAUDE.md

## Project Overview

**tfpp** (Terraform Provider Packager) is a Go CLI tool that packages Terraform provider assets for use in static asset-based private Terraform registries. It takes GoReleaser build output and produces a directory structure compatible with the Terraform provider installation protocol, suitable for hosting on S3, GitHub Pages, or any static HTTP server.

Repository: `github.com/marceloalmeida/tfpp`

## Repository Structure

```
.
├── main.go              # All application code (single-file Go program)
├── main_test.go         # Unit tests
├── go.mod               # Go module definition (Go 1.24.2, no external dependencies)
├── .goreleaser.yml      # GoReleaser v2 config for cross-platform builds
├── .golangci.yml        # golangci-lint v2 configuration
├── .releaserc.json      # semantic-release config (conventional commits)
├── .github/workflows/
│   ├── test.yml         # CI: build, lint (golangci-lint), and test on PRs/pushes
│   ├── lint-pr.yml      # PR title validation (conventional commits)
│   └── create-release.yml  # Automated releases via semantic-release + GoReleaser
└── .vscode/launch.json  # VS Code debug configuration
```

This is a single-package Go application — all code lives in `main.go` with tests in `main_test.go`. There are no external Go dependencies.

## Common Commands

### Build
```bash
go build -v .
```

### Run Tests
```bash
go test -v -cover .
```

### Run Linter
```bash
golangci-lint run
```

### Run the Tool
```bash
go run . -p <provider-name> -r <repo-name> -ns <namespace> -d <domain> -gf <gpg-fingerprint> -v <version> [-dp <dist-path>] [-gk <gpg-pubkey-file>]
```

Required flags: `-p`, `-r`, `-ns`, `-d`, `-gf`, `-v`
Optional flags: `-dp` (default: `dist`), `-gk` (default: `pubkey.txt`)

## Architecture & Key Concepts

### How It Works
1. Fetches (or creates default) `.well-known/terraform.json` from the registry domain
2. Downloads existing `versions` file from the registry (if available) and merges the new version
3. Creates the Terraform provider registry directory structure under `release/`
4. Copies SHA256SUMS files and GPG signatures
5. Copies build zip files to OS-specific download directories
6. Creates per-architecture JSON metadata files with download URLs and signing keys

### Key Data Types
- `WellKnown` — Terraform registry service discovery (providers.v1, modules.v1 paths)
- `Versions` / `Version` / `Platform` — Provider version metadata
- `Architecture` — Per-OS/arch download metadata including URLs, checksums, and GPG signing keys

### Protocol Versions
- Version metadata uses protocols `["5.0", "5.1"]`
- Architecture files use protocols `["4.0", "5.1"]`

## Code Conventions

- **Single-file structure**: All code in `main.go`, all tests in `main_test.go`
- **No external dependencies**: Uses only Go standard library
- **Table-driven tests**: Tests use the `[]struct{ name; ...; wantErr }` pattern with `t.Run` subtests
- **Error handling**: Mix of `log.Fatal`/`log.Fatalf` (exits on error) and returned errors
- **JSON serialization**: Struct tags use `snake_case` field names

## CI/CD

### On Pull Requests and Pushes
- **Build job**: `go mod download` → `go build -v .` → `golangci-lint` (latest)
- **Test job** (after build): `go test -v -cover .` with `TF_ACC=1`

### PR Title Linting
- PR titles must follow [Conventional Commits](https://www.conventionalcommits.org/) format
- Enforced by `amannn/action-semantic-pull-request`

### Releases
- Automated via `semantic-release` on the `main` branch (daily cron + manual trigger)
- Uses conventional commit types: `feat` (minor), `fix`/`docs`/`chore`/`refactor`/`test` (patch), `break`/BREAKING CHANGE (major)
- GoReleaser produces cross-platform builds (linux, darwin, windows, freebsd) for multiple architectures
- Builds are GPG-signed

## Commit Message Convention

Follow [Conventional Commits](https://www.conventionalcommits.org/):
```
<type>(<optional scope>): <description>
```

Types: `feat`, `fix`, `docs`, `style`, `refactor`, `perf`, `test`, `build`, `ci`, `chore`, `revert`

## Linting Rules

Configured in `.golangci.yml` (v2 format). Key enabled linters:
- `errcheck`, `govet`, `staticcheck`, `unused` — standard correctness checks
- `misspell`, `ineffassign`, `unconvert`, `unparam` — code quality
- `forcetypeassert`, `nilerr`, `makezero`, `predeclared` — safety checks
- `gofmt` — formatting
- `usetesting`, `copyloopvar`, `durationcheck` — Go-specific best practices
