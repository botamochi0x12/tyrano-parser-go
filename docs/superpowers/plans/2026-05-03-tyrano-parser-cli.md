# tyrano-parser CLI Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Build a `tyrano-parser` CLI that parses real TyranoScript projects (single `.ks` files, real-format `Config.tjs`, full project scans with cross-reference resolution) and emits JSON or human-readable reports — and along the way fix the regression in `parser/config.go` that breaks real `Config.tjs` files.

**Architecture:** Three layers. `parser/` stays pure (string-in / struct-out). New `loader/` package handles file I/O and project layout discovery. New `cmd/tyrano-parser/` package is the CLI entry point. `loader/` reads bytes and hands strings to `parser/`; CLI orchestrates.

**Tech Stack:** Go (stdlib only — no new deps). `flag` for CLI parsing, `encoding/json` for output, `testing/fstest.MapFS` for I/O tests, golden files for CLI tests.

**Reference:** Full design at `docs/superpowers/specs/2026-05-03-tyrano-parser-cli-design.md`. Real fixture at `StarGazers/data/`.

**Commit attribution:** Every commit message HEREDOC must end with:
```
Co-Authored-By: botamochi0x12 <botamochi0x12@gmail.com>
```

---

## Phase 1: Config Parser Fix

**Why first:** Standalone deliverable. Fixes the regression in commit `3847e78` where `parser/config.go` silently drops every line in real `Config.tjs` files. Independent of any later phase.

**Acceptance:** Real `StarGazers/data/system/Config.tjs` parses with zero errors and produces a `ConfigMap` containing all `;key = value;`-style assignments. `main.go` example still runs.

---

### Task 1.1: Write failing test for `;key = value;` happy path

**Files:**
- Modify: `parser/config_test.go`

- [ ] **Step 1: Read the existing test file to find the table-driven test pattern**

Run: `cat parser/config_test.go | head -60`
Note the existing test function names so the new test can follow the pattern.

- [ ] **Step 2: Append a new test function for the real Config.tjs format**

Add to the bottom of `parser/config_test.go`:

```go
func TestConfigParser_RealFormat_HappyPath(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		wantKey  string
		wantVal  string
	}{
		{
			name:    "leading semicolon prefix with trailing semicolon",
			input:   `;System.title = "StarGazers";`,
			wantKey: "System.title",
			wantVal: "StarGazers",
		},
		{
			name:    "leading semicolon prefix without trailing semicolon",
			input:   `;scWidth = 1280`,
			wantKey: "scWidth",
			wantVal: "1280",
		},
		{
			name:    "unquoted identifier value",
			input:   `;ScreenRatio = fix;`,
			wantKey: "ScreenRatio",
			wantVal: "fix",
		},
		{
			name:    "boolean value",
			input:   `;use3D = false;`,
			wantKey: "use3D",
			wantVal: "false",
		},
		{
			name:    "dotted key",
			input:   `;scPositionX.left = 160;`,
			wantKey: "scPositionX.left",
			wantVal: "160",
		},
		{
			name:    "hex literal",
			input:   `;defaultChColor = 0xffffff;`,
			wantKey: "defaultChColor",
			wantVal: "0xffffff",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cp := NewConfigParser(false)
			config, err := cp.Parse(tt.input)
			if err != nil {
				t.Fatalf("Parse returned error: %v", err)
			}
			got, ok := config[tt.wantKey]
			if !ok {
				t.Fatalf("key %q missing from config; got: %#v", tt.wantKey, config)
			}
			if got != tt.wantVal {
				t.Errorf("config[%q] = %q, want %q", tt.wantKey, got, tt.wantVal)
			}
		})
	}
}
```

- [ ] **Step 3: Run the test — must fail**

Run: `go test ./parser/ -run TestConfigParser_RealFormat_HappyPath -v`
Expected: every subtest fails because the current parser treats `;`-prefixed lines as comments.

- [ ] **Step 4: Commit the failing test**

```bash
git add parser/config_test.go
git commit -m "$(cat <<'EOF'
test(parser): add failing tests for real Config.tjs format

Cover ;key = value; with optional trailing semicolon, dotted keys,
unquoted identifiers, booleans, hex literals.

Co-Authored-By: botamochi0x12 <botamochi0x12@gmail.com>
EOF
)"
```

---

### Task 1.2: Add tests for inline comments and edge cases

**Files:**
- Modify: `parser/config_test.go`

- [ ] **Step 1: Append further test cases**

Add to `parser/config_test.go` below the previous test:

```go
func TestConfigParser_RealFormat_InlineComments(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantKey string
		wantVal string
	}{
		{
			name:    "inline comment after value",
			input:   `;cursorDefault = default; // 通常のマウスカーソル`,
			wantKey: "cursorDefault",
			wantVal: "default",
		},
		{
			name:    "inline comment without trailing semicolon",
			input:   `;configLeft    = -1     //コンフィグアイコンの左位置を指定`,
			wantKey: "configLeft",
			wantVal: "-1",
		},
		{
			name:    "double-slash inside quoted value is preserved",
			input:   `;url = "http://example.com/path";`,
			wantKey: "url",
			wantVal: "http://example.com/path",
		},
		{
			name:    "comma-separated quoted value with embedded quotes",
			input:   `;userFace = Quicksand, "Yu Gothic", sans-serif;`,
			wantKey: "userFace",
			wantVal: `Quicksand, "Yu Gothic", sans-serif`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cp := NewConfigParser(false)
			config, err := cp.Parse(tt.input)
			if err != nil {
				t.Fatalf("Parse returned error: %v", err)
			}
			got, ok := config[tt.wantKey]
			if !ok {
				t.Fatalf("key %q missing; got: %#v", tt.wantKey, config)
			}
			if got != tt.wantVal {
				t.Errorf("config[%q] = %q, want %q", tt.wantKey, got, tt.wantVal)
			}
		})
	}
}

func TestConfigParser_RealFormat_SkipComments(t *testing.T) {
	input := `// pure comment line
// another comment
;System.title = "StarGazers";
// trailing comment`
	cp := NewConfigParser(false)
	config, err := cp.Parse(input)
	if err != nil {
		t.Fatalf("Parse returned error: %v", err)
	}
	if len(config) != 1 {
		t.Errorf("expected 1 key, got %d: %#v", len(config), config)
	}
	if config["System.title"] != "StarGazers" {
		t.Errorf("System.title = %q, want %q", config["System.title"], "StarGazers")
	}
}
```

- [ ] **Step 2: Run — must fail**

Run: `go test ./parser/ -run TestConfigParser_RealFormat -v`
Expected: every new subtest fails.

- [ ] **Step 3: Commit**

```bash
git add parser/config_test.go
git commit -m "$(cat <<'EOF'
test(parser): add Config.tjs inline-comment and skip cases

Quote-aware comment stripping must not eat // inside quoted URLs.
Pure // lines must be skipped without producing an issue.

Co-Authored-By: botamochi0x12 <botamochi0x12@gmail.com>
EOF
)"
```

---

### Task 1.3: Implement quote-aware inline comment stripper

**Files:**
- Modify: `parser/config.go` (add helper function)
- Test: `parser/config_test.go` (unit test for helper)

- [ ] **Step 1: Write a unit test for the helper**

Add to `parser/config_test.go`:

```go
func TestStripInlineComment(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want string
	}{
		{"no comment", `default`, `default`},
		{"trailing comment", `default // note`, `default`},
		{"comment after semicolon-stripped value", `default ; // note`, `default ;`},
		{"// inside double quotes", `"http://example.com"`, `"http://example.com"`},
		{"// inside single quotes", `'a//b'`, `'a//b'`},
		{"escaped quote then //", `"a\"b" // c`, `"a\"b"`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := stripInlineComment(tt.in)
			if got != tt.want {
				t.Errorf("stripInlineComment(%q) = %q, want %q", tt.in, got, tt.want)
			}
		})
	}
}
```

- [ ] **Step 2: Run — must fail (function doesn't exist)**

Run: `go test ./parser/ -run TestStripInlineComment -v`
Expected: compile error `undefined: stripInlineComment`.

- [ ] **Step 3: Implement `stripInlineComment` in `parser/config.go`**

Add this function near the bottom of `parser/config.go` (above `isWhitespace`):

```go
// stripInlineComment removes a trailing "// ..." comment from the value side
// of a config line, respecting quoted regions. Backslash-escapes are honored
// inside quotes. Returns the input unchanged if no unquoted "//" exists.
func stripInlineComment(s string) string {
	inSingle := false
	inDouble := false
	i := 0
	for i < len(s) {
		c := s[i]
		if c == '\\' && i+1 < len(s) && (inSingle || inDouble) {
			i += 2
			continue
		}
		if c == '"' && !inSingle {
			inDouble = !inDouble
			i++
			continue
		}
		if c == '\'' && !inDouble {
			inSingle = !inSingle
			i++
			continue
		}
		if !inSingle && !inDouble && c == '/' && i+1 < len(s) && s[i+1] == '/' {
			return strings.TrimRight(s[:i], " \t")
		}
		i++
	}
	return s
}
```

- [ ] **Step 4: Run — must pass**

Run: `go test ./parser/ -run TestStripInlineComment -v`
Expected: all subtests PASS.

- [ ] **Step 5: Commit**

```bash
git add parser/config.go parser/config_test.go
git commit -m "$(cat <<'EOF'
feat(parser): add quote-aware inline comment stripper

Helper for the upcoming Config.tjs rewrite. Walks left to right,
tracks single/double quote state with backslash escapes, cuts at
the first unquoted "//".

Co-Authored-By: botamochi0x12 <botamochi0x12@gmail.com>
EOF
)"
```

---

### Task 1.4: Rewrite `ParseWithResult` for real Config.tjs format

**Files:**
- Modify: `parser/config.go`

- [ ] **Step 1: Replace the body of `ParseWithResult`**

In `parser/config.go`, replace the entire `ParseWithResult` function (the block from `func (cp *ConfigParser) ParseWithResult` through the closing `}` of that method) with:

```go
// ParseWithResult parses Config.tjs content and returns both the ConfigMap and ParseResult.
// Real-format grammar:
//   ;key = value;       (semicolon prefix marks an assignment; trailing ; optional)
//   ;key = value; // c  (inline // comment after value, quote-aware)
//   // pure comment
//   /* block */         (single-line only)
//   <other lines>       (TJS code outside our subset; ignored unless looks like assignment)
func (cp *ConfigParser) ParseWithResult(content string) (types.ConfigMap, *types.ParseResult) {
	cp.result = types.NewParseResult()
	config := types.NewConfigMap()

	for lineNum, raw := range strings.Split(content, "\n") {
		trimmed := strings.TrimSpace(raw)
		if trimmed == "" {
			continue
		}
		if strings.HasPrefix(trimmed, "//") {
			continue
		}
		if strings.HasPrefix(trimmed, "/*") && strings.HasSuffix(trimmed, "*/") {
			continue
		}
		if strings.HasPrefix(trimmed, ";") {
			cp.parseConfigLine(strings.TrimSpace(trimmed[1:]), lineNum+1, config)
			if cp.options.StrictMode && cp.result.HasErrors() {
				break
			}
			continue
		}
		// Non-`;` lines that contain "=" are unrecognized assignment forms.
		// Real Config.tjs never has these, but flag as a warning so consumers
		// can spot accidental drift.
		if strings.Contains(trimmed, "=") {
			issue := types.NewParseWarning(types.InvalidConfigError, lineNum+1, 1,
				"unrecognized assignment form (missing leading ';' prefix)")
			issue.WithContext(trimmed)
			cp.result.AddIssue(*issue)
		}
	}
	return config, cp.result
}
```

- [ ] **Step 2: Replace the body of `parseConfigLine`**

In `parser/config.go`, replace the entire `parseConfigLine` function with:

```go
// parseConfigLine parses a config assignment after the leading ";" has been stripped.
// Handles inline comments, optional trailing semicolon, quoted values.
func (cp *ConfigParser) parseConfigLine(line string, lineNum int, config types.ConfigMap) {
	originalLine := line
	if line == "" {
		return
	}

	// Strip inline "// ..." comment (quote-aware) from the right.
	line = strings.TrimSpace(stripInlineComment(line))
	if line == "" {
		return
	}

	// Strip optional trailing ";".
	line = strings.TrimSuffix(line, ";")
	line = strings.TrimSpace(line)

	// Find the first "=".
	equalsIndex := strings.Index(line, "=")
	if equalsIndex == -1 {
		issue := types.NewParseError(types.InvalidConfigError, lineNum, 1,
			"missing equals sign in config line")
		issue.WithContext(originalLine)
		cp.result.AddIssue(*issue)
		return
	}

	key := strings.TrimSpace(line[:equalsIndex])
	value := strings.TrimSpace(line[equalsIndex+1:])

	if key == "" {
		issue := types.NewParseError(types.InvalidConfigError, lineNum, 1,
			"empty key in config line")
		issue.WithContext(originalLine)
		cp.result.AddIssue(*issue)
		return
	}

	if err := cp.validateQuotedValue(value, lineNum, originalLine); err != nil {
		cp.result.AddIssue(*err)
		if cp.options.StrictMode {
			return
		}
	}

	value = cp.handleQuotedValue(value)

	if value == "" {
		severity := types.Error
		if !cp.options.StrictMode {
			severity = types.Warning
		}
		issue := types.NewParseIssue(types.InvalidConfigError, severity, lineNum,
			len(originalLine), "empty value in config line")
		issue.WithContext(originalLine)
		cp.result.AddIssue(*issue)
		if cp.options.StrictMode {
			return
		}
	}

	config.Set(key, value)
}
```

- [ ] **Step 3: Run all parser tests**

Run: `go test ./parser/ -v`
Expected: all `TestConfigParser_RealFormat_*` subtests PASS, all `TestStripInlineComment` subtests PASS. Some pre-existing tests written against the OLD broken format may now fail — that is expected; address them in Task 1.5.

- [ ] **Step 4: Commit**

```bash
git add parser/config.go
git commit -m "$(cat <<'EOF'
fix(parser): rewrite Config.tjs parser for real format

