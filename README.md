# TyranoScript Parser for Go

Go parser and CLI for TyranoScript projects. It parses individual `.ks` scenario
files, real-format `Config.tjs` files, and full project trees with `@call` /
`@jump` / `[link]` cross-reference checks.

The parser layer is pure string-in / struct-out. File I/O and project discovery
live in `loader/`, and the command-line interface lives in `cmd/tyrano-parser/`.

## Features

- Parse `.ks` files into KAG-compatible JSON-shaped structs.
- Parse real `Config.tjs` assignments such as `;key = value;`, with quote-aware
  `//` comment stripping.
- Scan a project under `data/scenario/` and `data/system/Config.tjs`.
- Report missing referenced files and missing labels, while skipping dynamic
  runtime references.
- Emit JSON or human-readable reports from the CLI.

## Quick start

Download a release archive, extract it, and put the binary on your PATH:

```console
VERSION=v0.2.0
curl -fsSLO "https://github.com/botamochi0x12/tyrano-parser-go/releases/download/${VERSION}/tyrano-parser_${VERSION}_linux_amd64.tar.gz"
tar -xzf "tyrano-parser_${VERSION}_linux_amd64.tar.gz"
install -Dm0755 tyrano-parser ~/.local/bin/tyrano-parser
```

Then scan a project:

```console
tyrano-parser scan --project-root path/to/project --format report
```

Other ways to install — `go install`, building from source, checksum
verification — are in [Installation](docs/installation.md).

## Documentation

| Page | What it covers |
| ---- | -------------- |
| [Installation](docs/installation.md) | Release downloads, `go install`, source builds, checksums |
| [CLI Reference](docs/cli.md) | Commands, flags, exit codes, JSON output |
| [Go API](docs/go-api.md) | Using `parser/` and `loader/` from Go |
| [TyranoScript Support](docs/tyranoscript-support.md) | Project layout, `Config.tjs` syntax, what is out of scope |
| [Development](docs/development.md) | Repository structure, verification suite, commit conventions |
| [Releasing](docs/releasing.md) | Semantic version tags and the release workflow |

Contributors and coding agents should start from [AGENTS.md](AGENTS.md), which
indexes the same material along with the conventions this repository expects.

## Development

```console
go run ./tools/ci
```

runs the full verification suite: tests, race, coverage, and vet. See
[Development](docs/development.md) for the rest.

## License

Released under the [MIT License](LICENSE). Every release archive ships a copy
alongside the binary.
