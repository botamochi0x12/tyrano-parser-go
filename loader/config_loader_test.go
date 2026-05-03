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
	if err := os.WriteFile(filepath.Join(cfgDir, "Config.tjs"), []byte(`;title = "Hello";`), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(tmp, "data", "scenario"), 0o755); err != nil {
		t.Fatal(err)
	}
	layout, err := LayoutFrom(tmp)
	if err != nil {
		t.Fatal(err)
	}
	config, result, err := LoadConfigFile(layout, parser.NewConfigParser(false))
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
	if err := os.MkdirAll(filepath.Join(tmp, "data", "scenario"), 0o755); err != nil {
		t.Fatal(err)
	}
	layout, err := LayoutFrom(tmp)
	if err != nil {
		t.Fatal(err)
	}
	config, result, err := LoadConfigFile(layout, parser.NewConfigParser(false))
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