Real Config.tjs uses ;key = value; as the assignment form (semicolon
is the prefix, not a comment) and // for comments. Trailing semicolon
is optional. Inline // comments after values are stripped quote-aware.

Reverts the previous fix (3847e78) which solved the example in main.go
at the cost of breaking every real Config.tjs file.

Co-Authored-By: botamochi0x12 <botamochi0x12@gmail.com>
EOF
)"
```

---

### Task 1.5: Repair pre-existing tests that assumed the old broken format

**Files:**
- Modify: `parser/config_test.go`

- [ ] **Step 1: Identify failing pre-existing tests**

Run: `go test ./parser/ -v 2>&1 | grep -E '^(--- FAIL|FAIL)' | head -30`
Note which tests reference inputs WITHOUT a leading `;` (the old format).

- [ ] **Step 2: Read each failing test and decide**

For each failing pre-existing test:
- If it tests format `key = value;` (no leading `;`), update the input to `;key = value;`.
- If it tests "missing semicolon" as an ERROR, change the assertion to expect SUCCESS (trailing `;` is now optional).
- If it tests comment lines starting with `;`, update to `//`.

Apply minimal edits. Show the diff of each test you change.

- [ ] **Step 3: Run all parser tests**

Run: `go test ./parser/ -v`
Expected: ALL tests PASS.

- [ ] **Step 4: Commit**

```bash
git add parser/config_test.go
git commit -m "$(cat <<'EOF'
test(parser): update pre-existing tests for real Config.tjs format

Update test inputs to use ;key = value; prefix, treat trailing
semicolon as optional, switch comment markers to //.

Co-Authored-By: botamochi0x12 <botamochi0x12@gmail.com>
EOF
)"
```

---

### Task 1.6: Integration test against real `StarGazers/data/system/Config.tjs`

**Files:**
- Test: `parser/config_integration_test.go` (new)

- [ ] **Step 1: Create the integration test**

Create `parser/config_integration_test.go`:

```go
package parser

import (
	"os"
	"path/filepath"
	"testing"
)

// TestConfigParser_RealStarGazersConfig parses the committed StarGazers
// fixture and asserts no errors plus a few known keys.
func TestConfigParser_RealStarGazersConfig(t *testing.T) {
	// Resolve the fixture path relative to the repo root regardless of CWD.
	wd, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	// parser/ is one level below repo root.
	repoRoot := filepath.Dir(wd)
	fixturePath := filepath.Join(repoRoot, "StarGazers", "data", "system", "Config.tjs")
	content, err := os.ReadFile(fixturePath)
	if err != nil {
		t.Skipf("fixture not present (%s): %v", fixturePath, err)
	}

	cp := NewConfigParser(false)
	config, result := cp.ParseWithResult(string(content))

	if result.HasErrors() {
		for _, e := range result.GetErrors() {
			t.Errorf("unexpected parse error: %s", e.Error())
		}
	}

	// Spot-check known keys from the fixture.
	mustEqual := map[string]string{
		"System.title":     "StarGazers",
		"projectID":        "dev.botamochi0x12.stargazers",
		"scWidth":          "1280",
		"scHeight":         "720",
		"scPositionX.left": "160",
		"defaultChColor":   "0xffffff",
	}
	for k, want := range mustEqual {
		got, ok := config[k]
		if !ok {
			t.Errorf("missing key %q", k)
			continue
		}
		if got != want {
			t.Errorf("config[%q] = %q, want %q", k, got, want)
		}
	}
}
```

- [ ] **Step 2: Run**

Run: `go test ./parser/ -run TestConfigParser_RealStarGazersConfig -v`
Expected: PASS. If it fails, examine which key is wrong; the test output names the key.

- [ ] **Step 3: Commit**

```bash
git add parser/config_integration_test.go
git commit -m "$(cat <<'EOF'
test(parser): integration test against real StarGazers Config.tjs

Parses the committed StarGazers fixture and asserts zero parse
errors plus six known key/value pairs. Skips gracefully if the
fixture is absent so the test stays portable.

Co-Authored-By: botamochi0x12 <botamochi0x12@gmail.com>
EOF
)"
```

---

### Task 1.7: Update `main.go` example content to real format

**Files:**
- Modify: `main.go`

- [ ] **Step 1: Replace the configContent literal**

In `main.go`, replace the `configContent` block (lines 44–50 in the current file) with:

```go
	// Example config content (real Config.tjs format)
	configContent := `
// Configuration file
;title = "My Game";
;width = 1280;
;height = 720
`
```

- [ ] **Step 2: Run the binary**

Run: `go run .`
Expected: stdout shows a "## Parsed config:" section with `{"title":"My Game","width":"1280","height":"720"}`.

- [ ] **Step 3: Commit**

```bash
git add main.go
git commit -m "$(cat <<'EOF'
docs(main): update example to real Config.tjs format

Use ;key = value; prefix and // comments so the example matches
what the parser now expects in real TyranoScript projects.

Co-Authored-By: botamochi0x12 <botamochi0x12@gmail.com>
EOF
)"
```

**Phase 1 done.** `go test -race ./...` should be green.

---

## Phase 2: Loader Core

**Why:** Adds file I/O without polluting `parser/`. Each loader function is testable with `testing/fstest.MapFS` so unit tests need no real disk.

**Acceptance:** `loader.DiscoverRoot`, `LoadConfigFile`, `LoadScenarioFile`, `WalkScenarios` all green with `fstest`-based tests. No new external dependencies.

---

### Task 2.1: Define `ProjectLayout` and sentinel errors

**Files:**
- Create: `loader/project.go`
- Create: `loader/project_test.go`

- [ ] **Step 1: Write the failing test**

Create `loader/project_test.go`:

```go
package loader

import (
	"errors"
	"testing"
)

func TestSentinelErrors_AreDistinct(t *testing.T) {
	if errors.Is(ErrProjectNotFound, ErrNoScenarioDir) {
		t.Error("ErrProjectNotFound and ErrNoScenarioDir must be distinct")
	}
}

func TestProjectLayout_FieldsExist(t *testing.T) {
	p := &ProjectLayout{
		Root:        "/tmp/proj",
		ScenarioDir: "/tmp/proj/data/scenario",
		ConfigPath:  "/tmp/proj/data/system/Config.tjs",
	}
	if p.Root == "" || p.ScenarioDir == "" || p.ConfigPath == "" {
		t.Error("expected all three fields populated")
	}
}
```

- [ ] **Step 2: Run — must fail (package doesn't exist)**

Run: `go test ./loader/ -v`
Expected: compile error `loader/project_test.go: no such package`.

- [ ] **Step 3: Implement `loader/project.go`**

Create `loader/project.go`:

```go
// Package loader handles file I/O and project layout discovery for
// TyranoScript projects. It wraps the pure parsers in package parser/.
package loader

import "errors"

// ProjectLayout describes a TyranoScript project on disk.
type ProjectLayout struct {
	Root        string // absolute project root (contains data/)
	ScenarioDir string // <root>/data/scenario
	ConfigPath  string // <root>/data/system/Config.tjs (may not exist)
}

// Sentinel errors returned by discovery and loading functions.
var (
	ErrProjectNotFound = errors.New("loader: no TyranoScript project found")
	ErrNoScenarioDir   = errors.New("loader: project missing data/scenario directory")
)
```

- [ ] **Step 4: Run — must pass**

Run: `go test ./loader/ -v`
Expected: both tests PASS.

- [ ] **Step 5: Commit**

```bash
git add loader/project.go loader/project_test.go
git commit -m "$(cat <<'EOF'
feat(loader): add ProjectLayout and sentinel errors

Foundation for the loader package. Pure types only; no I/O yet.

Co-Authored-By: botamochi0x12 <botamochi0x12@gmail.com>
EOF
)"
```

---

### Task 2.2: Implement `DiscoverRoot` with upward walk

**Files:**
- Modify: `loader/project.go`
- Modify: `loader/project_test.go`

- [ ] **Step 1: Write failing tests using `fstest.MapFS`**

Append to `loader/project_test.go`:

```go
import (
	"errors"
	"path/filepath"
	"testing"
	"testing/fstest"
)

func TestDiscoverRoot_FindsScenarioMarker(t *testing.T) {
	fsys := fstest.MapFS{
		"proj/data/scenario/first.ks":    {Data: []byte("")},
		"proj/data/system/Config.tjs":    {Data: []byte("")},
		"proj/sub/deeper/note.txt":       {Data: []byte("")},
	}
	got, err := discoverRootFS(fsys, "proj/sub/deeper")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.Root != "proj" {
		t.Errorf("Root = %q, want %q", got.Root, "proj")
	}
	if got.ScenarioDir != filepath.Join("proj", "data", "scenario") {
		t.Errorf("ScenarioDir = %q", got.ScenarioDir)
	}
	if got.ConfigPath != filepath.Join("proj", "data", "system", "Config.tjs") {
		t.Errorf("ConfigPath = %q", got.ConfigPath)
	}
}

func TestDiscoverRoot_NotFound(t *testing.T) {
	fsys := fstest.MapFS{
		"unrelated/file.txt": {Data: []byte("")},
	}
	_, err := discoverRootFS(fsys, "unrelated")
	if !errors.Is(err, ErrProjectNotFound) {
		t.Errorf("expected ErrProjectNotFound, got %v", err)
	}
}

func TestDiscoverRoot_FindsConfigOnlyMarker(t *testing.T) {
	fsys := fstest.MapFS{
		"proj/data/system/Config.tjs": {Data: []byte("")},
	}
	got, err := discoverRootFS(fsys, "proj")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.Root != "proj" {
		t.Errorf("Root = %q, want %q", got.Root, "proj")
	}
}
```

- [ ] **Step 2: Run — must fail (function doesn't exist)**

Run: `go test ./loader/ -v -run TestDiscoverRoot`
Expected: compile error `undefined: discoverRootFS`.

- [ ] **Step 3: Implement discovery**

Append to `loader/project.go`:

```go
import (
	"io/fs"
	"os"
	"path/filepath"
)

// DiscoverRoot walks upward from start (a directory path) until it finds
// data/scenario/ or data/system/Config.tjs. Returns ErrProjectNotFound if
// neither marker is found by filesystem root.
func DiscoverRoot(start string) (*ProjectLayout, error) {
	abs, err := filepath.Abs(start)
	if err != nil {
		return nil, fmt.Errorf("loader: resolve start path %q: %w", start, err)
	}
	return discoverRootFS(osFS{}, abs)
}

// discoverRootFS is the testable core of DiscoverRoot. It uses an fs.FS
// abstraction so unit tests can pass testing/fstest.MapFS.
func discoverRootFS(fsys fs.FS, start string) (*ProjectLayout, error) {
	cur := start
	for {
		if hasScenarioOrConfig(fsys, cur) {
			return layoutAt(cur), nil
		}
		parent := filepath.Dir(cur)
		if parent == cur {
			return nil, ErrProjectNotFound
		}
		cur = parent
	}
}

func hasScenarioOrConfig(fsys fs.FS, dir string) bool {
	if statDir(fsys, filepath.Join(dir, "data", "scenario")) {
		return true
	}
	if statFile(fsys, filepath.Join(dir, "data", "system", "Config.tjs")) {
		return true
	}
	return false
}

func layoutAt(root string) *ProjectLayout {
	return &ProjectLayout{
		Root:        root,
		ScenarioDir: filepath.Join(root, "data", "scenario"),
		ConfigPath:  filepath.Join(root, "data", "system", "Config.tjs"),
	}
}

func statDir(fsys fs.FS, p string) bool {
	info, err := fs.Stat(fsys, p)
	return err == nil && info.IsDir()
}

func statFile(fsys fs.FS, p string) bool {
	info, err := fs.Stat(fsys, p)
	return err == nil && !info.IsDir()
}

// osFS adapts the operating-system filesystem to fs.FS using absolute paths.
type osFS struct{}

func (osFS) Open(name string) (fs.File, error) { return os.Open(name) }
```

Add `"fmt"` to the existing import block at the top of `loader/project.go` if not already present.

- [ ] **Step 4: Run — must pass**

Run: `go test ./loader/ -v -run TestDiscoverRoot`
Expected: all three subtests PASS.

- [ ] **Step 5: Commit**

```bash
git add loader/project.go loader/project_test.go
git commit -m "$(cat <<'EOF'
feat(loader): implement DiscoverRoot with upward walk

Walks upward from a starting directory until it finds either
data/scenario/ or data/system/Config.tjs. Falls back to
ErrProjectNotFound at the filesystem root.

The fs.FS-based discoverRootFS is exposed package-internally so
unit tests can use testing/fstest.MapFS without touching disk.

Co-Authored-By: botamochi0x12 <botamochi0x12@gmail.com>
EOF
)"
```

---

### Task 2.3: Implement `LayoutFrom` for explicit root

**Files:**
- Modify: `loader/project.go`
- Modify: `loader/project_test.go`

- [ ] **Step 1: Write failing tests**

Append to `loader/project_test.go`:

```go
func TestLayoutFrom_ValidProject(t *testing.T) {
	tmp := t.TempDir()
	if err := os.MkdirAll(filepath.Join(tmp, "data", "scenario"), 0o755); err != nil {
		t.Fatal(err)
	}
	got, err := LayoutFrom(tmp)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.ScenarioDir != filepath.Join(tmp, "data", "scenario") {
		t.Errorf("ScenarioDir = %q", got.ScenarioDir)
	}
}

func TestLayoutFrom_MissingScenarioDir(t *testing.T) {
	tmp := t.TempDir()
	_, err := LayoutFrom(tmp)
	if !errors.Is(err, ErrNoScenarioDir) {
		t.Errorf("expected ErrNoScenarioDir, got %v", err)
	}
}
```

Add `"os"` to test file imports if missing.

- [ ] **Step 2: Run — must fail**

Run: `go test ./loader/ -run TestLayoutFrom -v`
Expected: compile error `undefined: LayoutFrom`.

- [ ] **Step 3: Implement**

Append to `loader/project.go`:

```go
// LayoutFrom constructs a ProjectLayout for an explicit root (no upward walk).
// Returns ErrNoScenarioDir if <root>/data/scenario does not exist.
func LayoutFrom(root string) (*ProjectLayout, error) {
	abs, err := filepath.Abs(root)
	if err != nil {
		return nil, fmt.Errorf("loader: resolve root %q: %w", root, err)
	}
	scenarioDir := filepath.Join(abs, "data", "scenario")
	info, err := os.Stat(scenarioDir)
	if err != nil || !info.IsDir() {
		return nil, ErrNoScenarioDir
	}
	return layoutAt(abs), nil
}
```

- [ ] **Step 4: Run — must pass**

Run: `go test ./loader/ -run TestLayoutFrom -v`
Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add loader/project.go loader/project_test.go
git commit -m "$(cat <<'EOF'
feat(loader): add LayoutFrom for explicit project roots

Bypasses the upward walk used by DiscoverRoot. Used by the CLI
when --project-root is passed.

Co-Authored-By: botamochi0x12 <botamochi0x12@gmail.com>
EOF
)"
```

---

### Task 2.4: Implement `LoadConfigFile`

**Files:**
- Create: `loader/config_loader.go`
- Create: `loader/config_loader_test.go`

- [ ] **Step 1: Write failing test**

Create `loader/config_loader_test.go`:

```go
package loader

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/botamochi0x12/tyrano-parser-go/parser"
)

