# tyrano-parser-go CLI — Design Doc

**Date:** 2026-05-03
**Status:** Approved (pre-implementation)
**Author:** botamochi0x12

**Fixture note:** This design doc references `StarGazers` as a project-specific
real-world fixture used during implementation. User-facing README examples use
the linked official TyranoScript sample project at `tyranoscript/data/`.

## Background

`tyrano-parser-go` currently exposes two pure parsers (`parser.ScenarioParser` for `.ks`, `parser.ConfigParser` for `Config.tjs`) consumed by an example `main.go`. Two gaps block real use:

1. **The config parser is wrong for real `Config.tjs` files.** A recent fix made it work for the example in `main.go` (`title="My Game";`) by treating any non-`;` line as an assignment, but real TyranoScript configs use `;key = value;` as the assignment form and `//` for comments — so the current parser silently drops every real line.
2. **There is no file I/O layer or CLI.** Consumers must write Go to parse anything, and there is no way to walk a project, follow `@call` / `@jump` references between scenarios, or extract metadata from `Config.tjs`.

## Goals

- Provide a single CLI binary `tyrano-parser` that:
  - parses one `.ks` scenario file,
  - parses one `Config.tjs` file (real format),
  - scans an entire TyranoScript project (full walk or BFS from an entrypoint),
  - emits JSON (default) or a human-readable report.
- Auto-discover the project root (`data/scenario/` and `data/system/Config.tjs`) from the current directory.
- Fix `parser.ConfigParser` to handle real `Config.tjs` syntax.
- Keep `parser/` pure (string-in / struct-out) so its tests stay free of file I/O.

## Non-Goals (v1)

- Full TJS evaluation. We treat config values as opaque strings (no expression evaluation for things like `960-32`).
- Subdirectory scenario layouts. `storage=` is resolved flat against `data/scenario/`.
- LSP / editor integration. The CLI is the only frontend.
- Live-reload / watch mode.
- Color output / TTY detection beyond plain ASCII status markers.

## Architecture

```
cmd/tyrano-parser/         CLI entry point (thin orchestration)
├── main.go
├── command_scenario.go
├── command_config.go
├── command_scan.go
├── render_json.go
└── render_report.go

loader/                    NEW: file I/O + project layout discovery
├── project.go             ProjectLayout, DiscoverRoot, LayoutFrom
├── scenario_loader.go     LoadScenarioFile, WalkScenarios
├── config_loader.go       LoadConfigFile
├── refs.go                ExtractRefs, ScenarioRef
└── scan.go                ScanProject, ResolveFrom, ProjectScan

parser/                    UNCHANGED contract: pure string-in / struct-out
├── parser.go              (unchanged)
├── scenario.go            (unchanged)
└── config.go              REWRITTEN for real Config.tjs format

types/                     extended with project-level result types
├── parse_issue.go         + MissingStorageError, MissingLabelError
└── project_scan.go        NEW
```

**Invariant:** `parser/` never touches the filesystem. `loader/` reads bytes and hands strings to `parser/`. CLI orchestrates.

**Why this shape:**
- Existing parser tests stay valid (no I/O to mock).
- Loader is testable with `testing/fstest.MapFS`.
- CLI is a thin wrapper — replaceable with a different frontend later.

## Config Parser Fix

Real `Config.tjs` grammar (confirmed against `StarGazers/data/system/Config.tjs`):

| Construct | Example | Today's parser |
|---|---|---|
| Assignment prefix is `;` | `;System.title = StarGazers` | treats as comment ❌ |
| `//` is the only comment | `// この行は…` | already handled ✓ |
| Trailing `;` is OPTIONAL | `;scWidth = 1280` | flagged as error ❌ |
| Inline `//` after value | `;cursorDefault = default; // 通常のマウスカーソル` | swallowed into value ❌ |
| Dotted keys | `scPositionX.left` | works (string key) ✓ |
| Unquoted identifiers/numbers/expressions | `960-32`, `webstorage`, `0xffffff` | works (raw string) ✓ |
| Trailing full-width whitespace | `;unReadTextSkip = false　` | TrimSpace handles ✓ |

### New parsing loop

```
for each line:
  trim
  skip empty
  skip if line starts with "//"
  skip if line starts with "/*" and ends with "*/"
  if line starts with ";":
    strip leading ";", trim
    strip inline "// ..." comment from the right (respecting quotes)
    strip optional trailing ";"
    split on first "=" → key, value
    handleQuotedValue(value)
    config.Set(key, value)
  else:
    if line has no "=":  ignore (TJS code outside our subset)
    else:                emit Warning issue (unrecognized assignment form)
```

