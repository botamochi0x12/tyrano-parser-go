package loader

import (
	"path/filepath"
	"testing"

	"github.com/botamochi0x12/tyrano-parser-go/parser"
	"github.com/botamochi0x12/tyrano-parser-go/types"
)

// macroProject writes a project whose choices go through a macro defined in a
// separate file, the shape that made every branch report an empty text and a
// raw &mp.* target.
func macroProject(t *testing.T) *ProjectLayout {
	t.Helper()
	tmp := t.TempDir()
	scn := filepath.Join(tmp, "data", "scenario")
	writeFile(t, filepath.Join(scn, "macro.ks"),
		"[macro name=\"sel\"]\n[link storage=%storage target=%target]%text[endlink]\n[endmacro]\n")
	writeFile(t, filepath.Join(scn, "first.ks"),
		"*start\n@call storage=\"macro.ks\"\n[sel storage=\"route_a.ks\" target=\"*good_end\" text=\"森へ行く\"]\n[s]\n")
	writeFile(t, filepath.Join(scn, "route_a.ks"), "*good_end\n[s]\n")
	layout, err := LayoutFrom(tmp)
	if err != nil {
		t.Fatal(err)
	}
	return layout
}

func choiceRef(t *testing.T, scan *types.ProjectScan) types.ScenarioRefRecord {
	t.Helper()
	for _, r := range scan.Refs {
		if r.Kind == RefKindChoice {
			return r
		}
	}
	t.Fatalf("no choice ref in %#v", scan.Refs)
	return types.ScenarioRefRecord{}
}

func TestScanProject_ResolvesChoiceThroughMacroDefinedElsewhere(t *testing.T) {
	scan, err := ScanProject(macroProject(t), parser.NewDefaultTyranoParser())
	if err != nil {
		t.Fatal(err)
	}
	r := choiceRef(t, scan)
	if r.From != "first.ks" {
		t.Errorf("From = %q, want the call site first.ks", r.From)
	}
	if r.Text != "森へ行く" {
		t.Errorf("Text = %q, want %q", r.Text, "森へ行く")
	}
	if r.Storage != "route_a.ks" || r.Target != "*good_end" || r.Label != "good_end" {
		t.Errorf("ref = %#v, want the call site destination with an asterisk-free label", r)
	}
	if !r.Resolved || r.UI {
		t.Errorf("ref = %#v, want a resolved story choice", r)
	}
	if scan.Issues.HasErrors() {
		t.Errorf("unexpected errors: %v", scan.Issues.GetErrors())
	}
}

func TestResolveFrom_ReachesStorageOnlyAMacroCallNames(t *testing.T) {
	layout := macroProject(t)
	scan, err := ResolveFrom(layout, parser.NewDefaultTyranoParser(), "first.ks")
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := scan.Scenarios["route_a.ks"]; !ok {
		t.Errorf("scenarios = %#v, want route_a.ks reached through the macro call", keysOf(scan.Scenarios))
	}
	r := choiceRef(t, scan)
	if r.Text != "森へ行く" || r.Storage != "route_a.ks" {
		t.Errorf("ref = %#v, want the call site text and destination", r)
	}
	if scan.Issues.HasErrors() {
		t.Errorf("unexpected errors: %v", scan.Issues.GetErrors())
	}
}

func keysOf(m map[string]*types.ParsedScenario) []string {
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	return out
}