func TestLoadConfigFile_HappyPath(t *testing.T) {
	tmp := t.TempDir()
	cfgDir := filepath.Join(tmp, "data", "system")
	if err := os.MkdirAll(cfgDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(cfgDir, "Config.tjs"),
		[]byte(`;title = "Hello";`), 0o644); err != nil {
		t.Fatal(err)
	}
	scnDir := filepath.Join(tmp, "data", "scenario")
	if err := os.MkdirAll(scnDir, 0o755); err != nil {
		t.Fatal(err)
	}

	layout, err := LayoutFrom(tmp)
	if err != nil {
		t.Fatal(err)
	}
	cp := parser.NewConfigParser(false)
	config, result, err := LoadConfigFile(layout, cp)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if config["title"] != "Hello" {
		t.Errorf("config[title] = %q, want %q", config["title"], "Hello")
	}
	if result.HasErrors() {
		t.Errorf("unexpected parse errors: %v", result.GetErrors())
	}
}

func TestLoadConfigFile_MissingFile(t *testing.T) {
	tmp := t.TempDir()
	scnDir := filepath.Join(tmp, "data", "scenario")
	if err := os.MkdirAll(scnDir, 0o755); err != nil {
		t.Fatal(err)
	}
	layout, err := LayoutFrom(tmp)
	if err != nil {
		t.Fatal(err)
	}
	cp := parser.NewConfigParser(false)
	config, result, err := LoadConfigFile(layout, cp)
	if err != nil {
		t.Fatalf("expected nil error for missing config, got %v", err)
	}
	if config != nil {
		t.Errorf("expected nil config for missing file, got %#v", config)
	}
	if result != nil {
		t.Errorf("expected nil result for missing file, got %#v", result)
	}
}
```

- [ ] **Step 2: Run — must fail**

Run: `go test ./loader/ -run TestLoadConfigFile -v`
Expected: `undefined: LoadConfigFile`.

- [ ] **Step 3: Implement**

Create `loader/config_loader.go`:

```go
package loader

import (
	"errors"
	"fmt"
	"io/fs"
	"os"

	"github.com/botamochi0x12/tyrano-parser-go/parser"
	"github.com/botamochi0x12/tyrano-parser-go/types"
)

// LoadConfigFile reads layout.ConfigPath, parses it with cp, and returns
// the ConfigMap plus ParseResult. Returns (nil, nil, nil) if the file does
// not exist — config is optional in TyranoScript projects.
func LoadConfigFile(layout *ProjectLayout, cp *parser.ConfigParser) (
	types.ConfigMap, *types.ParseResult, error) {
	content, err := os.ReadFile(layout.ConfigPath)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return nil, nil, nil
		}
		return nil, nil, fmt.Errorf("loader: read config %s: %w", layout.ConfigPath, err)
	}
	config, result := cp.ParseWithResult(string(content))
	return config, result, nil
}
```

- [ ] **Step 4: Run — must pass**

Run: `go test ./loader/ -run TestLoadConfigFile -v`
Expected: both subtests PASS.

- [ ] **Step 5: Commit**

```bash
git add loader/config_loader.go loader/config_loader_test.go
git commit -m "$(cat <<'EOF'
feat(loader): add LoadConfigFile

Reads <root>/data/system/Config.tjs and runs it through ConfigParser.
Returns (nil, nil, nil) when the file is absent — Config.tjs is
optional, missing config should not be an I/O error.

Co-Authored-By: botamochi0x12 <botamochi0x12@gmail.com>
EOF
)"
```

---

### Task 2.5: Implement `LoadScenarioFile`

**Files:**
- Create: `loader/scenario_loader.go`
- Create: `loader/scenario_loader_test.go`

- [ ] **Step 1: Write failing test**

Create `loader/scenario_loader_test.go`:

```go
package loader

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/botamochi0x12/tyrano-parser-go/parser"
)

func TestLoadScenarioFile_HappyPath(t *testing.T) {
	tmp := t.TempDir()
	path := filepath.Join(tmp, "first.ks")
	content := `*start
[cm]
Hello[p]
[s]
`
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}

	tp := parser.NewDefaultTyranoParser()
	scenario, result, err := LoadScenarioFile(path, tp)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if scenario == nil {
		t.Fatal("scenario is nil")
	}
	if _, ok := scenario.Labels["start"]; !ok {
		t.Errorf("expected label *start in Labels, got %#v", scenario.Labels)
	}
	if result.HasErrors() {
		t.Errorf("unexpected parse errors: %v", result.GetErrors())
	}
}

func TestLoadScenarioFile_FileNotFound(t *testing.T) {
	tp := parser.NewDefaultTyranoParser()
	_, _, err := LoadScenarioFile("/no/such/file.ks", tp)
	if err == nil {
		t.Fatal("expected error for missing file, got nil")
	}
}
```

- [ ] **Step 2: Run — must fail**

Run: `go test ./loader/ -run TestLoadScenarioFile -v`
Expected: `undefined: LoadScenarioFile`.

- [ ] **Step 3: Implement**

Create `loader/scenario_loader.go`:

```go
package loader

import (
	"fmt"
	"os"

	"github.com/botamochi0x12/tyrano-parser-go/parser"
	"github.com/botamochi0x12/tyrano-parser-go/types"
)

// LoadScenarioFile reads a .ks file and parses it with tp.
func LoadScenarioFile(path string, tp *parser.TyranoParser) (
	*types.ParsedScenario, *types.ParseResult, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return nil, nil, fmt.Errorf("loader: read scenario %s: %w", path, err)
	}
	scenario, result := tp.ParseScenarioWithResult(string(content))
	return scenario, result, nil
}
```

- [ ] **Step 4: Run — must pass**

Run: `go test ./loader/ -run TestLoadScenarioFile -v`
Expected: both subtests PASS.

- [ ] **Step 5: Commit**

```bash
git add loader/scenario_loader.go loader/scenario_loader_test.go
git commit -m "$(cat <<'EOF'
feat(loader): add LoadScenarioFile

Reads a single .ks file and runs it through TyranoParser. I/O errors
are wrapped with %w; parse issues stay in *ParseResult per the
loader contract.

