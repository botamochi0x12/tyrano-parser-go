# Installation

[← README](../README.md)

## Download a release build

Every `v*` tag publishes prebuilt archives on the
[Releases page](https://github.com/botamochi0x12/tyrano-parser-go/releases) for
linux, macOS, and Windows on amd64 and arm64. Each archive holds the
`tyrano-parser` binary with its executable bit, plus `README.md` and `LICENSE`.

Download the archive for your platform, extract it, and put `tyrano-parser` on
your PATH:

```console
VERSION=v0.2.0
curl -fsSLO "https://github.com/botamochi0x12/tyrano-parser-go/releases/download/${VERSION}/tyrano-parser_${VERSION}_linux_amd64.tar.gz"
tar -xzf "tyrano-parser_${VERSION}_linux_amd64.tar.gz"
install -Dm0755 tyrano-parser ~/.local/bin/tyrano-parser
tyrano-parser version
```

Swap `linux_amd64` for `linux_arm64`, `darwin_amd64`, `darwin_arm64`, or
`windows_amd64` (published as a `.zip`).

## Verify a download

Each release ships a `checksums.txt` covering every archive:

```console
curl -fsSLO "https://github.com/botamochi0x12/tyrano-parser-go/releases/download/${VERSION}/checksums.txt"
sha256sum -c --ignore-missing checksums.txt
```

## Install with Go

```console
go install github.com/botamochi0x12/tyrano-parser-go/cmd/tyrano-parser@latest
```

Pin a release by replacing `@latest` with a tag such as `@v0.2.0`.

## Build from source

```console
go build ./cmd/tyrano-parser
```

This creates a local `tyrano-parser` binary in the current directory. To run
without keeping a binary:

```console
go run ./cmd/tyrano-parser --help
```

Source builds report `dev` from `tyrano-parser version`; released binaries
report their tag, because the release build stamps it in at link time. See
[Releasing](releasing.md) for how that happens.
