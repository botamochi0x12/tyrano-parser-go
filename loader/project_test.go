package loader

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
	"testing/fstest"
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

func TestDiscoverRoot_FindsScenarioMarker(t *testing.T) {
	fsys := fstest.MapFS{
		"proj/data/scenario/first.ks": {Data: []byte("")},
		"proj/data/system/Config.tjs": {Data: []byte("")},
		"proj/sub/deeper/note.txt":    {Data: []byte("")},
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
	fsys := fstest.MapFS{"unrelated/file.txt": {Data: []byte("")}}
	_, err := discoverRootFS(fsys, "unrelated")
	if !errors.Is(err, ErrProjectNotFound) {
		t.Errorf("expected ErrProjectNotFound, got %v", err)
	}
}

func TestDiscoverRoot_FindsConfigOnlyMarker(t *testing.T) {
	fsys := fstest.MapFS{"proj/data/system/Config.tjs": {Data: []byte("")}}
	got, err := discoverRootFS(fsys, "proj")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.Root != "proj" {
		t.Errorf("Root = %q, want %q", got.Root, "proj")
	}
}

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

func TestDiscoverRoot_OSFilesystem(t *testing.T) {
	tmp := t.TempDir()
	nested := filepath.Join(tmp, "sub", "deeper")
	if err := os.MkdirAll(filepath.Join(tmp, "data", "scenario"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(nested, 0o755); err != nil {
		t.Fatal(err)
	}
	got, err := DiscoverRoot(nested)
	if err != nil {
		t.Fatalf("DiscoverRoot: %v", err)
	}
	if got.Root != tmp {
		t.Errorf("Root = %q, want %q", got.Root, tmp)
	}
}