Co-Authored-By: botamochi0x12 <botamochi0x12@gmail.com>
EOF
)"
```

---

### Task 2.6: Implement `WalkScenarios`

**Files:**
- Modify: `loader/scenario_loader.go`
- Modify: `loader/scenario_loader_test.go`

- [ ] **Step 1: Write failing tests**

Append to `loader/scenario_loader_test.go`:

```go
func TestWalkScenarios_FindsAllKsFiles(t *testing.T) {
	tmp := t.TempDir()
	scnDir := filepath.Join(tmp, "data", "scenario")
	if err := os.MkdirAll(scnDir, 0o755); err != nil {
		t.Fatal(err)
	}
	files := map[string]string{
		"first.ks":  `*start` + "\n[s]\n",
		"title.ks":  `*title` + "\n[s]\n",
		"README.md": "ignore me",
		"data.json": "{}",
	}
	for name, body := range files {
		if err := os.WriteFile(filepath.Join(scnDir, name), []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	layout, err := LayoutFrom(tmp)
	if err != nil {
		t.Fatal(err)
	}
	tp := parser.NewDefaultTyranoParser()
	scenarios, result, err := WalkScenarios(layout, tp)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(scenarios) != 2 {
		t.Errorf("expected 2 .ks files, got %d: %#v", len(scenarios), keys(scenarios))
	}
	if _, ok := scenarios["first.ks"]; !ok {
		t.Errorf("missing first.ks in result")
	}
	if _, ok := scenarios["title.ks"]; !ok {
		t.Errorf("missing title.ks in result")
	}
	if result.HasErrors() {
		t.Errorf("unexpected errors: %v", result.GetErrors())
	}
}

func keys[K comparable, V any](m map[K]V) []K {
	out := make([]K, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	return out
}
```

- [ ] **Step 2: Run — must fail**

Run: `go test ./loader/ -run TestWalkScenarios -v`
Expected: `undefined: WalkScenarios`.

- [ ] **Step 3: Implement**

Append to `loader/scenario_loader.go`:

```go
import (
	"io/fs"
	"path/filepath"
	"strings"
)

// WalkScenarios walks layout.ScenarioDir and parses every *.ks file found.
// Returns a map keyed by path RELATIVE to ScenarioDir. Parse issues from
// each file are merged into a single ParseResult.
func WalkScenarios(layout *ProjectLayout, tp *parser.TyranoParser) (
	map[string]*types.ParsedScenario, *types.ParseResult, error) {
	out := make(map[string]*types.ParsedScenario)
	combined := types.NewParseResult()

	err := filepath.WalkDir(layout.ScenarioDir, func(path string, d fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if d.IsDir() {
			return nil
		}
		if !strings.HasSuffix(strings.ToLower(d.Name()), ".ks") {
			return nil
		}
		scenario, result, loadErr := LoadScenarioFile(path, tp)
		if loadErr != nil {
			return loadErr
		}
		rel, relErr := filepath.Rel(layout.ScenarioDir, path)
		if relErr != nil {
			return fmt.Errorf("loader: rel path for %s: %w", path, relErr)
		}
		out[rel] = scenario
		if result != nil {
			combined.Merge(result)
		}
		return nil
	})
	if err != nil {
		return nil, nil, fmt.Errorf("loader: walk scenarios %s: %w", layout.ScenarioDir, err)
	}
	return out, combined, nil
}
```

Make sure all required imports (`io/fs`, `path/filepath`, `strings`) are added to the import block at the top.

- [ ] **Step 4: Run — must pass**

Run: `go test ./loader/ -run TestWalkScenarios -v`
Expected: PASS. Then run full loader suite: `go test ./loader/ -v -race`. Expected: all green.

- [ ] **Step 5: Commit**

```bash
git add loader/scenario_loader.go loader/scenario_loader_test.go
git commit -m "$(cat <<'EOF'
feat(loader): add WalkScenarios

Walks <root>/data/scenario/ and parses every .ks file. Returns a
map keyed by path relative to the scenario directory plus a merged
ParseResult covering every file. Non-.ks files are skipped silently.

Co-Authored-By: botamochi0x12 <botamochi0x12@gmail.com>
EOF
)"
```

**Phase 2 done.** `go test -race ./...` should be green.

---

## Phase 3: Refs + Scan

**Why:** Adds cross-reference extraction and the project-wide scan that ties everything together. After this phase, programmatic project scanning works without a CLI.

**Acceptance:** `loader.ScanProject` parses all scenarios, extracts `@call`/`@jump`/`@link` storage and target refs, flags missing storage and missing labels as issues. Integration test against `StarGazers/` passes.

---

### Task 3.1: Add `MissingStorageError` and `MissingLabelError` to `types/errors.go`

**Files:**
- Modify: `types/errors.go`
- Test: `types/errors_test.go`

- [ ] **Step 1: Write failing test**

Append to `types/errors_test.go`:

```go
func TestErrorType_NewConstants(t *testing.T) {
	if MissingStorageError.String() != "MissingStorageError" {
		t.Errorf("MissingStorageError.String() = %q", MissingStorageError.String())
	}
	if MissingLabelError.String() != "MissingLabelError" {
		t.Errorf("MissingLabelError.String() = %q", MissingLabelError.String())
	}
	if MissingStorageError == MissingLabelError {
		t.Error("constants must be distinct")
	}
}
```

- [ ] **Step 2: Run — must fail**

Run: `go test ./types/ -run TestErrorType_NewConstants -v`
Expected: `undefined: MissingStorageError`.

- [ ] **Step 3: Add constants**

In `types/errors.go`, modify the `const (...)` block at the top (lines 8–17 currently) to append the two new constants:

```go
const (
	SyntaxError ErrorType = iota
	DuplicateLabelError
	UnmatchedIfError
	InvalidConfigError
	MalformedTagError
	UnknownTagError
	MissingParameterError
	InvalidParameterError
	MissingStorageError // referenced .ks file not found in project
	MissingLabelError   // referenced *label not found in target scenario
)
```

In the `String()` method (lines 20–41), add cases:

```go
	case MissingStorageError:
		return "MissingStorageError"
	case MissingLabelError:
		return "MissingLabelError"
```

(insert before the `default` case).

- [ ] **Step 4: Run — must pass**

Run: `go test ./types/ -v`
Expected: all PASS.

- [ ] **Step 5: Commit**

```bash
git add types/errors.go types/errors_test.go
git commit -m "$(cat <<'EOF'
feat(types): add MissingStorageError and MissingLabelError

For cross-reference resolution in the loader scan: distinguish
"referenced .ks not found" from "referenced *label not found"
so consumers can present clear errors.

Co-Authored-By: botamochi0x12 <botamochi0x12@gmail.com>
EOF
)"
```

---

### Task 3.2: Define `ScenarioRef` and implement `ExtractRefs`

**Files:**
- Create: `loader/refs.go`
- Create: `loader/refs_test.go`

- [ ] **Step 1: Write failing test**

Create `loader/refs_test.go`:

```go
package loader

import (
	"testing"

	"github.com/botamochi0x12/tyrano-parser-go/parser"
)

func TestExtractRefs_CallAndJumpStorage(t *testing.T) {
	src := `*start
@call storage="tyrano.ks"
@jump storage="title.ks"
@jump target="*ending"
[s]
`
	tp := parser.NewDefaultTyranoParser()
	scenario, _ := tp.ParseScenarioWithResult(src)
	refs := ExtractRefs("first.ks", scenario)

	if len(refs) != 3 {
		t.Fatalf("expected 3 refs, got %d: %#v", len(refs), refs)
	}

	got := map[string]ScenarioRef{}
	for _, r := range refs {
		got[r.Tag+":"+r.Storage+":"+r.Target] = r
	}
	if _, ok := got["call:tyrano.ks:"]; !ok {
		t.Errorf("missing @call storage=tyrano.ks in refs: %#v", refs)
	}
	if _, ok := got["jump:title.ks:"]; !ok {
		t.Errorf("missing @jump storage=title.ks in refs: %#v", refs)
	}
	if _, ok := got["jump::*ending"]; !ok {
		t.Errorf("missing @jump target=*ending in refs: %#v", refs)
	}
	for _, r := range refs {
		if r.From != "first.ks" {
			t.Errorf("unexpected From: %q", r.From)
		}
	}
}

func TestExtractRefs_LinkTag(t *testing.T) {
	src := `[link storage="other.ks" target="*x"]choose me[endlink]
[s]
`
	tp := parser.NewDefaultTyranoParser()
	scenario, _ := tp.ParseScenarioWithResult(src)
	refs := ExtractRefs("menu.ks", scenario)
	if len(refs) != 1 {
		t.Fatalf("expected 1 link ref, got %d", len(refs))
	}
	r := refs[0]
	if r.Tag != "link" || r.Storage != "other.ks" || r.Target != "*x" {
		t.Errorf("unexpected ref: %#v", r)
	}
}
```

- [ ] **Step 2: Run — must fail**

Run: `go test ./loader/ -run TestExtractRefs -v`
Expected: `undefined: ScenarioRef` / `undefined: ExtractRefs`.

- [ ] **Step 3: Implement**

Create `loader/refs.go`:

```go
package loader

import "github.com/botamochi0x12/tyrano-parser-go/types"

// ScenarioRef represents a cross-file or cross-label reference produced by
// @call, @jump, or @link tags carrying storage= and/or target= parameters.
type ScenarioRef struct {
	From    string `json:"from"`    // source file (relative to ScenarioDir)
	Line    int    `json:"line"`
	Tag     string `json:"tag"`     // "call" | "jump" | "link"
	Storage string `json:"storage"` // empty when only target= is given
	Target  string `json:"target"`  // empty when only storage= is given
}

// refTags lists tag names whose storage=/target= parameters produce refs.
var refTags = map[string]bool{
	"call": true,
	"jump": true,
	"link": true,
}

// ExtractRefs scans a parsed scenario for @call/@jump/@link tags and
// returns one ScenarioRef per occurrence that carries storage= or target=.
// Tags without either parameter are ignored.
func ExtractRefs(from string, scenario *types.ParsedScenario) []ScenarioRef {
	if scenario == nil {
		return nil
	}
	out := make([]ScenarioRef, 0)
	for _, tag := range scenario.Elements {
		if !refTags[tag.Name] {
			continue
		}
		storage := tag.Parameters["storage"]
		target := tag.Parameters["target"]
		if storage == "" && target == "" {
			continue
		}
		out = append(out, ScenarioRef{
			From:    from,
			Line:    tag.Line,
			Tag:     tag.Name,
			Storage: storage,
			Target:  target,
		})
	}
	return out
}
```

- [ ] **Step 4: Run — must pass**

Run: `go test ./loader/ -run TestExtractRefs -v`
Expected: both subtests PASS.

- [ ] **Step 5: Commit**

```bash
git add loader/refs.go loader/refs_test.go
git commit -m "$(cat <<'EOF'
feat(loader): add ExtractRefs and ScenarioRef

Walks a parsed scenario's elements and emits one ScenarioRef per
@call/@jump/@link tag carrying storage= or target=. Foundation for
project-wide cross-reference resolution.

Co-Authored-By: botamochi0x12 <botamochi0x12@gmail.com>
EOF
)"
```

---

### Task 3.3: Define `ProjectScan` type

**Files:**
- Create: `types/project_scan.go`

- [ ] **Step 1: Write the type**

Create `types/project_scan.go`:

```go
package types

// ProjectScan is the aggregate result of scanning a TyranoScript project:
// parsed config, all parsed scenarios, every cross-reference, and merged
// parse + reference issues.
type ProjectScan struct {
	Root      string                     `json:"root"`
	Config    ConfigMap                  `json:"config"`
	Scenarios map[string]*ParsedScenario `json:"scenarios"`
	Refs      []ScenarioRefRecord        `json:"refs"`
	Issues    *ParseResult               `json:"issues"`
}

// ScenarioRefRecord is the JSON-friendly mirror of loader.ScenarioRef. The
// loader package owns ref extraction; types/ holds the wire format so the
// CLI can marshal a complete ProjectScan without importing loader/.
type ScenarioRefRecord struct {
	From    string `json:"from"`
	Line    int    `json:"line"`
	Tag     string `json:"tag"`
	Storage string `json:"storage"`
	Target  string `json:"target"`
}
```

- [ ] **Step 2: Confirm it compiles**

Run: `go build ./types/`
Expected: no output (success).

- [ ] **Step 3: Commit**

```bash
git add types/project_scan.go
git commit -m "$(cat <<'EOF'
feat(types): add ProjectScan and ScenarioRefRecord

Wire-format types for the upcoming loader.ScanProject and the CLI's
JSON renderer. Mirror of loader.ScenarioRef lives here so JSON
encoding of a ProjectScan does not pull loader into the CLI.

Co-Authored-By: botamochi0x12 <botamochi0x12@gmail.com>
EOF
)"
```

---

### Task 3.4: Implement `ScanProject` (full walk)

**Files:**
- Create: `loader/scan.go`
- Create: `loader/scan_test.go`

- [ ] **Step 1: Write failing test**

Create `loader/scan_test.go`:

```go
package loader

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/botamochi0x12/tyrano-parser-go/parser"
)

func writeFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

func TestScanProject_AllScenariosResolved(t *testing.T) {
	tmp := t.TempDir()
	scn := filepath.Join(tmp, "data", "scenario")
	writeFile(t, filepath.Join(scn, "first.ks"),
		`*start
@call storage="tyrano.ks"
@jump storage="title.ks"
[s]
`)
	writeFile(t, filepath.Join(scn, "tyrano.ks"), "*x\n[s]\n")
	writeFile(t, filepath.Join(scn, "title.ks"), "*y\n[s]\n")

	layout, err := LayoutFrom(tmp)
	if err != nil {
		t.Fatal(err)
	}
	tp := parser.NewDefaultTyranoParser()
	scan, err := ScanProject(layout, tp)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(scan.Scenarios) != 3 {
		t.Errorf("expected 3 scenarios, got %d", len(scan.Scenarios))
	}
	if len(scan.Refs) != 2 {
		t.Errorf("expected 2 refs, got %d", len(scan.Refs))
	}
	if scan.Issues.HasErrors() {
		t.Errorf("unexpected errors: %v", scan.Issues.GetErrors())
	}
}

func TestScanProject_FlagsMissingStorage(t *testing.T) {
	tmp := t.TempDir()
	scn := filepath.Join(tmp, "data", "scenario")
	writeFile(t, filepath.Join(scn, "first.ks"),
		`*start
@jump storage="missing.ks"
[s]
`)

	layout, err := LayoutFrom(tmp)
	if err != nil {
		t.Fatal(err)
	}
	tp := parser.NewDefaultTyranoParser()
	scan, err := ScanProject(layout, tp)
	if err != nil {
		t.Fatal(err)
	}
	errs := scan.Issues.GetErrors()
	if len(errs) != 1 {
		t.Fatalf("expected 1 missing-storage error, got %d: %v", len(errs), errs)
	}
	// The new ErrorType added in Task 3.1.
	wantTypeStr := "MissingStorageError"
	if errs[0].Type.String() != wantTypeStr {
		t.Errorf("error type = %q, want %q", errs[0].Type.String(), wantTypeStr)
	}
}
```

- [ ] **Step 2: Run — must fail**

Run: `go test ./loader/ -run TestScanProject -v`
Expected: `undefined: ScanProject`.

- [ ] **Step 3: Implement**

Create `loader/scan.go`:

```go
package loader

import (
	"fmt"

	"github.com/botamochi0x12/tyrano-parser-go/parser"
	"github.com/botamochi0x12/tyrano-parser-go/types"
)

