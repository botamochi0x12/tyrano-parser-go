# Releasing

[← README](../README.md)

## Cutting a release

Releases are cut by pushing a [semantic version](https://semver.org/) tag:

```console
git tag v1.2.3
git push origin v1.2.3
```

The tag must match `vMAJOR.MINOR.PATCH` with an optional prerelease or build
suffix. Anything else fails the workflow before a release is published. A tag
carrying a prerelease suffix such as `v1.2.3-rc.1` is published as a prerelease.

## What the workflow does

For each tag the release workflow:

1. runs `go run ./tools/ci`
2. builds one archive per platform, stamping the tag into the binary through
   `-ldflags -X main.version`
3. unpacks the linux archive and asks the binary to report its version, so a
   stripped executable bit, a wrong layout, or a mis-stamped build fails here
   rather than reaching users
4. uploads the archives plus `checksums.txt` to the GitHub Release, with
   generated release notes and install instructions

Every pull request runs the same packaging path as a `dev` snapshot and attaches
the archives to the CI run, so a broken release surfaces in review rather than
on a tag that is already public.

## Locally

To produce the same artifacts in `dist/` without publishing:

```console
go run ./tools/release-snapshot
go run ./tools/release-snapshot --version v1.2.3
```

`dist/` ends up holding exactly what would be published: one archive per target
plus `checksums.txt`. The raw binaries are staged inside `dist/` and removed
again.

## Targets

| GOOS    | GOARCH | Archive  |
| ------- | ------ | -------- |
| darwin  | amd64  | `.tar.gz` |
| darwin  | arm64  | `.tar.gz` |
| linux   | amd64  | `.tar.gz` |
| linux   | arm64  | `.tar.gz` |
| windows | amd64  | `.zip`   |

Binaries are built with `CGO_ENABLED=0` and `-trimpath`, so a download runs on
any host of that platform without a matching libc.
