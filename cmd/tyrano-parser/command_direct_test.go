package main

import (
	"bytes"
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/botamochi0x12/tyrano-parser-go/types"
)

func captureStdout(t *testing.T, fn func() int) ([]byte, int) {
	t.Helper()
	old := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	os.Stdout = w
	code := fn()
	if err := w.Close(); err != nil {
		t.Fatal(err)
	}
	os.Stdout = old
	var buf bytes.Buffer
	if _, err := io.Copy(&buf, r); err != nil {
		t.Fatal(err)
	}
	return buf.Bytes(), code
}

func TestRunScenario_ReportDirect(t *testing.T) {
	path := filepath.Join("testdata", "projects", "minimal", "data", "scenario", "first.ks")
	stdout, code := captureStdout(t, func() int {
		return runScenario([]string{path, "--format", "report"})
	})
	if code != 0 {
		t.Fatalf("exit = %d", code)
	}
	if !bytes.Contains(stdout, []byte("Scenario:")) {
		t.Fatalf("expected report output, got:\n%s", stdout)
	}
}

func TestRunConfig_JSONDirect(t *testing.T) {
	path := filepath.Join("testdata", "projects", "minimal", "data", "system", "Config.tjs")
	stdout, code := captureStdout(t, func() int {
		return runConfig([]string{path})
	})
	if code != 0 {
		t.Fatalf("exit = %d", code)
	}
	var got configOutput
	if err := json.Unmarshal(stdout, &got); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	if got.Config["System.title"] != "Minimal" {
		t.Errorf("System.title = %q", got.Config["System.title"])
	}
}

func TestRunConfig_ProjectRootReportDirect(t *testing.T) {
	root := filepath.Join("testdata", "projects", "minimal")
	stdout, code := captureStdout(t, func() int {
		return runConfig([]string{"--project-root", root, "--format", "report"})
	})
	if code != 0 {
		t.Fatalf("exit = %d", code)
	}
	if !bytes.Contains(stdout, []byte("System.title = Minimal")) {
		t.Fatalf("expected config report, got:\n%s", stdout)
	}
}

func TestRunScan_JSONDirect(t *testing.T) {
	root := filepath.Join("testdata", "projects", "minimal")
	stdout, code := captureStdout(t, func() int {
		return runScan([]string{"--project-root", root})
	})
	if code != 0 {
		t.Fatalf("exit = %d", code)
	}
	var got scanOutput
	if err := json.Unmarshal(stdout, &got); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	if got.Kind != "project_scan" || len(got.Scenarios) != 1 {
		t.Errorf("unexpected scan output: %+v", got)
	}
}

func TestRunScan_EntrypointReportDirect(t *testing.T) {
	root := filepath.Join("testdata", "projects", "minimal")
	stdout, code := captureStdout(t, func() int {
		return runScan([]string{"--project-root", root, "first.ks", "--format", "report"})
	})
	if code != 0 {
		t.Fatalf("exit = %d", code)
	}
	if !bytes.Contains(stdout, []byte("Project:")) {
		t.Fatalf("expected scan report, got:\n%s", stdout)
	}
}

func TestCommandUsageErrorsDirect(t *testing.T) {
	if code := runScenario(nil); code != 3 {
		t.Errorf("runScenario missing arg exit = %d, want 3", code)
	}
	if code := runScenario([]string{"/no/such/file.ks"}); code != 2 {
		t.Errorf("runScenario missing file exit = %d, want 2", code)
	}
	if code := runConfig([]string{"/no/such/Config.tjs"}); code != 2 {
		t.Errorf("runConfig missing file exit = %d, want 2", code)
	}
	if code := runScan([]string{"--format", "xml"}); code != 3 {
		t.Errorf("runScan bad format exit = %d, want 3", code)
	}
}

func TestJSONHelpers(t *testing.T) {
	result := types.NewParseResult()
	result.AddWarning(types.InvalidConfigError, 1, 1, "warn")
	result.AddError(types.SyntaxError, 2, 1, "err")
	if exitCodeForResult(result, false) != 1 {
		t.Error("errors should exit 1")
	}
	if exitCodeForResult(result, true) != 1 {
		t.Error("strict issues should exit 1")
	}
	if len(issuesOrEmpty(result, true)) != 1 {
		t.Error("quiet should keep only errors")
	}
	var buf bytes.Buffer
	if err := renderJSON(&buf, map[string]string{"ok": "true"}); err != nil {
		t.Fatal(err)
	}
	if !bytes.Contains(buf.Bytes(), []byte(`"ok"`)) {
		t.Errorf("unexpected JSON: %s", buf.Bytes())
	}
}

func TestMainWithArgsUsagePaths(t *testing.T) {
	stdout, err := os.CreateTemp(t.TempDir(), "stdout")
	if err != nil {
		t.Fatal(err)
	}
	stderr, err := os.CreateTemp(t.TempDir(), "stderr")
	if err != nil {
		t.Fatal(err)
	}
	if code := mainWithArgs(nil, stdout, stderr); code != 3 {
		t.Errorf("no args exit = %d, want 3", code)
	}
	if code := mainWithArgs([]string{"--help"}, stdout, stderr); code != 0 {
		t.Errorf("help exit = %d, want 0", code)
	}
	if code := mainWithArgs([]string{"bogus"}, stdout, stderr); code != 3 {
		t.Errorf("bogus exit = %d, want 3", code)
	}
}

func TestReportRenderersWithIssuesAndRefShapes(t *testing.T) {
	issue := *types.NewParseWarning(types.InvalidConfigError, 3, 2, "warn")
	scenario := types.NewParsedScenario()
	var scenarioBuf bytes.Buffer
	if err := renderScenarioReport(&scenarioBuf, scenarioOutput{
		Path:     "x.ks",
		Scenario: scenario,
		Issues:   []types.ParseIssue{issue},
	}); err != nil {
		t.Fatal(err)
	}
	if !bytes.Contains(scenarioBuf.Bytes(), []byte("[warning]")) {
		t.Errorf("scenario report missing issue: %s", scenarioBuf.Bytes())
	}

	var scanBuf bytes.Buffer
	if err := renderScanReport(&scanBuf, scanOutput{
		Root:      "/p",
		Config:    types.ConfigMap{},
		Scenarios: map[string]*types.ParsedScenario{"x.ks": types.NewParsedScenario()},
		Refs: []types.ScenarioRefRecord{
			{From: "x.ks", Line: 1, Tag: "jump", Storage: "y.ks", Target: "*start"},
			{From: "x.ks", Line: 2, Tag: "jump", Target: "*local"},
		},
		Issues: []types.ParseIssue{issue},
	}); err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{`storage="y.ks" target="*start"`, `target="*local"`, "[warning]"} {
		if !bytes.Contains(scanBuf.Bytes(), []byte(want)) {
			t.Errorf("scan report missing %q:\n%s", want, scanBuf.Bytes())
		}
	}
}