// ScanProject performs a full walk: parses the config (if present), every
// .ks file in ScenarioDir, then resolves every cross-reference. Missing
// storage and missing labels become issues in the returned ProjectScan.
func ScanProject(layout *ProjectLayout, tp *parser.TyranoParser) (*types.ProjectScan, error) {
	scan := &types.ProjectScan{
		Root:      layout.Root,
		Scenarios: nil,
		Refs:      nil,
		Issues:    types.NewParseResult(),
	}

	// 1. Config (optional).
	cp := parser.NewConfigParser(false)
	config, configResult, err := LoadConfigFile(layout, cp)
	if err != nil {
		return nil, fmt.Errorf("loader: scan config: %w", err)
	}
	scan.Config = config
	if configResult != nil {
		scan.Issues.Merge(configResult)
	}

	// 2. All scenarios.
	scenarios, walkResult, err := WalkScenarios(layout, tp)
	if err != nil {
		return nil, fmt.Errorf("loader: scan scenarios: %w", err)
	}
	scan.Scenarios = scenarios
	scan.Issues.Merge(walkResult)

	// 3. Extract refs from every scenario.
	allRefs := make([]ScenarioRef, 0)
	for relPath, scenario := range scenarios {
		allRefs = append(allRefs, ExtractRefs(relPath, scenario)...)
	}

	// 4. Resolve each ref against the parsed scenario set.
	for _, r := range allRefs {
		resolveRef(r, scenarios, scan.Issues)
	}

	// 5. Convert refs to wire-format.
	scan.Refs = make([]types.ScenarioRefRecord, 0, len(allRefs))
	for _, r := range allRefs {
		scan.Refs = append(scan.Refs, types.ScenarioRefRecord{
			From: r.From, Line: r.Line, Tag: r.Tag,
			Storage: r.Storage, Target: r.Target,
		})
	}

	return scan, nil
}

// resolveRef checks one reference's storage and target against the parsed
// scenario set, appending issues for any that don't resolve.
func resolveRef(r ScenarioRef, scenarios map[string]*types.ParsedScenario, issues *types.ParseResult) {
	// Determine which scenario's labels to check for the target.
	var labelHost *types.ParsedScenario
	if r.Storage != "" {
		host, ok := scenarios[r.Storage]
		if !ok {
			issue := types.NewParseError(types.MissingStorageError, r.Line, 1,
				fmt.Sprintf("referenced storage %q not found", r.Storage))
			issue.WithContext(fmt.Sprintf("from=%s tag=%s", r.From, r.Tag))
			issues.AddIssue(*issue)
			return
		}
		labelHost = host
	} else {
		labelHost = scenarios[r.From]
	}

	if r.Target == "" {
		return
	}
	labelName := r.Target
	if len(labelName) > 0 && labelName[0] == '*' {
		labelName = labelName[1:]
	}
	if labelHost == nil {
		return
	}
	if _, ok := labelHost.Labels[labelName]; !ok {
		issue := types.NewParseError(types.MissingLabelError, r.Line, 1,
			fmt.Sprintf("referenced label %q not found", r.Target))
		issue.WithContext(fmt.Sprintf("from=%s tag=%s storage=%s", r.From, r.Tag, r.Storage))
		issues.AddIssue(*issue)
	}
}
```

- [ ] **Step 4: Run — must pass**

Run: `go test ./loader/ -run TestScanProject -v`
Expected: both subtests PASS. Then run `go test -race ./loader/`. Expected: green.

- [ ] **Step 5: Commit**

```bash
git add loader/scan.go loader/scan_test.go
git commit -m "$(cat <<'EOF'
feat(loader): add ScanProject for full project scans

Parses Config.tjs, walks all .ks files, extracts refs, resolves each
ref against the parsed scenario set. Missing storage and missing
labels become MissingStorageError / MissingLabelError issues.

Co-Authored-By: botamochi0x12 <botamochi0x12@gmail.com>
EOF
)"
```

---

### Task 3.5: Implement `ResolveFrom` (BFS from entrypoint)

**Files:**
- Modify: `loader/scan.go`
- Modify: `loader/scan_test.go`

- [ ] **Step 1: Write failing test**

Append to `loader/scan_test.go`:

```go
func TestResolveFrom_OnlyVisitsReachable(t *testing.T) {
	tmp := t.TempDir()
	scn := filepath.Join(tmp, "data", "scenario")
	writeFile(t, filepath.Join(scn, "first.ks"),
		`*start
@call storage="tyrano.ks"
[s]
`)
	writeFile(t, filepath.Join(scn, "tyrano.ks"), "*x\n[s]\n")
	writeFile(t, filepath.Join(scn, "unreached.ks"), "*z\n[s]\n")

	layout, err := LayoutFrom(tmp)
	if err != nil {
		t.Fatal(err)
	}
	tp := parser.NewDefaultTyranoParser()
	scan, err := ResolveFrom(layout, tp, "first.ks")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(scan.Scenarios) != 2 {
		t.Errorf("expected 2 reachable scenarios, got %d: %v",
			len(scan.Scenarios), keys(scan.Scenarios))
	}
	if _, ok := scan.Scenarios["unreached.ks"]; ok {
		t.Errorf("unreached.ks should not have been parsed")
	}
}
```

- [ ] **Step 2: Run — must fail**

Run: `go test ./loader/ -run TestResolveFrom -v`
Expected: `undefined: ResolveFrom`.

- [ ] **Step 3: Implement**

Append to `loader/scan.go`:

```go
import (
	"path/filepath"
)

// ResolveFrom is a BFS variant of ScanProject: it parses only scenarios
// reachable from entrypoint via @call/@jump/@link storage= chains.
// entrypoint may be an absolute path, a path relative to CWD, or just a
// filename inside ScenarioDir.
func ResolveFrom(layout *ProjectLayout, tp *parser.TyranoParser, entrypoint string) (*types.ProjectScan, error) {
	scan := &types.ProjectScan{
		Root:      layout.Root,
		Scenarios: make(map[string]*types.ParsedScenario),
		Refs:      nil,
		Issues:    types.NewParseResult(),
	}

	// Config (optional).
	cp := parser.NewConfigParser(false)
	config, configResult, err := LoadConfigFile(layout, cp)
	if err != nil {
		return nil, fmt.Errorf("loader: resolve config: %w", err)
	}
	scan.Config = config
	if configResult != nil {
		scan.Issues.Merge(configResult)
	}

	// Normalize entrypoint to a relative path inside ScenarioDir.
	startRel, err := relativizeScenario(layout, entrypoint)
	if err != nil {
		return nil, err
	}

	queue := []string{startRel}
	visited := map[string]bool{}
	allRefs := make([]ScenarioRef, 0)

	for len(queue) > 0 {
		rel := queue[0]
		queue = queue[1:]
		if visited[rel] {
			continue
		}
		visited[rel] = true

		path := filepath.Join(layout.ScenarioDir, rel)
		scenario, parseResult, err := LoadScenarioFile(path, tp)
		if err != nil {
			issue := types.NewParseError(types.MissingStorageError, 0, 0,
				fmt.Sprintf("referenced storage %q not found", rel))
			issue.WithContext(fmt.Sprintf("entrypoint or call chain"))
			scan.Issues.AddIssue(*issue)
			continue
		}
		scan.Scenarios[rel] = scenario
		if parseResult != nil {
			scan.Issues.Merge(parseResult)
		}

		refs := ExtractRefs(rel, scenario)
		allRefs = append(allRefs, refs...)
		for _, r := range refs {
			if r.Storage != "" && !visited[r.Storage] {
				queue = append(queue, r.Storage)
			}
		}
	}

	// Resolve label refs against what we visited (storage refs to non-visited
	// files are reported by LoadScenarioFile failure above).
	for _, r := range allRefs {
		resolveRefVisited(r, scan.Scenarios, scan.Issues)
	}

	scan.Refs = make([]types.ScenarioRefRecord, 0, len(allRefs))
	for _, r := range allRefs {
		scan.Refs = append(scan.Refs, types.ScenarioRefRecord{
			From: r.From, Line: r.Line, Tag: r.Tag,
			Storage: r.Storage, Target: r.Target,
		})
	}
	return scan, nil
}

// resolveRefVisited is like resolveRef but only checks labels — storage
// resolution is handled by the BFS itself.
func resolveRefVisited(r ScenarioRef, scenarios map[string]*types.ParsedScenario, issues *types.ParseResult) {
	if r.Target == "" {
		return
	}
	host := r.From
	if r.Storage != "" {
		host = r.Storage
	}
	scenario, ok := scenarios[host]
	if !ok {
		return // already reported as MissingStorageError
	}
	labelName := r.Target
	if len(labelName) > 0 && labelName[0] == '*' {
		labelName = labelName[1:]
	}
	if _, ok := scenario.Labels[labelName]; !ok {
		issue := types.NewParseError(types.MissingLabelError, r.Line, 1,
			fmt.Sprintf("referenced label %q not found", r.Target))
		issue.WithContext(fmt.Sprintf("from=%s tag=%s storage=%s", r.From, r.Tag, r.Storage))
		issues.AddIssue(*issue)
	}
}

// relativizeScenario takes an entrypoint as given on the command line and
// returns a path relative to layout.ScenarioDir.
func relativizeScenario(layout *ProjectLayout, entrypoint string) (string, error) {
	if filepath.IsAbs(entrypoint) {
		rel, err := filepath.Rel(layout.ScenarioDir, entrypoint)
		if err != nil {
			return "", fmt.Errorf("loader: relativize %s: %w", entrypoint, err)
		}
		return rel, nil
	}
	abs, err := filepath.Abs(entrypoint)
	if err != nil {
		return "", fmt.Errorf("loader: abs %s: %w", entrypoint, err)
	}
	rel, err := filepath.Rel(layout.ScenarioDir, abs)
	if err == nil && !startsWithDotDot(rel) {
		return rel, nil
	}
	// Treat as bare filename inside ScenarioDir.
	return entrypoint, nil
}

func startsWithDotDot(p string) bool {
	return len(p) >= 2 && p[0] == '.' && p[1] == '.'
}
```

- [ ] **Step 4: Run — must pass**

Run: `go test ./loader/ -run TestResolveFrom -v`
Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add loader/scan.go loader/scan_test.go
git commit -m "$(cat <<'EOF'
feat(loader): add ResolveFrom for BFS-from-entrypoint scans

Parses only scenarios reachable from the entrypoint via storage=
references. Unreachable .ks files are not parsed; missing referenced
files become MissingStorageError. Label resolution runs after the
BFS against the visited set.

Co-Authored-By: botamochi0x12 <botamochi0x12@gmail.com>
EOF
)"
```

---

### Task 3.6: Integration test against real `StarGazers/`

**Files:**
- Create: `loader/integration_test.go`

- [ ] **Step 1: Write the integration test**

Create `loader/integration_test.go`:

```go
package loader

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/botamochi0x12/tyrano-parser-go/parser"
)

func TestScanProject_RealStarGazers(t *testing.T) {
	wd, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	repoRoot := filepath.Dir(wd)
	starGazers := filepath.Join(repoRoot, "StarGazers")
	if _, err := os.Stat(starGazers); err != nil {
		t.Skipf("StarGazers fixture missing: %v", err)
	}

	layout, err := LayoutFrom(starGazers)
	if err != nil {
		t.Fatalf("LayoutFrom: %v", err)
	}
	tp := parser.NewDefaultTyranoParser()
	scan, err := ScanProject(layout, tp)
	if err != nil {
		t.Fatalf("ScanProject: %v", err)
	}

	if scan.Config["System.title"] != "StarGazers" {
		t.Errorf("config[System.title] = %q, want %q",
			scan.Config["System.title"], "StarGazers")
	}
	// StarGazers ships 8 .ks files in data/scenario.
	if len(scan.Scenarios) < 1 {
		t.Errorf("expected scenarios to be discovered, got %d", len(scan.Scenarios))
	}
	// Print a summary on failure to help diagnose.
	if t.Failed() {
		t.Logf("scanned %d scenarios, %d refs, %d issues",
			len(scan.Scenarios), len(scan.Refs), scan.Issues.GetIssueCount())
	}
}
```

- [ ] **Step 2: Run**

Run: `go test ./loader/ -run TestScanProject_RealStarGazers -v`
Expected: PASS. If it skips, ensure `StarGazers/` is a sibling of `loader/` (it is — confirmed in repo).

- [ ] **Step 3: Run full suite with race detector**

Run: `go test -race ./...`
Expected: green.

- [ ] **Step 4: Commit**

```bash
git add loader/integration_test.go
git commit -m "$(cat <<'EOF'
test(loader): integration test against real StarGazers project

End-to-end check that ScanProject works against an actual TyranoScript
project: discovers config, parses every .ks file, and reads
config[System.title] = "StarGazers". Skips gracefully when the
fixture is absent.

Co-Authored-By: botamochi0x12 <botamochi0x12@gmail.com>
EOF
)"
```

