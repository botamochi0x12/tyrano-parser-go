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
	scan, err := ScanProject(layout, parser.NewDefaultTyranoParser())
	if err != nil {
		t.Fatalf("ScanProject: %v", err)
	}
	if scan.Config["System.title"] != "StarGazers" {
		t.Errorf("config[System.title] = %q, want %q", scan.Config["System.title"], "StarGazers")
	}
	if len(scan.Scenarios) < 1 {
		t.Errorf("expected scenarios to be discovered, got %d", len(scan.Scenarios))
	}
	if t.Failed() {
		t.Logf("scanned %d scenarios, %d refs, %d issues",
			len(scan.Scenarios), len(scan.Refs), scan.Issues.GetIssueCount())
	}
}
