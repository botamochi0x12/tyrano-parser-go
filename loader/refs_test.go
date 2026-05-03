package loader

import (
	"testing"

	"github.com/botamochi0x12/tyrano-parser-go/parser"
)

func TestExtractRefs_CallAndJumpStorage(t *testing.T) {
	src := "*start\n@call storage=\"tyrano.ks\"\n@jump storage=\"title.ks\"\n@jump target=\"*ending\"\n[s]\n"
	scenario, _ := parser.NewDefaultTyranoParser().ParseScenarioWithResult(src)
	refs := ExtractRefs("first.ks", scenario)
	if len(refs) != 3 {
		t.Fatalf("expected 3 refs, got %d: %#v", len(refs), refs)
	}
	got := map[string]ScenarioRef{}
	for _, r := range refs {
		got[r.Tag+":"+r.Storage+":"+r.Target] = r
		if r.From != "first.ks" {
			t.Errorf("unexpected From: %q", r.From)
		}
	}
	for _, key := range []string{"call:tyrano.ks:", "jump:title.ks:", "jump::*ending"} {
		if _, ok := got[key]; !ok {
			t.Errorf("missing %s in refs: %#v", key, refs)
		}
	}
}

func TestExtractRefs_LinkTag(t *testing.T) {
	src := "[link storage=\"other.ks\" target=\"*x\"]choose me[endlink]\n[s]\n"
	scenario, _ := parser.NewDefaultTyranoParser().ParseScenarioWithResult(src)
	refs := ExtractRefs("menu.ks", scenario)
	if len(refs) != 1 {
		t.Fatalf("expected 1 link ref, got %d", len(refs))
	}
	r := refs[0]
	if r.Tag != "link" || r.Storage != "other.ks" || r.Target != "*x" {
		t.Errorf("unexpected ref: %#v", r)
	}
}

func TestExtractRefs_NilScenario(t *testing.T) {
	if refs := ExtractRefs("x.ks", nil); refs != nil {
		t.Errorf("ExtractRefs(nil) = %#v, want nil", refs)
	}
}