**Phase 3 done.** Programmatic project scans work.

---

## Phase 4: CLI + JSON Renderer

**Why:** Wraps Phase 1–3 in a `tyrano-parser` binary with the three commands. Default output is JSON, designed for piping into `jq`.

**Acceptance:** `go run ./cmd/tyrano-parser scenario data/scenario/first.ks`, `... config`, `... scan` all produce JSON; exit codes match the spec; golden tests pass.

---

### Task 4.1: Scaffold `cmd/tyrano-parser` and the JSON envelope types

**Files:**
- Create: `cmd/tyrano-parser/main.go`
- Create: `cmd/tyrano-parser/render_json.go`

- [ ] **Step 1: Create `cmd/tyrano-parser/main.go`**

```go
// Command tyrano-parser parses TyranoScript scenario and config files.
package main

import (
	"fmt"
	"os"
)

func main() {
	if len(os.Args) < 2 {
		usage(os.Stderr)
		os.Exit(3)
	}
	cmd := os.Args[1]
	args := os.Args[2:]
	switch cmd {
	case "scenario":
		os.Exit(runScenario(args))
	case "config":
		os.Exit(runConfig(args))
	case "scan":
		os.Exit(runScan(args))
	case "-h", "--help", "help":
		usage(os.Stdout)
		os.Exit(0)
	default:
		fmt.Fprintf(os.Stderr, "tyrano-parser: unknown command %q\n", cmd)
		usage(os.Stderr)
		os.Exit(3)
	}
}

func usage(w *os.File) {
	fmt.Fprint(w, `tyrano-parser <command> [flags] [args]

Commands:
  scenario <file.ks>           Parse one scenario file.
  config   [<Config.tjs>]      Parse one config file (default: auto-discover).
  scan     [<entrypoint.ks>]   Scan project (default: walk all scenarios).

Global flags:
  --project-root <dir>   Skip auto-discovery, use this as project root.
  --format json|report   Output format (default: json).
  --strict               Strict mode: any parse issue → exit non-zero.
  --quiet                Suppress warnings (errors still emitted).
`)
}
```

- [ ] **Step 2: Create stub renderer**

Create `cmd/tyrano-parser/render_json.go`:

```go
package main

import (
	"encoding/json"
	"fmt"
	"io"
)

// renderJSON marshals v as pretty-printed JSON to w.
func renderJSON(w io.Writer, v any) error {
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	if err := enc.Encode(v); err != nil {
		return fmt.Errorf("renderJSON: %w", err)
	}
	return nil
}
```

- [ ] **Step 3: Add stub command handlers so it compiles**

Append to `cmd/tyrano-parser/main.go`:

```go
func runScenario(args []string) int { fmt.Fprintln(os.Stderr, "scenario: not implemented yet"); return 3 }
func runConfig(args []string) int   { fmt.Fprintln(os.Stderr, "config: not implemented yet"); return 3 }
func runScan(args []string) int     { fmt.Fprintln(os.Stderr, "scan: not implemented yet"); return 3 }
```

- [ ] **Step 4: Build**

Run: `go build ./cmd/tyrano-parser`
Expected: no output (success). A `tyrano-parser` binary should appear in the repo root — leave it untracked.

- [ ] **Step 5: Verify usage works**

Run: `go run ./cmd/tyrano-parser --help`
Expected: usage text printed to stdout, exit code 0.

Run: `go run ./cmd/tyrano-parser bogus`
Expected: error message + usage on stderr, exit code 3.

- [ ] **Step 6: Commit**

```bash
git add cmd/tyrano-parser/main.go cmd/tyrano-parser/render_json.go
git commit -m "$(cat <<'EOF'
feat(cmd): scaffold tyrano-parser CLI entry point

Subcommand dispatch (scenario / config / scan), usage text, exit
code 3 for unknown commands. Subcommands are stubs that return
"not implemented yet" — implementation lands in the next tasks.

Co-Authored-By: botamochi0x12 <botamochi0x12@gmail.com>
EOF
)"
```

---

### Task 4.2: Implement shared flag parsing

**Files:**
- Create: `cmd/tyrano-parser/flags.go`
- Create: `cmd/tyrano-parser/flags_test.go`

- [ ] **Step 1: Write failing test**

Create `cmd/tyrano-parser/flags_test.go`:

```go
package main

import (
	"testing"
)

func TestParseGlobalFlags_Defaults(t *testing.T) {
	g, rest, err := parseGlobalFlags([]string{"file.ks"})
	if err != nil {
		t.Fatal(err)
	}
	if g.format != "json" {
		t.Errorf("default format = %q, want json", g.format)
	}
	if g.strict {
		t.Error("default strict should be false")
	}
	if g.quiet {
		t.Error("default quiet should be false")
	}
	if g.projectRoot != "" {
		t.Error("default projectRoot should be empty")
	}
	if len(rest) != 1 || rest[0] != "file.ks" {
		t.Errorf("rest = %v", rest)
	}
}

func TestParseGlobalFlags_AllSet(t *testing.T) {
	g, rest, err := parseGlobalFlags([]string{
		"--project-root", "/tmp/proj",
		"--format", "report",
		"--strict", "--quiet",
		"file.ks",
	})
	if err != nil {
		t.Fatal(err)
	}
	if g.projectRoot != "/tmp/proj" {
		t.Errorf("projectRoot = %q", g.projectRoot)
	}
	if g.format != "report" {
		t.Errorf("format = %q", g.format)
	}
	if !g.strict || !g.quiet {
		t.Errorf("strict/quiet not set: %+v", g)
	}
	if len(rest) != 1 || rest[0] != "file.ks" {
		t.Errorf("rest = %v", rest)
	}
}

func TestParseGlobalFlags_BadFormat(t *testing.T) {
	_, _, err := parseGlobalFlags([]string{"--format", "xml"})
	if err == nil {
		t.Error("expected error for invalid format")
	}
}
```

- [ ] **Step 2: Run — must fail**

Run: `go test ./cmd/tyrano-parser/ -v -run TestParseGlobalFlags`
Expected: `undefined: parseGlobalFlags`.

- [ ] **Step 3: Implement**

Create `cmd/tyrano-parser/flags.go`:

```go
package main

import (
	"flag"
	"fmt"
)

type globalFlags struct {
	projectRoot string
	format      string
	strict      bool
	quiet       bool
}

// parseGlobalFlags consumes the global flags from args and returns the
// remaining positional arguments.
func parseGlobalFlags(args []string) (globalFlags, []string, error) {
	fs := flag.NewFlagSet("global", flag.ContinueOnError)
	g := globalFlags{format: "json"}
	fs.StringVar(&g.projectRoot, "project-root", "", "explicit project root")
	fs.StringVar(&g.format, "format", "json", "output format: json or report")
	fs.BoolVar(&g.strict, "strict", false, "strict: any parse issue exits non-zero")
	fs.BoolVar(&g.quiet, "quiet", false, "suppress warnings")
	if err := fs.Parse(args); err != nil {
		return g, nil, err
	}
	if g.format != "json" && g.format != "report" {
		return g, nil, fmt.Errorf("--format must be 'json' or 'report', got %q", g.format)
	}
	return g, fs.Args(), nil
}
```

- [ ] **Step 4: Run — must pass**

Run: `go test ./cmd/tyrano-parser/ -v -run TestParseGlobalFlags`
Expected: all subtests PASS.

- [ ] **Step 5: Commit**

```bash
git add cmd/tyrano-parser/flags.go cmd/tyrano-parser/flags_test.go
git commit -m "$(cat <<'EOF'
feat(cmd): add shared global flag parsing

--project-root, --format, --strict, --quiet. Format validation
rejects values other than json/report.

Co-Authored-By: botamochi0x12 <botamochi0x12@gmail.com>
EOF
)"
```

---

### Task 4.3: Implement the `scenario` command

**Files:**
- Create: `cmd/tyrano-parser/command_scenario.go`
- Modify: `cmd/tyrano-parser/main.go`

- [ ] **Step 1: Implement `runScenario`**

Create `cmd/tyrano-parser/command_scenario.go`:

```go
package main

import (
	"fmt"
	"os"

	"github.com/botamochi0x12/tyrano-parser-go/loader"
	"github.com/botamochi0x12/tyrano-parser-go/parser"
	"github.com/botamochi0x12/tyrano-parser-go/types"
)

type scenarioOutput struct {
	Kind     string                 `json:"kind"`
	Path     string                 `json:"path"`
	Scenario *types.ParsedScenario  `json:"scenario"`
	Issues   []types.ParseIssue     `json:"issues"`
}

func runScenario(args []string) int {
	g, rest, err := parseGlobalFlags(args)
	if err != nil {
		fmt.Fprintln(os.Stderr, "scenario:", err)
		return 3
	}
	if len(rest) == 0 {
		fmt.Fprintln(os.Stderr, "scenario: file path required")
		return 3
	}
	path := rest[0]

	tp := parser.NewDefaultTyranoParser()
	scenario, result, err := loader.LoadScenarioFile(path, tp)
	if err != nil {
		fmt.Fprintln(os.Stderr, "scenario:", err)
		return 2
	}

	out := scenarioOutput{
		Kind:     "scenario",
		Path:     path,
		Scenario: scenario,
		Issues:   issuesOrEmpty(result, g.quiet),
	}
	if err := renderJSON(os.Stdout, out); err != nil {
		fmt.Fprintln(os.Stderr, "scenario:", err)
		return 1
	}
	return exitCodeForResult(result, g.strict)
}
```

- [ ] **Step 2: Add `issuesOrEmpty` and `exitCodeForResult` helpers**

Append to `cmd/tyrano-parser/render_json.go`:

```go
import "github.com/botamochi0x12/tyrano-parser-go/types"

// issuesOrEmpty returns result's issues, filtered when quiet is true.
// Returns a non-nil empty slice so JSON consistently renders "issues": [].
func issuesOrEmpty(result *types.ParseResult, quiet bool) []types.ParseIssue {
	if result == nil {
		return []types.ParseIssue{}
	}
	if !quiet {
		out := make([]types.ParseIssue, len(result.Issues))
		copy(out, result.Issues)
		return out
	}
	out := make([]types.ParseIssue, 0)
	for _, i := range result.Issues {
		if i.IsError() {
			out = append(out, i)
		}
	}
	return out
}

// exitCodeForResult applies strict mode and standard severity rules.
//   0 — no errors; or no issues at all
//   1 — errors present (or any issue with strict)
func exitCodeForResult(result *types.ParseResult, strict bool) int {
	if result == nil {
		return 0
	}
	if strict && result.HasIssues() {
		return 1
	}
	if result.HasErrors() {
		return 1
	}
	return 0
}
```

- [ ] **Step 3: Build and smoke-test**

Run: `go build ./cmd/tyrano-parser`
Expected: success.

Run: `go run ./cmd/tyrano-parser scenario StarGazers/data/scenario/first.ks`
Expected: JSON document on stdout with `"kind": "scenario"`, exit 0.

Run: `echo $?` (or `$status` in fish)
Expected: 0.

- [ ] **Step 4: Commit**

```bash
git add cmd/tyrano-parser/command_scenario.go cmd/tyrano-parser/render_json.go
git commit -m "$(cat <<'EOF'
feat(cmd): implement scenario subcommand

Parses one .ks file, emits a JSON envelope with the parsed
scenario and issue list. --strict and --quiet are honored.
Exit codes: 0 success, 1 parse error, 2 I/O error.

Co-Authored-By: botamochi0x12 <botamochi0x12@gmail.com>
EOF
)"
```

---

### Task 4.4: Implement the `config` command

**Files:**
- Create: `cmd/tyrano-parser/command_config.go`

- [ ] **Step 1: Implement**

Create `cmd/tyrano-parser/command_config.go`:

