# AGENTS.md

Orientation for coding agents working in this repository. Read this file first,
then open only the page you need — the details live in `docs/`, not here.

## What this is

A Go library and CLI that parses TyranoScript projects: `.ks` scenario files,
real-format `Config.tjs` files, and whole project trees with cross-reference
checks. Module path `github.com/botamochi0x12/tyrano-parser-go`. No third-party
dependencies — the standard library only.

## Architecture

Dependencies point one way:

```text
types/  ←  parser/  ←  loader/  ←  cmd/
```

- `types/` — JSON/wire structs and parse issue types. No logic.
- `parser/` — pure string-in / struct-out. **Never touches the filesystem.**
- `loader/` — file I/O, project discovery, reference extraction, project scans.
- `cmd/tyrano-parser/` — CLI entry point, flag parsing, JSON and report renderers.
- `tools/` — Go programs that run CI, build releases, and check doc links.

Putting file access into `parser/` is the mistake to avoid: the parser stays
testable from plain strings precisely because it has none.

## Non-negotiables

1. **Test first.** Write a failing test, run it, confirm it fails for the reason
   you expect, then write the smallest code that passes. A test that passes the
   moment you write it is testing existing behavior — fix the test.
2. **Verify before claiming done.** `go run ./tools/ci` must pass. It runs
   `go test ./...`, `go test -race ./...`, `go test -cover ./...`, and
   `go vet ./...`, stopping at the first failure.
3. **One context per commit.** Stage explicit paths, never `git add -p`.
   [Conventional Commits](https://www.conventionalcommits.org/en/v1.0.0/) with a
   capitalized verb: `feat: Enable ...`, `test: Assert ...`, `fix: Preserve ...`.
   The body says why the change is needed.
4. **No new dependencies** without a reason the standard library cannot cover.
5. **Coverage** stays at or above 80% per package for `parser/`, `loader/`,
   `types/`, and `cmd/`.

## Commands

```console
go run ./tools/ci                 # full verification suite
go run ./tools/release-snapshot   # build and verify release archives into dist/
go run ./tools/docs-links         # check relative links in Markdown
go run ./cmd/tyrano-parser --help # CLI
```

## Where to read next

| If you are… | Read |
| ----------- | ---- |
| adding or changing CLI behavior | [docs/cli.md](docs/cli.md) |
| working in `parser/` or `loader/` | [docs/go-api.md](docs/go-api.md), [docs/tyranoscript-support.md](docs/tyranoscript-support.md) |
| deciding what syntax to support | [docs/tyranoscript-support.md](docs/tyranoscript-support.md) |
| touching the build, CI, or `tools/` | [docs/development.md](docs/development.md) |
| cutting a release | [docs/releasing.md](docs/releasing.md) |
| changing install instructions | [docs/installation.md](docs/installation.md) |

Design records for past work are under `docs/superpowers/`. They are dated
snapshots of decisions, not living documentation — read them for context, do not
edit them to match new behavior.

## Skills

Repeated workflows are captured as skills in `.claude/skills/`:

- `parser-feature` — adding syntax or CLI behavior across the layers, test-first
- `releasing` — cutting a semantic version release and verifying it landed

## Traps

- `parser/` must not read files. Project discovery belongs in `loader/`.
- Release archives must keep the binary's executable bit and carry `LICENSE`;
  `tools/release-snapshot` verifies this by unpacking and running the result.
- A release tag must be valid semver (`vMAJOR.MINOR.PATCH`) or the workflow
  fails before publishing.
- Moving or renaming a Markdown page breaks relative links; `go test ./...`
  catches it through `tools/docs-links`.
