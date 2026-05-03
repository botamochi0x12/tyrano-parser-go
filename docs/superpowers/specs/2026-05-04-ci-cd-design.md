# tyrano-parser-go CI/CD - Design Doc

**Date:** 2026-05-04
**Status:** Approved (pre-implementation)
**Author:** botamochi0x12

## Background

`tyrano-parser-go` now has a real parser library, project loader, and
`tyrano-parser` CLI. The project has a clear local verification sequence:

```bash
go test ./...
go test -race ./...
go test -cover ./...
go vet ./...
```

Those checks should run consistently in GitHub Actions and on contributor
machines, including Windows machines. The repository does not currently have a
CI workflow, release workflow, Makefile, Taskfile, or script entrypoint.

## Goals

- Add GitHub Actions CI for pull requests and pushes to `main`.
- Add a tag-based GitHub release workflow for CLI binaries.
- Provide local commands that run the same core steps as CI/CD.
- Keep the local workflow friendly to Windows users.
- Avoid extra task-runner dependencies such as Make or Task.
- Document the development and release commands in `README.md`.

## Non-Goals

- Package manager distribution such as Homebrew, Scoop, Chocolatey, or npm.
- Publishing releases from non-tag builds.
- Adding external coverage services.
- Adding Windows or macOS CI jobs in v1.
- Signing artifacts.

## Approach

Use small Go helper commands under `tools/` as the shared local and CI entrypoints:

```text
tools/
├── ci/
│   └── main.go
└── release-snapshot/
    └── main.go
```

The helpers are invoked with `go run`, so contributors only need Go installed:

```bash
go run ./tools/ci
go run ./tools/release-snapshot
```

GitHub Actions calls the same helpers instead of duplicating all commands in
workflow YAML. That keeps local and remote behavior aligned and avoids
platform-specific shell assumptions.

## Helper Commands

### `tools/ci`

`tools/ci` runs the project quality gates in order:

```bash
go test ./...
go test -race ./...
go test -cover ./...
go vet ./...
```

Behavior:

- Print a compact step header before each command.
- Stream stdout and stderr directly.
- Stop at the first failing command.
- Exit with a non-zero status when any command fails.
- Use `exec.Command` with argument slices instead of shell strings.

### `tools/release-snapshot`

`tools/release-snapshot` builds local release artifacts into `dist/`.

Behavior:

- Remove and recreate `dist/`.
- Cross-compile `./cmd/tyrano-parser` for the release platform matrix.
- Use deterministic artifact names.
- Write SHA-256 checksums to `dist/checksums.txt`.
- Stop at the first failed build.
- Use `exec.Command` with explicit environment variables for `GOOS` and
  `GOARCH`.

Initial release matrix:

```text
darwin/amd64   -> tyrano-parser_darwin_amd64
darwin/arm64   -> tyrano-parser_darwin_arm64
linux/amd64    -> tyrano-parser_linux_amd64
linux/arm64    -> tyrano-parser_linux_arm64
windows/amd64  -> tyrano-parser_windows_amd64.exe
```

The helper can later grow flags for custom output directories or reduced
platform matrices, but v1 should keep the interface simple.

## GitHub Actions

### CI Workflow

Add `.github/workflows/ci.yml`.

Triggers:

- Pull requests.
- Pushes to `main`.

Behavior:

```text
checkout
setup-go using go.mod
go run ./tools/ci
```

The workflow should use read-only permissions. The helper command owns the
actual check list.

### Release Workflow

Add `.github/workflows/release.yml`.

Triggers:

- Tags matching `v*`.

Behavior:

```text
checkout
setup-go using go.mod
go run ./tools/ci
go run ./tools/release-snapshot
upload dist artifacts to a GitHub Release
```

The release workflow should set `permissions: contents: write` because it needs
to create or update GitHub Releases. CI should not use write permissions.

## Testing

- Add focused unit tests for release platform definitions, artifact naming, and
  checksum manifest behavior.
- Add focused tests for command-runner failure propagation where practical.
- Avoid testing `os/exec` itself.
- Keep helper code small enough that the existing package coverage standard
  remains achievable.
- Continue using the current project checks as the source of truth:
  `go test ./...`, `go test -race ./...`, `go test -cover ./...`, and
  `go vet ./...`.

## Documentation

Update `README.md` with a concise development section:

```bash
go run ./tools/ci
go run ./tools/release-snapshot
```

Also add a release note explaining that tags like `vX.Y.Z` publish CLI
artifacts through GitHub Actions. The CI/CD docs should not mention
`StarGazers`; README examples should remain generic or use the linked official
TyranoScript sample project where project data is needed.

## Error Handling

- Local helper failures should return non-zero exit codes and leave the failing
  command visible in output.
- Release builds should fail fast and avoid publishing partial artifacts.
- If `dist/` cannot be removed or recreated, `tools/release-snapshot` should
  stop before starting builds.
- Checksum writing errors should fail the release snapshot.

## Security

- CI uses read-only GitHub token permissions.
- Release workflow uses `contents: write` only because publishing releases
  requires it.
- Helpers do not execute user-provided shell strings.
- No credentials are stored in repository files.

## Acceptance Criteria

- `go run ./tools/ci` runs all local quality gates and returns non-zero on
  failure.
- `go run ./tools/release-snapshot` creates deterministic CLI artifacts and
  `dist/checksums.txt`.
- Pull request and `main` branch CI call `go run ./tools/ci`.
- Version tags matching `v*` run CI, build release artifacts, and publish a
  GitHub Release.
- `README.md` documents local CI and release snapshot commands.