```go
package main

import (
	"fmt"
	"os"

	"github.com/botamochi0x12/tyrano-parser-go/loader"
	"github.com/botamochi0x12/tyrano-parser-go/parser"
	"github.com/botamochi0x12/tyrano-parser-go/types"
)

type configOutput struct {
	Kind   string             `json:"kind"`
	Path   string             `json:"path"`
	Config types.ConfigMap    `json:"config"`
	Issues []types.ParseIssue `json:"issues"`
}

func runConfig(args []string) int {
	g, rest, err := parseGlobalFlags(args)
	if err != nil {
		fmt.Fprintln(os.Stderr, "config:", err)
		return 3
	}

	cp := parser.NewConfigParser(false)
	var path string
	var config types.ConfigMap
	var result *types.ParseResult

	if len(rest) > 0 {
		path = rest[0]
		content, readErr := os.ReadFile(path)
		if readErr != nil {
			fmt.Fprintln(os.Stderr, "config:", readErr)
			return 2
		}
		config, result = cp.ParseWithResult(string(content))
	} else {
		layout, err := resolveLayout(g)
		if err != nil {
			fmt.Fprintln(os.Stderr, "config:", err)
			return 2
		}
		path = layout.ConfigPath
		var loadErr error
		config, result, loadErr = loader.LoadConfigFile(layout, cp)
		if loadErr != nil {
			fmt.Fprintln(os.Stderr, "config:", loadErr)
			return 2
		}
	}

	out := configOutput{
		Kind:   "config",
		Path:   path,
		Config: config,
		Issues: issuesOrEmpty(result, g.quiet),
	}
	if err := renderJSON(os.Stdout, out); err != nil {
		fmt.Fprintln(os.Stderr, "config:", err)
		return 1
	}
	return exitCodeForResult(result, g.strict)
}
```

- [ ] **Step 2: Add `resolveLayout` helper**

Create `cmd/tyrano-parser/layout.go`:

```go
package main

import (
	"os"

	"github.com/botamochi0x12/tyrano-parser-go/loader"
)

// resolveLayout uses --project-root if set, otherwise auto-discovers from CWD.
func resolveLayout(g globalFlags) (*loader.ProjectLayout, error) {
	if g.projectRoot != "" {
		return loader.LayoutFrom(g.projectRoot)
	}
	cwd, err := os.Getwd()
	if err != nil {
		return nil, err
	}
	return loader.DiscoverRoot(cwd)
}
```

- [ ] **Step 3: Build and smoke-test**

Run: `cd StarGazers && go run ../cmd/tyrano-parser config && cd ..`
Expected: JSON on stdout with `System.title=StarGazers` among the keys, exit 0.

- [ ] **Step 4: Commit**

```bash
git add cmd/tyrano-parser/command_config.go cmd/tyrano-parser/layout.go
git commit -m "$(cat <<'EOF'
feat(cmd): implement config subcommand

With explicit path: reads and parses that file.
Without args: auto-discovers project root, parses
data/system/Config.tjs. Honors --project-root, --strict, --quiet.

Co-Authored-By: botamochi0x12 <botamochi0x12@gmail.com>
EOF
)"
```

---

### Task 4.5: Implement the `scan` command

**Files:**
- Create: `cmd/tyrano-parser/command_scan.go`

- [ ] **Step 1: Implement**

Create `cmd/tyrano-parser/command_scan.go`:

```go
package main

import (
	"fmt"
	"os"

	"github.com/botamochi0x12/tyrano-parser-go/loader"
	"github.com/botamochi0x12/tyrano-parser-go/parser"
	"github.com/botamochi0x12/tyrano-parser-go/types"
)

type scanOutput struct {
	Kind      string             `json:"kind"`
	Root      string             `json:"root"`
	Config    types.ConfigMap    `json:"config"`
	Scenarios map[string]*types.ParsedScenario `json:"scenarios"`
	Refs      []types.ScenarioRefRecord        `json:"refs"`
	Issues    []types.ParseIssue               `json:"issues"`
}

func runScan(args []string) int {
	g, rest, err := parseGlobalFlags(args)
	if err != nil {
		fmt.Fprintln(os.Stderr, "scan:", err)
		return 3
	}

	layout, err := resolveLayout(g)
	if err != nil {
		fmt.Fprintln(os.Stderr, "scan:", err)
		return 2
	}
	tp := parser.NewDefaultTyranoParser()

	var scan *types.ProjectScan
	if len(rest) > 0 {
		scan, err = loader.ResolveFrom(layout, tp, rest[0])
	} else {
		scan, err = loader.ScanProject(layout, tp)
	}
	if err != nil {
		fmt.Fprintln(os.Stderr, "scan:", err)
		return 2
	}

	out := scanOutput{
		Kind:      "project_scan",
		Root:      scan.Root,
		Config:    scan.Config,
		Scenarios: scan.Scenarios,
		Refs:      scan.Refs,
		Issues:    issuesOrEmpty(scan.Issues, g.quiet),
	}
	if err := renderJSON(os.Stdout, out); err != nil {
		fmt.Fprintln(os.Stderr, "scan:", err)
		return 1
	}
	return exitCodeForResult(scan.Issues, g.strict)
}
```

- [ ] **Step 2: Build and smoke-test**

Run: `cd StarGazers && go run ../cmd/tyrano-parser scan > /tmp/scan.json && cd ..`
Expected: exit 0, `/tmp/scan.json` contains a `"kind": "project_scan"` document.

Run: `cat /tmp/scan.json | head -30`
Expected: pretty-printed JSON with root, config, scenarios, refs, issues.

- [ ] **Step 3: Commit**

```bash
git add cmd/tyrano-parser/command_scan.go
git commit -m "$(cat <<'EOF'
feat(cmd): implement scan subcommand

Without args: full walk via loader.ScanProject.
With entrypoint: BFS via loader.ResolveFrom.
Emits a project_scan JSON envelope. Honors --project-root,
--strict, --quiet.

Co-Authored-By: botamochi0x12 <botamochi0x12@gmail.com>
EOF
)"
```

---

### Task 4.6: Add golden tests for JSON output

**Files:**
- Create: `cmd/tyrano-parser/cli_test.go`
- Create: `cmd/tyrano-parser/testdata/projects/minimal/data/scenario/first.ks`
- Create: `cmd/tyrano-parser/testdata/projects/minimal/data/system/Config.tjs`

- [ ] **Step 1: Create the minimal fixture project**

```bash
mkdir -p cmd/tyrano-parser/testdata/projects/minimal/data/scenario
mkdir -p cmd/tyrano-parser/testdata/projects/minimal/data/system
```

Create `cmd/tyrano-parser/testdata/projects/minimal/data/scenario/first.ks`:

```
*start
[cm]
Hello[p]
[s]
```

Create `cmd/tyrano-parser/testdata/projects/minimal/data/system/Config.tjs`:

```
;System.title = "Minimal";
;scWidth = 800;
;scHeight = 600;
```

- [ ] **Step 2: Write golden test**

Create `cmd/tyrano-parser/cli_test.go`:

```go
package main

import (
	"bytes"
	"encoding/json"
	"os/exec"
	"path/filepath"
	"runtime"
	"testing"
)

// runCLI compiles the binary fresh and invokes it.
func runCLI(t *testing.T, args ...string) (stdout, stderr []byte, exitCode int) {
	t.Helper()
	bin := filepath.Join(t.TempDir(), "tyrano-parser")
	if runtime.GOOS == "windows" {
		bin += ".exe"
	}
	build := exec.Command("go", "build", "-o", bin, ".")
	build.Stderr = &bytes.Buffer{}
	if err := build.Run(); err != nil {
		t.Fatalf("build failed: %v: %s", err, build.Stderr)
	}
	cmd := exec.Command(bin, args...)
	var out, errOut bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &errOut
	err := cmd.Run()
	code := 0
	if exitErr, ok := err.(*exec.ExitError); ok {
		code = exitErr.ExitCode()
	} else if err != nil {
		t.Fatalf("cmd run failed: %v", err)
	}
	return out.Bytes(), errOut.Bytes(), code
}

func TestCLI_Config_MinimalProject(t *testing.T) {
	root := filepath.Join("testdata", "projects", "minimal")
	stdout, stderr, code := runCLI(t, "config", "--project-root", root)
	if code != 0 {
		t.Fatalf("exit = %d, stderr = %s", code, stderr)
	}
	var got map[string]any
	if err := json.Unmarshal(stdout, &got); err != nil {
		t.Fatalf("invalid JSON: %v\noutput: %s", err, stdout)
	}
	if got["kind"] != "config" {
		t.Errorf("kind = %v, want config", got["kind"])
	}
	cfg, _ := got["config"].(map[string]any)
	if cfg["System.title"] != "Minimal" {
		t.Errorf("System.title = %v, want Minimal", cfg["System.title"])
	}
}

func TestCLI_Scenario_MinimalProject(t *testing.T) {
	path := filepath.Join("testdata", "projects", "minimal",
		"data", "scenario", "first.ks")
	stdout, stderr, code := runCLI(t, "scenario", path)
	if code != 0 {
		t.Fatalf("exit = %d, stderr = %s", code, stderr)
	}
	var got map[string]any
	if err := json.Unmarshal(stdout, &got); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	if got["kind"] != "scenario" {
		t.Errorf("kind = %v, want scenario", got["kind"])
	}
}

func TestCLI_Scan_MinimalProject(t *testing.T) {
	root := filepath.Join("testdata", "projects", "minimal")
	stdout, stderr, code := runCLI(t, "scan", "--project-root", root)
	if code != 0 {
		t.Fatalf("exit = %d, stderr = %s", code, stderr)
	}
	var got map[string]any
	if err := json.Unmarshal(stdout, &got); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	if got["kind"] != "project_scan" {
		t.Errorf("kind = %v, want project_scan", got["kind"])
	}
}

func TestCLI_UnknownCommand_Exit3(t *testing.T) {
	_, _, code := runCLI(t, "bogus")
	if code != 3 {
		t.Errorf("exit = %d, want 3", code)
	}
}

func TestCLI_MissingFile_Exit2(t *testing.T) {
	_, _, code := runCLI(t, "scenario", "/no/such/file.ks")
	if code != 2 {
		t.Errorf("exit = %d, want 2", code)
	}
}
```

- [ ] **Step 3: Run**

Run: `go test ./cmd/tyrano-parser/ -v`
Expected: all PASS.

- [ ] **Step 4: Commit**

```bash
git add cmd/tyrano-parser/cli_test.go cmd/tyrano-parser/testdata
git commit -m "$(cat <<'EOF'
test(cmd): golden tests for CLI commands

Build the binary fresh per test, run against a minimal fixture
project, assert kind/key fields and exit codes for all three
commands plus the unknown-command and missing-file paths.

Co-Authored-By: botamochi0x12 <botamochi0x12@gmail.com>
EOF
)"
```

---

### Task 4.7: Move root `main.go` into `cmd/tyrano-parser-example/`

**Files:**
- Create: `cmd/tyrano-parser-example/main.go`
- Delete: `main.go` (root)

- [ ] **Step 1: Move the example**

Run: `mkdir -p cmd/tyrano-parser-example && mv main.go cmd/tyrano-parser-example/main.go`

- [ ] **Step 2: Update the package comment**

Edit `cmd/tyrano-parser-example/main.go` — at the top, replace `package main` with:

```go
// Command tyrano-parser-example demonstrates programmatic use of the
// parser package. The actual CLI is cmd/tyrano-parser.
package main
```

- [ ] **Step 3: Verify both binaries build**

Run: `go build ./cmd/tyrano-parser ./cmd/tyrano-parser-example`
Expected: success, two binaries appear in repo root. Remove them after if you don't want them tracked: `rm -f tyrano-parser tyrano-parser-example`.

- [ ] **Step 4: Verify example still runs**

Run: `go run ./cmd/tyrano-parser-example`
Expected: same output as before — Parsed scenario, Parsed config, Example error.

- [ ] **Step 5: Commit**

```bash
git add -A
git commit -m "$(cat <<'EOF'
refactor: move example main into cmd/tyrano-parser-example

The root main.go was an embedded example, but the project now has
a real CLI at cmd/tyrano-parser. Move the example to its own
subdir under cmd/ so it does not collide with the CLI binary.

Co-Authored-By: botamochi0x12 <botamochi0x12@gmail.com>
EOF
)"
```

**Phase 4 done.** `tyrano-parser` works as a CLI with JSON output.

---

## Phase 5: Report Renderer

**Why:** Adds the human-readable `--format report` output so the CLI is usable from a terminal without piping through `jq`.

**Acceptance:** All three commands accept `--format report`. Output matches the design doc's section order. Golden tests pass.

---

### Task 5.1: Implement `renderReport` skeleton

**Files:**
- Create: `cmd/tyrano-parser/render_report.go`
- Create: `cmd/tyrano-parser/render_report_test.go`

- [ ] **Step 1: Write failing test for `renderConfigReport`**

Create `cmd/tyrano-parser/render_report_test.go`:

