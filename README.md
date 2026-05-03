# TyranoScript Parser for Go

Go parser and CLI for TyranoScript projects. It can parse individual `.ks`
scenario files, real-format `Config.tjs` files, and full TyranoScript project
trees with basic `@call` / `@jump` / `[link]` cross-reference checks.

The parser layer is pure string-in / struct-out. File I/O and project discovery
live in `loader/`, and the command-line interface lives in `cmd/tyrano-parser/`.

## Features

- Parse TyranoScript `.ks` files into KAG-compatible JSON-shaped structs.
- Parse real TyranoScript `Config.tjs` assignments such as `;key = value;`.
- Strip `//` inline comments from config values without breaking quoted URLs.
- Load and scan TyranoScript project layouts under `data/scenario/` and
  `data/system/Config.tjs`.
- Extract references from `@call`, `@jump`, and `[link]` tags.
- Report missing referenced `.ks` files and missing labels for static refs.
- Skip dynamic runtime refs such as `storage=&tf.storage`.
- Emit JSON or human-readable reports from the CLI.

## Install / Build

```bash
go build ./cmd/tyrano-parser
```

This creates a local `tyrano-parser` binary in the current directory. To run
without keeping a binary:

```bash
go run ./cmd/tyrano-parser --help
```

## CLI Usage

```text
tyrano-parser <command> [flags] [args]

Commands:
  scenario <file.ks>           Parse one scenario file.
  config   [<Config.tjs>]      Parse one config file (default: auto-discover).
  scan     [<entrypoint.ks>]   Scan project (default: walk all scenarios).

Global flags:
  --project-root <dir>   Skip auto-discovery, use this as project root.
  --format json|report   Output format (default: json).
  --strict               Strict mode: any parse issue exits non-zero.
  --quiet                Suppress warnings (errors still emitted).
```

Examples below use the linked official TyranoScript sample project. Its `data/`
directory is available at `tyranoscript/data/`, so the project root passed to
`--project-root` is `tyranoscript`.

Examples:

```bash
go run ./cmd/tyrano-parser scenario tyranoscript/data/scenario/first.ks
go run ./cmd/tyrano-parser scenario tyranoscript/data/scenario/first.ks --format report

go run ./cmd/tyrano-parser config --project-root tyranoscript
go run ./cmd/tyrano-parser config tyranoscript/data/system/Config.tjs --format report

go run ./cmd/tyrano-parser scan --project-root tyranoscript
go run ./cmd/tyrano-parser scan --project-root tyranoscript first.ks --format report
```

Exit codes:

- `0`: success, no parse errors
- `1`: parse errors, or any issue when `--strict` is set
- `2`: I/O or project discovery error
- `3`: usage error

## JSON Output

`scenario` emits:

```json
{
  "kind": "scenario",
  "path": "data/scenario/first.ks",
  "scenario": {
    "array_s": [],
    "map_label": {}
  },
  "issues": []
}
```

`config` emits:

```json
{
  "kind": "config",
  "path": "data/system/Config.tjs",
  "config": {
    "System.title": "ティラノスクリプト",
    "scWidth": "1280"
  },
  "issues": []
}
```

`scan` emits:

```json
{
  "kind": "project_scan",
  "root": "/path/to/project",
  "config": {},
  "scenarios": {},
  "refs": [],
  "issues": []
}
```

## Go API

### Pure Parsers

```go
package main

import "github.com/botamochi0x12/tyrano-parser-go/parser"

func parseStrings(scenarioContent, configContent string) error {
    p := parser.NewDefaultTyranoParser()

    scenario, scenarioResult := p.ParseScenarioWithResult(scenarioContent)
    _ = scenario
    if scenarioResult.HasErrors() {
        return scenarioResult
    }

    config, configResult := p.ParseConfigWithResult(configContent)
    _ = config
    if configResult.HasErrors() {
        return configResult
    }

    return nil
}
```

### Loader / Project Scans

```go
package main

import (
    "github.com/botamochi0x12/tyrano-parser-go/loader"
    "github.com/botamochi0x12/tyrano-parser-go/parser"
)

func scanProject(root string) error {
    layout, err := loader.LayoutFrom(root)
    if err != nil {
        return err
    }

    scan, err := loader.ScanProject(layout, parser.NewDefaultTyranoParser())
    if err != nil {
        return err
    }

    if scan.Issues.HasErrors() {
        return scan.Issues
    }
    return nil
}
```

## Config.tjs Support

Real TyranoScript configs use a leading semicolon as the assignment marker:

```tjs
// comment
;System.title = "My Game";
;scWidth = 1280
;userFace = Quicksand, "Yu Gothic", sans-serif; // inline comment
;url = "http://example.com/path";
```

Supported behavior:

- leading `;` assignment prefix
- optional trailing `;`
- dotted keys
- quoted and unquoted values
- quote-aware `//` inline comment stripping
- raw string preservation for expressions such as `960-32`

Full TJS evaluation and legacy non-UTF-8 encodings are out of scope.

## Project Layout

The loader expects a standard TyranoScript layout:

```text
project/
└── data/
    ├── scenario/
    │   ├── first.ks
    │   └── title.ks
    └── system/
        └── Config.tjs
```

`loader.DiscoverRoot` walks upward from a starting directory until it finds
`data/scenario/` or `data/system/Config.tjs`. `loader.LayoutFrom` uses an
explicit project root and requires `data/scenario/`.

## Repository Structure

```text
cmd/tyrano-parser/          CLI entry point and output renderers
cmd/tyrano-parser-example/  Programmatic usage example
loader/                     File I/O, project discovery, refs, scans
parser/                     Pure TyranoScript and Config.tjs parsers
types/                      JSON/wire structs and parse issue types
docs/superpowers/           Design and implementation planning docs
```

## Development

Run the full verification suite:

```bash
go test -race ./...
go test -cover ./...
go vet ./...
```

Current coverage target is at least 80% per package.

Run the example:

```bash
go run ./cmd/tyrano-parser-example
```