### Inline-comment stripping

Must respect quotes. Walk the value left-to-right, track quote state, cut at first unquoted `//`. Required because values like
`userFace = Quicksand, ..., "Yu Gothic", ..., sans-serif, Arial;`
contain commas and quotes that must not be confused with comment markers.

### `main.go` example

The current example content (`title="My Game";`) does not match real `Config.tjs`. Update it to `;title = "My Game";` so the example matches what the CLI will consume. The example is otherwise preserved as a smoke test.

## `loader/` Package API

```go
package loader

type ProjectLayout struct {
    Root        string  // absolute project root (contains data/)
    ScenarioDir string  // <root>/data/scenario
    ConfigPath  string  // <root>/data/system/Config.tjs (may not exist)
}

// DiscoverRoot walks upward from start until it finds data/scenario/
// or data/system/Config.tjs. Returns ErrProjectNotFound if neither
// marker exists by filesystem root.
func DiscoverRoot(start string) (*ProjectLayout, error)

// LayoutFrom constructs a ProjectLayout for an explicit root (no walk).
// Use when --project-root is passed.
func LayoutFrom(root string) (*ProjectLayout, error)

// LoadConfigFile reads ConfigPath, parses it, returns ConfigMap + ParseResult.
// Returns (nil, nil, nil) if ConfigPath does not exist.
func LoadConfigFile(layout *ProjectLayout, p *parser.ConfigParser) (
    types.ConfigMap, *types.ParseResult, error)

// LoadScenarioFile reads one .ks path (absolute or relative to ScenarioDir).
func LoadScenarioFile(path string, p parser.TyranoParser) (
    *types.ParsedScenario, *types.ParseResult, error)

// WalkScenarios walks ScenarioDir and parses every *.ks file.
// Returns map keyed by path relative to ScenarioDir.
func WalkScenarios(layout *ProjectLayout, p parser.TyranoParser) (
    map[string]*types.ParsedScenario, *types.ParseResult, error)
```

### Cross-reference extraction

```go
type ScenarioRef struct {
    From    string  // source file (relative to ScenarioDir)
    Line    int
    Tag     string  // "call" | "jump" | "link"
    Storage string  // "title.ks" or "" if same-file label jump
    Target  string  // "*start" or ""
}

// ExtractRefs scans a parsed scenario for @call/@jump/@link tags carrying
// storage= or target= parameters and returns them.
func ExtractRefs(from string, scenario *types.ParsedScenario) []ScenarioRef
```

### Project scan

```go
type ProjectScan struct {
    Layout    *ProjectLayout
    Config    types.ConfigMap
    Scenarios map[string]*types.ParsedScenario  // by relative path
    Refs      []ScenarioRef
    Issues    *types.ParseResult                // aggregated parse + ref issues
}

// ScanProject: full walk of ScenarioDir.
func ScanProject(layout *ProjectLayout, p parser.TyranoParser) (*ProjectScan, error)

// ResolveFrom: BFS from a single entrypoint, parsing only reachable files.
func ResolveFrom(layout *ProjectLayout, p parser.TyranoParser, entrypoint string) (*ProjectScan, error)
```

### Sentinel errors

```go
var (
    ErrProjectNotFound = errors.New("no TyranoScript project found")
    ErrNoScenarioDir   = errors.New("project missing data/scenario directory")
)
```

## Cross-Reference Resolution

### Full walk (`scan` without entrypoint)

```
1. DiscoverRoot(cwd) → ProjectLayout
2. LoadConfigFile(layout) → ConfigMap + ParseResult     [skip on missing, warn]
3. WalkScenarios(layout) → map[relPath]*ParsedScenario
4. for each scenario: ExtractRefs(rel, scenario) → append to global refs
5. for each ref:
     - if storage != "" && storage not in map      → MissingStorageError
     - if target != "" && storage == "":
         look up source scenario's Labels           → MissingLabelError if absent
     - if target != "" && storage != "":
         look up target scenario's Labels           → MissingLabelError if absent
6. assemble ProjectScan
```

### BFS from entrypoint (`scan <file>`)

```
1. DiscoverRoot
2. LoadConfigFile
3. queue := [entrypoint]; visited := {}
4. while queue not empty:
     f := pop; skip if visited
     parse f → scenario; collect issues
     refs := ExtractRefs(f, scenario)
     for ref with non-empty storage:
       resolve storage path against ScenarioDir
       if file missing → Issue
       else queue.push(storage)
5. validate label refs (same rules as full walk)
6. assemble ProjectScan with only visited files
```

### Path resolution for `storage="..."`

