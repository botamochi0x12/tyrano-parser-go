package parser

import (
	"os"
	"path/filepath"
	"testing"
)

func TestConfigParser_RealStarGazersConfig(t *testing.T) {
	wd, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
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