```go
package main

import (
	"bytes"
	"strings"
	"testing"

	"github.com/botamochi0x12/tyrano-parser-go/types"
)

func TestRenderConfigReport_HappyPath(t *testing.T) {
	out := configOutput{
		Kind:   "config",
		Path:   "/p/data/system/Config.tjs",
		Config: types.ConfigMap{"System.title": "Foo", "scWidth": "800"},
		Issues: nil,
	}
	var buf bytes.Buffer
	if err := renderConfigReport(&buf, out); err != nil {
		t.Fatal(err)
	}
	got := buf.String()
	for _, want := range []string{
		"Config: /p/data/system/Config.tjs",
		"keys: 2",
		"System.title",
		"scWidth",
	} {
		if !strings.Contains(got, want) {
			t.Errorf("missing %q in:\n%s", want, got)
		}
	}
}

func TestRenderConfigReport_WithErrors(t *testing.T) {
	issue := types.NewParseError(types.InvalidConfigError, 5, 1, "bad line")
	out := configOutput{
		Kind:   "config",
		Path:   "x",
		Config: types.ConfigMap{},
		Issues: []types.ParseIssue{*issue},
	}
	var buf bytes.Buffer
	if err := renderConfigReport(&buf, out); err != nil {
		t.Fatal(err)
	}
	got := buf.String()
	if !strings.Contains(got, "[error]") {
		t.Errorf("expected [error] marker in:\n%s", got)
	}
	if !strings.Contains(got, "bad line") {
		t.Errorf("expected error message in:\n%s", got)
	}
}
```

- [ ] **Step 2: Run — must fail**

Run: `go test ./cmd/tyrano-parser/ -run TestRenderConfigReport -v`
Expected: `undefined: renderConfigReport`.

- [ ] **Step 3: Implement skeleton + config renderer**

Create `cmd/tyrano-parser/render_report.go`:

```go
package main

import (
	"fmt"
	"io"
	"sort"

	"github.com/botamochi0x12/tyrano-parser-go/types"
)

func renderConfigReport(w io.Writer, out configOutput) error {
	fmt.Fprintf(w, "Config: %s\n", out.Path)
	fmt.Fprintf(w, "keys: %d\n\n", len(out.Config))

	keys := make([]string, 0, len(out.Config))
	for k := range out.Config {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		fmt.Fprintf(w, "  %s = %s\n", k, out.Config[k])
	}
	if len(out.Issues) > 0 {
		fmt.Fprintln(w, "\nIssues:")
		for _, i := range out.Issues {
			fmt.Fprintf(w, "  [%s] line %d:%d %s\n",
				i.Severity.String(), i.Line, i.Column, i.Message)
		}
	}
	return nil
}

func writeIssue(w io.Writer, i types.ParseIssue) {
	fmt.Fprintf(w, "  [%s] line %d:%d %s\n",
		i.Severity.String(), i.Line, i.Column, i.Message)
}
```

- [ ] **Step 4: Run — must pass**

Run: `go test ./cmd/tyrano-parser/ -run TestRenderConfigReport -v`
Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add cmd/tyrano-parser/render_report.go cmd/tyrano-parser/render_report_test.go
git commit -m "$(cat <<'EOF'
feat(cmd): add report renderer for config output

Sorted key/value listing with optional issues block.
Foundation for the scenario and scan report renderers.

Co-Authored-By: botamochi0x12 <botamochi0x12@gmail.com>
EOF
)"
```

---

### Task 5.2: Add scenario and scan report renderers

**Files:**
- Modify: `cmd/tyrano-parser/render_report.go`
- Modify: `cmd/tyrano-parser/render_report_test.go`

- [ ] **Step 1: Write failing tests**

Append to `cmd/tyrano-parser/render_report_test.go`:

```go
func TestRenderScenarioReport_HappyPath(t *testing.T) {
	scenario := types.NewParsedScenario()
	scenario.Labels["start"] = types.NewLabelInfo("start", 1, 0, "")
	scenario.Elements = []types.ParsedTag{{Name: "cm", Line: 2}}
	out := scenarioOutput{
		Kind:     "scenario",
		Path:     "first.ks",
		Scenario: scenario,
		Issues:   nil,
	}
	var buf bytes.Buffer
	if err := renderScenarioReport(&buf, out); err != nil {
		t.Fatal(err)
	}
	got := buf.String()
	for _, want := range []string{"Scenario: first.ks", "elements: 1", "labels: 1"} {
		if !strings.Contains(got, want) {
			t.Errorf("missing %q in:\n%s", want, got)
		}
	}
}

func TestRenderScanReport_HappyPath(t *testing.T) {
	out := scanOutput{
		Kind: "project_scan",
		Root: "/p",
		Config: types.ConfigMap{"System.title": "X"},
		Scenarios: map[string]*types.ParsedScenario{
			"first.ks": types.NewParsedScenario(),
		},
		Refs: []types.ScenarioRefRecord{
			{From: "first.ks", Line: 2, Tag: "call", Storage: "tyrano.ks"},
		},
		Issues: nil,
	}
	var buf bytes.Buffer
	if err := renderScanReport(&buf, out); err != nil {
		t.Fatal(err)
	}
	got := buf.String()
	for _, want := range []string{"Project: /p", "Scenarios (1)", "Cross-references (1)", "first.ks:2"} {
		if !strings.Contains(got, want) {
			t.Errorf("missing %q in:\n%s", want, got)
		}
	}
}
```

- [ ] **Step 2: Run — must fail**

Run: `go test ./cmd/tyrano-parser/ -run TestRender -v`
Expected: `undefined: renderScenarioReport` / `renderScanReport`.

- [ ] **Step 3: Implement**

Append to `cmd/tyrano-parser/render_report.go`:

```go
func renderScenarioReport(w io.Writer, out scenarioOutput) error {
	fmt.Fprintf(w, "Scenario: %s\n", out.Path)
	if out.Scenario != nil {
		fmt.Fprintf(w, "elements: %d\n", len(out.Scenario.Elements))
		fmt.Fprintf(w, "labels: %d\n", len(out.Scenario.Labels))
	}
	if len(out.Issues) > 0 {
		fmt.Fprintln(w, "\nIssues:")
		for _, i := range out.Issues {
			writeIssue(w, i)
		}
	}
	return nil
}

func renderScanReport(w io.Writer, out scanOutput) error {
	fmt.Fprintf(w, "Project: %s\n", out.Root)
	fmt.Fprintf(w, "Config keys: %d\n\n", len(out.Config))

	fmt.Fprintf(w, "Scenarios (%d):\n", len(out.Scenarios))
	scnNames := make([]string, 0, len(out.Scenarios))
	for k := range out.Scenarios {
		scnNames = append(scnNames, k)
	}
	sort.Strings(scnNames)
	for _, name := range scnNames {
		s := out.Scenarios[name]
		fmt.Fprintf(w, "  %s — %d elements, %d labels\n",
			name, len(s.Elements), len(s.Labels))
	}

	fmt.Fprintf(w, "\nCross-references (%d):\n", len(out.Refs))
	for _, r := range out.Refs {
		switch {
		case r.Storage != "" && r.Target != "":
			fmt.Fprintf(w, "  %s:%d  @%s storage=%q target=%q\n",
				r.From, r.Line, r.Tag, r.Storage, r.Target)
		case r.Storage != "":
			fmt.Fprintf(w, "  %s:%d  @%s storage=%q\n",
				r.From, r.Line, r.Tag, r.Storage)
		default:
			fmt.Fprintf(w, "  %s:%d  @%s target=%q\n",
				r.From, r.Line, r.Tag, r.Target)
		}
	}

	if len(out.Issues) > 0 {
		fmt.Fprintln(w, "\nIssues:")
		for _, i := range out.Issues {
			writeIssue(w, i)
		}
	}

	errors, warnings := 0, 0
	for _, i := range out.Issues {
		if i.IsError() {
			errors++
		} else if i.IsWarning() {
			warnings++
		}
	}
	fmt.Fprintf(w, "\nSummary: %d error(s), %d warning(s)\n", errors, warnings)
	return nil
}
```

- [ ] **Step 4: Run — must pass**

Run: `go test ./cmd/tyrano-parser/ -run TestRender -v`
Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add cmd/tyrano-parser/render_report.go cmd/tyrano-parser/render_report_test.go
git commit -m "$(cat <<'EOF'
feat(cmd): add scenario and scan report renderers

Sorted scenario list with element/label counts, cross-reference
list keyed by source line, optional issues block, summary tally.

Co-Authored-By: botamochi0x12 <botamochi0x12@gmail.com>
EOF
)"
```

---

### Task 5.3: Wire `--format report` into all three commands

**Files:**
- Modify: `cmd/tyrano-parser/command_scenario.go`
- Modify: `cmd/tyrano-parser/command_config.go`
- Modify: `cmd/tyrano-parser/command_scan.go`

- [ ] **Step 1: Replace the renderJSON call in `runScenario`**

In `cmd/tyrano-parser/command_scenario.go`, replace:

```go
	if err := renderJSON(os.Stdout, out); err != nil {
		fmt.Fprintln(os.Stderr, "scenario:", err)
		return 1
	}
```

with:

```go
	switch g.format {
	case "report":
		if err := renderScenarioReport(os.Stdout, out); err != nil {
			fmt.Fprintln(os.Stderr, "scenario:", err)
			return 1
		}
	default:
		if err := renderJSON(os.Stdout, out); err != nil {
			fmt.Fprintln(os.Stderr, "scenario:", err)
			return 1
		}
	}
```

- [ ] **Step 2: Apply the same pattern in `runConfig` and `runScan`**

In `command_config.go`, swap the renderJSON call for:

```go
	switch g.format {
	case "report":
		if err := renderConfigReport(os.Stdout, out); err != nil {
			fmt.Fprintln(os.Stderr, "config:", err)
			return 1
		}
	default:
		if err := renderJSON(os.Stdout, out); err != nil {
			fmt.Fprintln(os.Stderr, "config:", err)
			return 1
		}
	}
```

In `command_scan.go`, swap for:

```go
	switch g.format {
	case "report":
		if err := renderScanReport(os.Stdout, out); err != nil {
			fmt.Fprintln(os.Stderr, "scan:", err)
			return 1
		}
	default:
		if err := renderJSON(os.Stdout, out); err != nil {
			fmt.Fprintln(os.Stderr, "scan:", err)
			return 1
		}
	}
```

- [ ] **Step 3: Add CLI test for report format**

Append to `cmd/tyrano-parser/cli_test.go`:

```go
func TestCLI_Config_ReportFormat(t *testing.T) {
	root := filepath.Join("testdata", "projects", "minimal")
	stdout, stderr, code := runCLI(t, "config", "--project-root", root, "--format", "report")
	if code != 0 {
		t.Fatalf("exit = %d, stderr = %s", code, stderr)
	}
	got := string(stdout)
	for _, want := range []string{"Config:", "keys:", "System.title = Minimal"} {
		if !bytes.Contains(stdout, []byte(want)) {
			t.Errorf("missing %q in report output:\n%s", want, got)
		}
	}
}

func TestCLI_Scan_ReportFormat(t *testing.T) {
	root := filepath.Join("testdata", "projects", "minimal")
	_, stderr, code := runCLI(t, "scan", "--project-root", root, "--format", "report")
	if code != 0 {
		t.Fatalf("exit = %d, stderr = %s", code, stderr)
	}
}
```

- [ ] **Step 4: Build, run smoke tests**

Run: `go build ./cmd/tyrano-parser`
Expected: success.

Run: `cd StarGazers && go run ../cmd/tyrano-parser scan --format report && cd ..`
Expected: human-readable scan report on stdout, exit 0.

- [ ] **Step 5: Run full test suite with race**

Run: `go test -race ./...`
Expected: all green.

- [ ] **Step 6: Commit**

```bash
git add cmd/tyrano-parser
git commit -m "$(cat <<'EOF'
feat(cmd): wire --format report into all three subcommands

Route output through renderScenarioReport / renderConfigReport /
renderScanReport when --format=report. Default remains JSON.
Adds two CLI golden tests for the report path.

Co-Authored-By: botamochi0x12 <botamochi0x12@gmail.com>
EOF
)"
```

**Phase 5 done.** Full CLI works in both formats.

---

## Final Verification

- [ ] **All tests pass with race detector**

Run: `go test -race ./...`
Expected: all packages green.

- [ ] **Coverage check**

Run: `go test -cover ./...`
Expected: parser, loader, cmd packages each ≥80% coverage. If any are short, add table cases for the gap before declaring done.

- [ ] **Smoke test against StarGazers**

Run from project root:
```bash
go run ./cmd/tyrano-parser config --project-root StarGazers --format report
go run ./cmd/tyrano-parser scan --project-root StarGazers --format report
go run ./cmd/tyrano-parser scenario StarGazers/data/scenario/first.ks --format report
```
Expected: each prints a sensible report, exit 0.

- [ ] **`go vet`**

Run: `go vet ./...`
Expected: no warnings.

- [ ] **No leftover binaries in repo root**

Run: `ls *.exe tyrano-parser tyrano-parser-example 2>/dev/null && echo "cleanup needed" || echo "clean"`
Expected: "clean". If any binaries are present, remove them — the build outputs were never meant to be tracked.

- [ ] **Final commit if anything was tweaked**

If the verification surfaced anything (lint warning, coverage gap, stray binary), fix and commit with a `chore:` prefix and the same Co-Authored-By trailer.
