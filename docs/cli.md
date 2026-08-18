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

### Reference records

Each entry in `refs` is one outgoing reference, listed in scenario-name order so
two runs over the same project produce the same output:

```json
{
  "from": "scene1.ks",
  "line": 4,
  "tag": "link",
  "kind": "choice",
  "storage": "route_a.ks",
  "target": "*good_end",
  "label": "good_end",
  "text": "森へ行く",
  "macro": "sel",
  "resolved": true,
  "ui": false
}
```

| Field | Meaning |
| ----- | ------- |
| `from`, `line` | Where the reference is written. For a macro call this is the call site, not the macro definition. |
| `tag` | The tag that made the reference: `call`, `jump`, `link`, `glink`, `button`, `s_button`, `showbutton`. |
| `kind` | `choice` for a control the player picks, `flow` for `call` / `jump`. |
| `storage`, `target` | Exactly what the script wrote, including a value only known at runtime such as `&f.next`. |
| `label` | `target` without its leading `*`, so it can be printed without Markdown reading it as italics. Empty while `target` is unresolved. |
| `text` | The wording a player reads: the `text` parameter, or the body of a `[link] … [endlink]` pair. |
| `macro` | The macro this reference was expanded from. Absent when the reference is written directly. |
| `dynamic` | The fields still holding runtime expressions — any of `storage`, `target`, `text`. Absent when there are none. |
| `resolved` | True only when `dynamic` is empty. |
| `ui` | True for a textless control bound to a runtime destination: a system menu button rather than a story branch. |

Macro calls are expanded before references are collected, so a choice written as
`[my_choice storage="route_a.ks" target="*good_end" text="森へ行く"]` reports what
the call site binds instead of the `%storage` and `&mp.target` placeholders in
the macro body. References written *inside* a `[macro]` body are not reported on
their own — they only become branches once called. When a call site leaves an
argument unbound, the field stays as written and is listed in `dynamic`.

To call the same functionality from Go instead, see the [Go API](go-api.md).
