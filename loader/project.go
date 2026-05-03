package loader

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

type ProjectLayout struct {
	Root        string
	ScenarioDir string
	ConfigPath  string
}

var (
	ErrProjectNotFound = errors.New("loader: no TyranoScript project found")
	ErrNoScenarioDir   = errors.New("loader: project missing data/scenario directory")
)

func DiscoverRoot(start string) (*ProjectLayout, error) {
	abs, err := filepath.Abs(start)
	if err != nil {
		return nil, fmt.Errorf("loader: resolve start path %q: %w", start, err)
	}
	return discoverRootFS(osFS{}, abs)
}

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

func discoverRootFS(fsys fs.FS, start string) (*ProjectLayout, error) {
	cur := filepath.Clean(start)
	for {
		if hasScenarioOrConfig(fsys, cur) {
			return layoutAt(cur), nil
		}
		parent := filepath.Dir(cur)
		if parent == cur || parent == "." {
			return nil, ErrProjectNotFound
		}
		cur = parent
	}
}

func hasScenarioOrConfig(fsys fs.FS, dir string) bool {
	return statDir(fsys, filepath.Join(dir, "data", "scenario")) ||
		statFile(fsys, filepath.Join(dir, "data", "system", "Config.tjs"))
}

func layoutAt(root string) *ProjectLayout {
	return &ProjectLayout{
		Root:        root,
		ScenarioDir: filepath.Join(root, "data", "scenario"),
		ConfigPath:  filepath.Join(root, "data", "system", "Config.tjs"),
	}
}

func statDir(fsys fs.FS, p string) bool {
	info, err := fs.Stat(fsys, fsPath(p))
	return err == nil && info.IsDir()
}

func statFile(fsys fs.FS, p string) bool {
	info, err := fs.Stat(fsys, fsPath(p))
	return err == nil && !info.IsDir()
}

func fsPath(p string) string {
	clean := filepath.Clean(p)
	if filepath.IsAbs(clean) {
		return clean
	}
	return strings.TrimPrefix(filepath.ToSlash(clean), "/")
}

type osFS struct{}

func (osFS) Open(name string) (fs.File, error) { return os.Open(name) }
