package loader

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/botamochi0x12/tyrano-parser-go/parser"
	"github.com/botamochi0x12/tyrano-parser-go/types"
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
	writeFile(t, filepath.Join(scn, "first.ks"), "*start\n@call storage=\"tyrano.ks\"\n@jump storage=\"title.ks\"\n[s]\n")
	writeFile(t, filepath.Join(scn, "tyrano.ks"), "*x\n[s]\n")
	writeFile(t, filepath.Join(scn, "title.ks"), "*y\n[s]\n")
	layout, err := LayoutFrom(tmp)
	if err != nil {
		t.Fatal(err)
	}
	scan, err := ScanProject(layout, parser.NewDefaultTyranoParser())
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
	writeFile(t, filepath.Join(scn, "first.ks"), "*start\n@jump storage=\"missing.ks\"\n[s]\n")
	layout, err := LayoutFrom(tmp)
	if err != nil {
		t.Fatal(err)
	}
	scan, err := ScanProject(layout, parser.NewDefaultTyranoParser())
	if err != nil {
		t.Fatal(err)
	}
	errs := scan.Issues.GetErrors()
	if len(errs) != 1 {
		t.Fatalf("expected 1 missing-storage error, got %d: %v", len(errs), errs)
	}
	if errs[0].Type.String() != "MissingStorageError" {
		t.Errorf("error type = %q, want %q", errs[0].Type.String(), "MissingStorageError")
	}
}

func TestScanProject_SkipsDynamicRefs(t *testing.T) {
	tmp := t.TempDir()
	scn := filepath.Join(tmp, "data", "scenario")
	writeFile(t, filepath.Join(scn, "first.ks"), "*start\n@jump target=&tf.target_page\n@jump storage=&tf.storage target=&tf.target\n[s]\n")
	layout, err := LayoutFrom(tmp)
	if err != nil {
		t.Fatal(err)
	}
	scan, err := ScanProject(layout, parser.NewDefaultTyranoParser())
	if err != nil {
		t.Fatal(err)
	}
	if scan.Issues.HasErrors() {
		t.Errorf("dynamic refs should not produce missing-ref errors: %v", scan.Issues.GetErrors())
	}
}

func TestResolveFrom_OnlyVisitsReachable(t *testing.T) {
	tmp := t.TempDir()
	scn := filepath.Join(tmp, "data", "scenario")
	writeFile(t, filepath.Join(scn, "first.ks"), "*start\n@call storage=\"tyrano.ks\"\n[s]\n")
	writeFile(t, filepath.Join(scn, "tyrano.ks"), "*x\n[s]\n")
	writeFile(t, filepath.Join(scn, "unreached.ks"), "*z\n[s]\n")
	layout, err := LayoutFrom(tmp)
	if err != nil {
		t.Fatal(err)
	}
	scan, err := ResolveFrom(layout, parser.NewDefaultTyranoParser(), "first.ks")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(scan.Scenarios) != 2 {
		t.Errorf("expected 2 reachable scenarios, got %d: %v", len(scan.Scenarios), keys(scan.Scenarios))
	}
	if _, ok := scan.Scenarios["unreached.ks"]; ok {
		t.Errorf("unreached.ks should not have been parsed")
	}
}

func TestResolveFrom_AbsoluteEntrypointAndMissingLabel(t *testing.T) {
	tmp := t.TempDir()
	scn := filepath.Join(tmp, "data", "scenario")
	first := filepath.Join(scn, "first.ks")
	writeFile(t, first, "*start\n@jump target=\"*missing\"\n[s]\n")
	layout, err := LayoutFrom(tmp)
	if err != nil {
		t.Fatal(err)
	}
	scan, err := ResolveFrom(layout, parser.NewDefaultTyranoParser(), first)
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := scan.Scenarios["first.ks"]; !ok {
		t.Fatalf("first.ks missing from scan: %#v", scan.Scenarios)
	}
	errs := scan.Issues.GetErrors()
	if len(errs) != 1 || errs[0].Type != types.MissingLabelError {
		t.Fatalf("expected one MissingLabelError, got %v", errs)
	}
}

func TestResolveFrom_MissingStorageInChain(t *testing.T) {
	tmp := t.TempDir()
	scn := filepath.Join(tmp, "data", "scenario")
	writeFile(t, filepath.Join(scn, "first.ks"), "*start\n@call storage=\"missing.ks\"\n[s]\n")
	layout, err := LayoutFrom(tmp)
	if err != nil {
		t.Fatal(err)
	}
	scan, err := ResolveFrom(layout, parser.NewDefaultTyranoParser(), "first.ks")
	if err != nil {
		t.Fatal(err)
	}
	errs := scan.Issues.GetErrors()
	if len(errs) != 1 || errs[0].Type != types.MissingStorageError {
		t.Fatalf("expected one MissingStorageError, got %v", errs)
	}
}
