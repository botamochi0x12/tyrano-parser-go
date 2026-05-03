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
	content := "*start\n[cm]\nHello[p]\n[s]\n"
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
	scenario, result, err := LoadScenarioFile(path, parser.NewDefaultTyranoParser())
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
	_, _, err := LoadScenarioFile("/no/such/file.ks", parser.NewDefaultTyranoParser())
	if err == nil {
		t.Fatal("expected error for missing file, got nil")
	}
}

func TestWalkScenarios_FindsAllKsFiles(t *testing.T) {
	tmp := t.TempDir()
	scnDir := filepath.Join(tmp, "data", "scenario")
	if err := os.MkdirAll(scnDir, 0o755); err != nil {
		t.Fatal(err)
	}
	files := map[string]string{
		"first.ks":  "*start\n[s]\n",
		"title.ks":  "*title\n[s]\n",
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
	scenarios, result, err := WalkScenarios(layout, parser.NewDefaultTyranoParser())
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
