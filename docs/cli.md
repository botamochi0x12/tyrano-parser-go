# CLI Reference

[← README](../README.md)

```man
tyrano-parser <command> [flags] [args]

Commands:
  scenario <file.ks>           Parse one scenario file.
  config   [<Config.tjs>]      Parse one config file (default: auto-discover).
  scan     [<entrypoint.ks>]   Scan project (default: walk all scenarios).
  version                      Print the tyrano-parser version.

Global flags:
  --project-root <dir>   Skip auto-discovery, use this as project root.
  --format json|report   Output format (default: json).
  --strict               Strict mode: any parse issue exits non-zero.
  --quiet                Suppress warnings (errors still emitted).
```

## Examples

The examples below use the official TyranoScript sample project. Its `data/`
directory sits at `tyranoscript/data/`, so the project root passed to
`--project-root` is `tyranoscript`.

```console
go run ./cmd/tyrano-parser scenario tyranoscript/data/scenario/first.ks
go run ./cmd/tyrano-parser scenario tyranoscript/data/scenario/first.ks --format report

go run ./cmd/tyrano-parser config --project-root tyranoscript
go run ./cmd/tyrano-parser config tyranoscript/data/system/Config.tjs --format report

go run ./cmd/tyrano-parser scan --project-root tyranoscript
go run ./cmd/tyrano-parser scan --project-root tyranoscript first.ks --format report
```

## Exit codes

- `0`: success, no parse errors
- `1`: parse errors, or any issue when `--strict` is set
- `2`: I/O or project discovery error
- `3`: usage error

## JSON output

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

To call the same functionality from Go instead, see the [Go API](go-api.md).
