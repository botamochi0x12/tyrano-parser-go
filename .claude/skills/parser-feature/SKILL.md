---
name: parser-feature
description: Use when adding or changing TyranoScript syntax support, loader behavior, or CLI commands in this repository - walks the test-first path through types, parser, loader, and cmd, and says which layer owns what.
---

# Adding a parser feature

## Pick the layer first

Dependencies point one way: `types/` ← `parser/` ← `loader/` ← `cmd/`. Change the
lowest layer that can own the behavior, then work outward only as far as needed.

| The change is about… | It belongs in |
| -------------------- | ------------- |
| a new field on the JSON/wire shape, or a new issue kind | `types/` |
| tag syntax, lexing, `Config.tjs` assignment forms | `parser/` |
| finding files, walking a project, extracting or resolving references | `loader/` |
| a command, a flag, an exit code, an output format | `cmd/tyrano-parser/` |

**`parser/` must never touch the filesystem.** If a change seems to need a path
or a file read inside `parser/`, the behavior belongs in `loader/` instead —
pass the already-read string in.

## The cycle

For each behavior, one at a time:

1. **Write one failing test.** Put it beside the code: `parser/tag_test.go`,
   `loader/scan_test.go`, `cmd/tyrano-parser/cli_test.go`. Follow the existing
   style — table-driven subtests, standard `testing` only, no assertion library.
2. **Run it and read the failure.** `go test ./parser/ -run TestYourCase`.
   Confirm it fails because the behavior is missing, not because of a typo.
   If it passes immediately, the test is describing behavior that already
   exists — rewrite it until it fails.
3. **Write the smallest code that passes.** No extra options, no speculative
   generality.
4. **Run the test again, then the package.** Both must be green and the output
   clean.
5. **Refactor** if it helps, keeping the tests green.

## Layer-specific notes

**`parser/`** — Inputs are plain strings, so tests need no fixtures. Cover the
awkward cases explicitly: quoted values containing `//`, unquoted values,
dotted keys, missing trailing `;`, raw expressions such as `960-32`.

**`loader/`** — Build a project tree under `t.TempDir()` rather than adding
fixture files, unless the case needs a checked-in project. Existing fixtures
live at `cmd/tyrano-parser/testdata/projects/minimal/`.

**`cmd/tyrano-parser/`** — `runCLI` in `cli_test.go` builds the binary and runs
it, so tests assert real end-to-end behavior including exit codes. Use
`runCLIWithLDFlags` when the test needs a linker-stamped value. Keep the usage
text in `main.go` and the CLI table in `docs/cli.md` in step with any new
command or flag.

**A new exit code or output field** is a contract change: update
`docs/cli.md` in the same commit.

## Before finishing

```console
go run ./tools/ci
```

All four steps must pass. Coverage stays at or above 80% per package for
`parser/`, `loader/`, `types/`, and `cmd/`.

Then commit one context at a time, staging explicit paths:

```console
git add parser/tag.go parser/tag_test.go
git commit -m "feat: Enable ..."
```

Conventional Commits, capitalized verb, body explaining why the change is
needed.
