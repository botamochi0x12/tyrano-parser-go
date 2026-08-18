package loader

import (
	"testing"

	"github.com/botamochi0x12/tyrano-parser-go/parser"
)

func refsOf(t *testing.T, from, src string) []ScenarioRef {
	t.Helper()
	scenario, _ := parser.NewDefaultTyranoParser().ParseScenarioWithResult(src)
	return ExtractRefs(from, scenario)
}

func soleRef(t *testing.T, from, src string) ScenarioRef {
	t.Helper()
	refs := refsOf(t, from, src)
	if len(refs) != 1 {
		t.Fatalf("ExtractRefs() returned %d refs, want 1: %#v", len(refs), refs)
	}
	return refs[0]
}

func TestExtractRefs_LinkBodyBecomesChoiceText(t *testing.T) {
	r := soleRef(t, "menu.ks", "[link storage=\"other.ks\" target=\"*x\"]森へ行く[endlink]\n")
	if r.Text != "森へ行く" {
		t.Errorf("Text = %q, want %q", r.Text, "森へ行く")
	}
}

func TestExtractRefs_LinkBodySpanningLinesBecomesChoiceText(t *testing.T) {
	r := soleRef(t, "menu.ks", "[link target=\"*x\"]\n川へ行く\n[endlink]\n")
	if r.Text != "川へ行く" {
		t.Errorf("Text = %q, want %q", r.Text, "川へ行く")
	}
}

func TestExtractRefs_GlinkTextParameterBecomesChoiceText(t *testing.T) {
	r := soleRef(t, "menu.ks", "[glink storage=\"other.ks\" target=\"*x\" text=\"街へ行く\"]\n")
	if r.Tag != "glink" || r.Text != "街へ行く" {
		t.Errorf("ref = %#v, want glink with text %q", r, "街へ行く")
	}
}

func TestExtractRefs_ButtonFamilyIsCollected(t *testing.T) {
	src := "[button graphic=\"a.png\" target=\"*x\" text=\"開始\"]\n" +
		"[s_button storage=\"b.ks\" text=\"設定\"]\n" +
		"[showbutton target=\"*y\" text=\"戻る\"]\n"
	refs := refsOf(t, "menu.ks", src)
	if len(refs) != 3 {
		t.Fatalf("ExtractRefs() returned %d refs, want 3: %#v", len(refs), refs)
	}
	for _, r := range refs {
		if r.Text == "" {
			t.Errorf("ref %#v has no choice text", r)
		}
		if r.Kind != RefKindChoice {
			t.Errorf("ref %#v Kind = %q, want %q", r, r.Kind, RefKindChoice)
		}
	}
}

func TestExtractRefs_JumpAndCallAreFlowKind(t *testing.T) {
	refs := refsOf(t, "first.ks", "@jump storage=\"title.ks\"\n@call storage=\"sub.ks\"\n")
	if len(refs) != 2 {
		t.Fatalf("ExtractRefs() returned %d refs, want 2: %#v", len(refs), refs)
	}
	for _, r := range refs {
		if r.Kind != RefKindFlow {
			t.Errorf("ref %#v Kind = %q, want %q", r, r.Kind, RefKindFlow)
		}
	}
}

func TestExtractRefs_LabelStripsAsterisk(t *testing.T) {
	r := soleRef(t, "first.ks", "@jump target=\"*ending\"\n")
	if r.Target != "*ending" {
		t.Errorf("Target = %q, want the raw %q", r.Target, "*ending")
	}
	if r.Label != "ending" {
		t.Errorf("Label = %q, want %q", r.Label, "ending")
	}
}

func TestExtractRefs_DynamicTargetIsReportedUnresolved(t *testing.T) {
	r := soleRef(t, "macro_call.ks", "@jump storage=&f.next target=&mp.target\n")
	if r.Resolved {
		t.Errorf("ref %#v Resolved = true, want false for runtime values", r)
	}
	if len(r.Dynamic) != 2 || r.Dynamic[0] != "storage" || r.Dynamic[1] != "target" {
		t.Errorf("Dynamic = %#v, want [storage target]", r.Dynamic)
	}
	if r.Label != "" {
		t.Errorf("Label = %q, want empty for an unresolved target", r.Label)
	}
}

func TestExtractRefs_StaticRefIsReportedResolved(t *testing.T) {
	r := soleRef(t, "first.ks", "@jump storage=\"title.ks\" target=\"*start\"\n")
	if !r.Resolved || len(r.Dynamic) != 0 {
		t.Errorf("ref %#v, want Resolved with no dynamic fields", r)
	}
}

func TestExtractRefs_LinkWithoutEndlinkStopsAtTheNextStop(t *testing.T) {
	src := "[link storage=\"a.ks\"]森へ行く\n[s]\nこれは選択肢ではない本文。\n"
	r := soleRef(t, "menu.ks", src)
	if r.Text != "森へ行く" {
		t.Errorf("Text = %q, want %q — a missing [endlink] must not swallow the narration", r.Text, "森へ行く")
	}
}

func TestExtractRefs_LinkBodyStopsAtTheNextLabel(t *testing.T) {
	src := "[link target=\"*x\"]戻る\n*x\n別の場面の本文。\n"
	r := soleRef(t, "menu.ks", src)
	if r.Text != "戻る" {
		t.Errorf("Text = %q, want %q", r.Text, "戻る")
	}
}

func TestExtractRefs_TextlessRuntimeButtonIsFlaggedUI(t *testing.T) {
	r := soleRef(t, "config.ks", "[button target=&mp.target]\n")
	if !r.UI {
		t.Errorf("ref = %#v, want UI true for a textless button with a runtime target", r)
	}
}

func TestExtractRefs_AuthoredChoiceIsNotFlaggedUI(t *testing.T) {
	r := soleRef(t, "scene1.ks", "[link storage=\"a.ks\" target=\"*x\"]森へ行く[endlink]\n")
	if r.UI {
		t.Errorf("ref = %#v, want UI false for an authored choice", r)
	}
}

func TestExtractRefs_TextlessStaticButtonIsNotFlaggedUI(t *testing.T) {
	r := soleRef(t, "title.ks", "[button graphic=\"start.png\" storage=\"first.ks\"]\n")
	if r.UI {
		t.Errorf("ref = %#v, want UI false when the destination is static", r)
	}
}

func TestExtractRefs_FlowTagIsNeverFlaggedUI(t *testing.T) {
	r := soleRef(t, "config.ks", "@jump storage=&f.next\n")
	if r.UI {
		t.Errorf("ref = %#v, want UI false for control flow", r)
	}
}

func TestExtractRefs_SkipsMacroDefinitionBody(t *testing.T) {
	src := "[macro name=\"sel\"]\n[link storage=%storage target=%target]%text[endlink]\n[endmacro]\n" +
		"@jump storage=\"title.ks\"\n"
	refs := refsOf(t, "macro.ks", src)
	if len(refs) != 1 || refs[0].Tag != "jump" {
		t.Errorf("ExtractRefs() = %#v, want only the jump outside the macro body", refs)
	}
}