`storage` values are filenames, not paths — `storage="title.ks"` resolves to `<ScenarioDir>/title.ks`. Confirmed by `StarGazers/data/scenario/first.ks` containing `@call storage="tyrano.ks"`. Subdirectory layouts are out of scope for v1.

### New issue types

```go
const (
    MissingStorageError ParseErrorType = "missing_storage"
    MissingLabelError   ParseErrorType = "missing_label"
)
```

## CLI Commands & Flags

Single binary `tyrano-parser` with subcommands. `flag` from stdlib (no cobra).

```
tyrano-parser <command> [flags] [args]

Commands:
  scenario <file.ks>           Parse one scenario file. Path required.
  config   [<Config.tjs>]      Parse one config file. Default: auto-discover.
  scan     [<entrypoint.ks>]   Scan project. No arg: walk all. With arg: BFS.

Global flags:
  --project-root <dir>   Skip auto-discovery, use this as root.
  --format json|report   Output format. Default: json.
  --strict               Strict mode: any parse issue → exit non-zero.
  --quiet                Suppress warnings (errors still emitted).
```

### Behavior

| Command | No arg | With arg | Auto-discover root? |
|---|---|---|---|
| `scenario` | error: file required | parse the file | only if `--project-root` not set |
| `config` | parse `<root>/data/system/Config.tjs` | parse explicit path | yes |
| `scan` | walk all `.ks` in scenario dir | BFS from entrypoint | yes |

### Auto-discovery

1. Start from CWD (or first arg's dir if it's a file path).
2. Walk upward looking for `data/scenario/` or `data/system/Config.tjs`.
3. Stop at filesystem root → exit 2 with "no TyranoScript project found from <start>".
4. `--project-root` short-circuits the walk.

### Exit codes

- `0` — success, no errors (warnings OK)
- `1` — parse errors (or any issues with `--strict`)
- `2` — I/O errors (file not found, permission denied)
- `3` — usage errors (bad flags, missing required args)

### Examples

```bash
tyrano-parser scenario data/scenario/first.ks
tyrano-parser config --format report
tyrano-parser scan > project.json
tyrano-parser scan --strict data/scenario/first.ks
```

## Output Formats

### JSON (default)

Single document, pretty-printed (2-space indent), schema-stable for `jq` and downstream tooling. Empty `issues: []` always present.

**`scenario`:**
```json
{
  "kind": "scenario",
  "path": "data/scenario/first.ks",
  "scenario": { "elements": [...], "labels": {...} },
  "issues": [...]
}
```

**`config`:**
```json
{
  "kind": "config",
  "path": "data/system/Config.tjs",
  "config": { "System.title": "StarGazers", "scWidth": "1280" },
  "issues": [...]
}
```

**`scan`:**
```json
{
  "kind": "project_scan",
  "root": "/abs/path/to/StarGazers",
  "config": { ... },
  "scenarios": {
    "first.ks": { "elements": [...], "labels": {...} },
    "title.ks": { ... }
  },
  "refs": [
    { "from": "first.ks", "line": 12, "tag": "call",
      "storage": "tyrano.ks", "target": "" }
  ],
  "issues": [
    { "type": "missing_storage", "severity": "error",
      "line": 0, "column": 0,
      "message": "referenced storage 'missing.ks' not found",
      "context": "from=first.ks line=42" }
  ]
}
```

### Report (`--format report`)

```
Project: /abs/path/to/StarGazers
Config:  data/system/Config.tjs (87 keys)

Scenarios (8 files):
  ✓ first.ks         42 elements, 3 labels
  ⚠ scene1.ks        91 elements, 4 labels — 2 warnings
  ✗ make.ks          —— parse error

Cross-references (12):
  first.ks:12   @call storage="tyrano.ks"     → ✓
  scene1.ks:34  @jump storage="missing.ks"    → ✗ not found

Issues:
  [error]   make.ks:5:3        unmatched bracket in tag
  [error]   scene1.ks:34:1     referenced storage 'missing.ks' not found
  [warning] scene1.ks:67:12    deprecated tag 'old_tag'

Summary: 1 error, 2 warnings, 11/12 refs resolved
```

`✓ ⚠ ✗` are ASCII-safe markers. `--quiet` suppresses warnings from the `Issues:` section.

### Stream conventions

- JSON output goes to stdout; exit code reflects severity.
- Report output goes to stdout (not stderr) so it can be redirected.
- Pre-parse I/O errors write a one-line message to stderr and exit 2 *before* any stdout output.

## Error Handling

