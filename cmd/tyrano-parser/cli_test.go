package main

import (
	"bytes"
	"encoding/json"
	"os/exec"
	"path/filepath"
	"runtime"
	"testing"
)

func runCLI(t *testing.T, args ...string) (stdout, stderr []byte, exitCode int) {
	t.Helper()
	return runCLIWithLDFlags(t, "", args...)
}

func runCLIWithLDFlags(t *testing.T, ldflags string, args ...string) (stdout, stderr []byte, exitCode int) {
	t.Helper()
	bin := filepath.Join(t.TempDir(), "tyrano-parser")
	if runtime.GOOS == "windows" {
		bin += ".exe"
	}
	buildArgs := []string{"build", "-o", bin}
	if ldflags != "" {
		buildArgs = append(buildArgs, "-ldflags", ldflags)
	}
	buildArgs = append(buildArgs, ".")
	build := exec.Command("go", buildArgs...)
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
	path := filepath.Join("testdata", "projects", "minimal", "data", "scenario", "first.ks")
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

func TestCLI_Config_ReportFormat(t *testing.T) {
	root := filepath.Join("testdata", "projects", "minimal")
	stdout, stderr, code := runCLI(t, "config", "--project-root", root, "--format", "report")
	if code != 0 {
		t.Fatalf("exit = %d, stderr = %s", code, stderr)
	}
	for _, want := range []string{"Config:", "keys:", "System.title = Minimal"} {
		if !bytes.Contains(stdout, []byte(want)) {
			t.Errorf("missing %q in report output:\n%s", want, stdout)
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

func TestCLI_Version_DefaultsToDev(t *testing.T) {
	stdout, stderr, code := runCLI(t, "version")
	if code != 0 {
		t.Fatalf("exit = %d, stderr = %s", code, stderr)
	}
	if !bytes.Contains(stdout, []byte("dev")) {
		t.Errorf("stdout = %q, want the default version %q", stdout, "dev")
	}
}

func TestCLI_Version_ReportsInjectedVersion(t *testing.T) {
	stdout, stderr, code := runCLIWithLDFlags(t, "-X main.version=v9.9.9", "version")
	if code != 0 {
		t.Fatalf("exit = %d, stderr = %s", code, stderr)
	}
	if !bytes.Contains(stdout, []byte("v9.9.9")) {
		t.Errorf("stdout = %q, want the linker-injected version %q", stdout, "v9.9.9")
	}
}

func TestCLI_VersionFlag_MatchesVersionCommand(t *testing.T) {
	fromCommand, _, code := runCLI(t, "version")
	if code != 0 {
		t.Fatalf("version command exit = %d", code)
	}
	fromFlag, _, code := runCLI(t, "--version")
	if code != 0 {
		t.Fatalf("--version exit = %d", code)
	}
	if !bytes.Equal(fromCommand, fromFlag) {
		t.Errorf("--version = %q, want same output as version command %q", fromFlag, fromCommand)
	}
}
