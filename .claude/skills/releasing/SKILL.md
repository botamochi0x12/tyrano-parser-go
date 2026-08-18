---
name: releasing
description: Use when cutting a release of tyrano-parser-go, choosing a version number, or diagnosing a release workflow that failed - covers the semantic version rules, the tag push that publishes, and how to confirm the artifacts actually landed.
---

# Cutting a release

Pushing a `v*` tag is what publishes. Merging does not.

## 1. Choose the version

The tag must be `vMAJOR.MINOR.PATCH`, optionally with a prerelease or build
suffix. Anything else fails the workflow before anything is published.

| The release contains… | Bump |
| --------------------- | ---- |
| a breaking change to the CLI contract or the Go API | MAJOR |
| a backward-compatible feature — new command, flag, syntax support | MINOR |
| only fixes, docs, or internal changes | PATCH |

While the project is below `1.0.0`, a breaking change bumps MINOR. Exit codes,
flag names, and JSON field names are part of the CLI contract.

A tag with a prerelease suffix such as `v1.2.3-rc.1` is published as a
prerelease automatically.

## 2. Verify locally before tagging

```console
go run ./tools/ci
go run ./tools/release-snapshot --version vX.Y.Z
```

The snapshot builds every target, packages the archives, and unpacks the one
built for this host to confirm the binary runs and reports `vX.Y.Z`. If that
last step fails, do not tag — the same failure would happen after publishing.

`dist/` should hold exactly one archive per target plus `checksums.txt`.

## 3. Tag and push

Tag the commit on `main` that the release is cut from:

```console
git fetch origin main
git tag -a vX.Y.Z <merge-commit> -m "Release vX.Y.Z: <one line>"
git push origin vX.Y.Z
```

Do not sign the tag unless asked.

## 4. Watch the workflow

```console
gh run list --workflow=Release --limit 1
gh run watch <run-id> --exit-status
```

The job runs the CI suite, builds and self-verifies the archives, then publishes
the GitHub Release with generated notes.

## 5. Confirm it landed

Do not stop at "the workflow was green" — check what a user would actually get:

```console
gh release view vX.Y.Z
```

Expect one archive per target plus `checksums.txt`. Then download and run it:

```console
curl -fsSLO "https://github.com/botamochi0x12/tyrano-parser-go/releases/download/vX.Y.Z/tyrano-parser_vX.Y.Z_linux_amd64.tar.gz"
curl -fsSLO "https://github.com/botamochi0x12/tyrano-parser-go/releases/download/vX.Y.Z/checksums.txt"
sha256sum -c --ignore-missing checksums.txt
tar -xzf tyrano-parser_vX.Y.Z_linux_amd64.tar.gz
./tyrano-parser version
```

The reported version must match the tag.

## When the workflow fails

- **Rejected tag** — the tag is not valid semver. Delete it
  (`git push --delete origin vX.Y.Z`), fix the name, tag again.
- **Verification step fails** — the packaged binary did not run or reported the
  wrong version. Reproduce with `go run ./tools/release-snapshot --version vX.Y.Z`
  and fix it on a branch; do not retag until it passes locally.
- **A tag already published** — do not move or force-push it. Cut the next
  PATCH version instead.

Details of what the workflow does are in [docs/releasing.md](../../../docs/releasing.md).