| Layer | Returns | Why |
|---|---|---|
| `parser/` | `(result, *ParseResult)` — never `error` for parse issues | Issues are data, not exceptions |
| `loader/` | `(result, error)` — `error` for I/O only; parse issues stay in `*ParseResult` | I/O failures are real errors |
| `cmd/` | nothing returned; writes to stdout/stderr, calls `os.Exit` | Only place that decides exit codes |

Wrapping convention (per `~/.claude/rules/golang/coding-style.md`):
```go
return nil, fmt.Errorf("loader: read config %s: %w", path, err)
```

## Testing Strategy

### `parser/config.go` (rewrite)

Table-driven tests covering every line shape from real `Config.tjs`:
- `;key = value;` happy path
- `;key = value` no trailing semicolon
- `;key = value; // inline comment` inline stripped
- `;key.dotted = value;` dotted keys preserved
- `;key = "Quicksand, ..., \"Yu Gothic\", ...";` quoted with embedded `,` and `"`
- `// pure comment line` skipped
- `;key = false　` full-width trailing whitespace
- malformed: `;= value`, `;key`, `;key = "unterminated`

### `loader/`

`testing/fstest.MapFS` for fake project structures. No real disk I/O in unit tests.

- `DiscoverRoot` from nested CWD
- `DiscoverRoot` failing past filesystem root
- `WalkScenarios` skipping non-`.ks` files
- `ExtractRefs` extracting `storage=` and `target=`
- `ScanProject` flagging missing storage/labels

### `cmd/tyrano-parser/`

Golden-file tests. Run binary against `testdata/projects/minimal/`, diff stdout against `testdata/expected/<case>.json` and `<case>.txt`.

### Integration test

A single test parses real `StarGazers/` and asserts:
- zero parse errors,
- expected number of scenario files,
- `config["System.title"] == "StarGazers"`.

Catches regressions if grammar evolves.

### Coverage

- 80% minimum (per `~/.claude/rules/common/testing.md`).
- Run with `go test -race ./...`.

## File Layout (final)

```
tyrano-parser-go/
├── cmd/tyrano-parser/
│   ├── main.go
│   ├── command_scenario.go
│   ├── command_config.go
│   ├── command_scan.go
│   ├── render_json.go
│   ├── render_report.go
│   └── *_test.go
├── loader/
│   ├── project.go
│   ├── scenario_loader.go
│   ├── config_loader.go
│   ├── refs.go
│   ├── scan.go
│   └── *_test.go
├── parser/
│   ├── parser.go         (unchanged)
│   ├── scenario.go       (unchanged)
│   ├── config.go         (rewritten)
│   └── config_test.go    (expanded)
├── types/
│   ├── scenario.go       (unchanged)
│   ├── config.go         (unchanged)
│   ├── parse_result.go   (unchanged)
│   ├── parse_issue.go    (+ MissingStorageError, MissingLabelError)
│   └── project_scan.go   (NEW)
├── testdata/projects/minimal/   (fixture for loader+cmd tests)
├── main.go               → DELETED (moved into cmd/)
├── docs/superpowers/specs/2026-05-03-tyrano-parser-cli-design.md
└── go.mod                (no new deps)
```

## Phasing

Each phase ends with passing tests + a commit. Phases 2–5 are independent of P1 once it lands.

| Phase | Scope | Ships standalone? |
|---|---|---|
| **P1** | Rewrite `parser/config.go`, expand tests, integration test against real `Config.tjs`. Update `main.go` example. | ✓ fixes the regression even before the CLI |
| **P2** | `loader/` core: `project.go`, `scenario_loader.go`, `config_loader.go` with `fstest` tests | depends on P1 |
| **P3** | `loader/refs.go`, `loader/scan.go`; integration test against `StarGazers/` | depends on P2 |
| **P4** | `cmd/tyrano-parser/` + JSON renderer; golden tests; move/delete root `main.go` | depends on P3 |
| **P5** | Report renderer; golden tests for human-readable output | depends on P4 |

## Risks & Open Questions

- **TJS expression values.** Some configs contain expressions like `mw = 960-32`. We store these as raw strings; downstream consumers must evaluate themselves. Document this in the JSON schema.
- **Subdirectory scenarios.** Some projects may organize `.ks` into subdirs. Out of scope for v1; revisit if a real project surfaces.
- **Encoding.** Real `Config.tjs` files are UTF-8 in practice but historically may be Shift-JIS. v1 assumes UTF-8; document and revisit if it bites.
- **Macro/include tags.** TyranoScript supports `[macro]` definitions and `@iscript` blocks. The current parser already handles `[iscript]/[endscript]`; macros are not specially resolved in v1 (treated as opaque tags).

## Dependencies

None new. Stdlib only.
