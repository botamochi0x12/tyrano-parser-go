# Development

[← README](../README.md)

## Repository structure

```text
cmd/tyrano-parser/          CLI entry point and output renderers
cmd/tyrano-parser-example/  Programmatic usage example
loader/                     File I/O, project discovery, refs, scans
parser/                     Pure TyranoScript and Config.tjs parsers
types/                      JSON/wire structs and parse issue types
tools/ci/                   Local runner for the verification suite
tools/release-snapshot/     Cross-platform release build and packaging
tools/docs-links/           Relative Markdown link checker
docs/                       Documentation pages
docs/superpowers/           Design and implementation planning docs
```

Dependencies point one way: `types` ← `parser` ← `loader` ← `cmd`. The parser
layer never touches the filesystem, which is what keeps it testable from plain
strings.

The module has no third-party dependencies. Keep it that way unless there is a
reason the standard library cannot cover.

## Verification

Run the same suite CI runs:

```console
go run ./tools/ci
```

It runs, stopping at the first failure:

```console
go test ./...
go test -race ./...
go test -cover ./...
go vet ./...
```

Coverage target is at least 80% per package for the library and CLI packages.

## Working on the code

Changes are developed test-first: write a failing test, watch it fail for the
reason you expect, then write the smallest code that passes. Details of the
per-layer workflow live in the `parser-feature` skill under `.claude/skills/`.

Run the example program:

```console
go run ./cmd/tyrano-parser-example
```

Build and verify release archives locally, without publishing:

```console
go run ./tools/release-snapshot
```

Check that every relative link in the Markdown files resolves:

```console
go run ./tools/docs-links
```

The same check runs as an ordinary test, so `go test ./...` already covers it.

## Commits

Commits follow [Conventional Commits](https://www.conventionalcommits.org/en/v1.0.0/)
with a capitalized verb after the type, for example `feat: Enable ...`,
`test: Assert ...`, `fix: Preserve ...`. Each commit covers one context; the
subject says what the change enables and the body says why it is needed.

## Releasing

Cutting a release is a separate flow. See [Releasing](releasing.md).
