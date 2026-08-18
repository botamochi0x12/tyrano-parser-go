package loader

import (
	"testing"

	"github.com/botamochi0x12/tyrano-parser-go/parser"
)

func macrosOf(t *testing.T, from, src string) MacroTable {
	t.Helper()
	scenario, _ := parser.NewDefaultTyranoParser().ParseScenarioWithResult(src)
	return CollectMacros(from, scenario)
}

func refsWithMacros(t *testing.T, from, src string, macros MacroTable) []ScenarioRef {
	t.Helper()
	scenario, _ := parser.NewDefaultTyranoParser().ParseScenarioWithResult(src)
	return ExtractRefsWithMacros(from, scenario, macros)
}

func TestExtractRefsWithMacros_BindsCallSiteArguments(t *testing.T) {
	macros := macrosOf(t, "macro.ks",
		"[macro name=\"sel\"]\n[link storage=%storage target=%target]%text[endlink]\n[endmacro]\n")

	refs := refsWithMacros(t, "scene1.ks",
		"[sel storage=\"route_a.ks\" target=\"*good_end\" text=\"森へ行く\"]\n", macros)

	if len(refs) != 1 {
		t.Fatalf("ExtractRefsWithMacros() returned %d refs, want 1: %#v", len(refs), refs)
	}
	r := refs[0]
	if r.Storage != "route_a.ks" || r.Target != "*good_end" {
		t.Errorf("ref = %#v, want the call site storage and target", r)
	}
	if r.Text != "森へ行く" {
		t.Errorf("Text = %q, want %q", r.Text, "森へ行く")
	}
	if r.Label != "good_end" {
		t.Errorf("Label = %q, want %q", r.Label, "good_end")
	}
	if !r.Resolved {
		t.Errorf("Resolved = false, want true once the call site binds every value")
	}
	if r.Macro != "sel" {
		t.Errorf("Macro = %q, want %q", r.Macro, "sel")
	}
	if r.From != "scene1.ks" || r.Line != 1 {
		t.Errorf("ref located at %s:%d, want the call site scene1.ks:1", r.From, r.Line)
	}
}

func TestExtractRefsWithMacros_ResolvesMpReferences(t *testing.T) {
	macros := macrosOf(t, "macro.ks",
		"[macro name=\"go\"]\n[jump storage=&mp.storage target=&mp.target]\n[endmacro]\n")

	refs := refsWithMacros(t, "scene1.ks", "[go storage=\"b.ks\" target=\"*x\"]\n", macros)

	if len(refs) != 1 {
		t.Fatalf("ExtractRefsWithMacros() returned %d refs, want 1: %#v", len(refs), refs)
	}
	if refs[0].Storage != "b.ks" || refs[0].Target != "*x" || !refs[0].Resolved {
		t.Errorf("ref = %#v, want &mp.* resolved from the call site", refs[0])
	}
}

func TestExtractRefsWithMacros_ReportsArgumentTheCallSiteOmits(t *testing.T) {
	macros := macrosOf(t, "macro.ks",
		"[macro name=\"sel\"]\n[link storage=%storage target=%target]%text[endlink]\n[endmacro]\n")

	refs := refsWithMacros(t, "scene1.ks", "[sel target=\"*x\"]\n", macros)

	if len(refs) != 1 {
		t.Fatalf("ExtractRefsWithMacros() returned %d refs, want 1: %#v", len(refs), refs)
	}
	r := refs[0]
	if r.Resolved {
		t.Errorf("ref = %#v, want Resolved false when the call site omits an argument", r)
	}
	if len(r.Dynamic) != 2 || r.Dynamic[0] != "storage" || r.Dynamic[1] != "text" {
		t.Errorf("Dynamic = %#v, want [storage text]", r.Dynamic)
	}
}

func TestExtractRefsWithMacros_UsesMacroDefaultArgument(t *testing.T) {
	macros := macrosOf(t, "macro.ks",
		"[macro name=\"back\"]\n[jump storage=%storage|title.ks]\n[endmacro]\n")

	refs := refsWithMacros(t, "scene1.ks", "[back]\n", macros)

	if len(refs) != 1 || refs[0].Storage != "title.ks" || !refs[0].Resolved {
		t.Errorf("ExtractRefsWithMacros() = %#v, want the macro default title.ks", refs)
	}
}

func TestExtractRefsWithMacros_ExpandsNestedMacro(t *testing.T) {
	src := "[macro name=\"inner\"]\n[jump storage=%storage]\n[endmacro]\n" +
		"[macro name=\"outer\"]\n[inner storage=%storage]\n[endmacro]\n"
	macros := macrosOf(t, "macro.ks", src)

	refs := refsWithMacros(t, "scene1.ks", "[outer storage=\"deep.ks\"]\n", macros)

	if len(refs) != 1 || refs[0].Storage != "deep.ks" {
		t.Errorf("ExtractRefsWithMacros() = %#v, want storage deep.ks through two macros", refs)
	}
}

func TestExtractRefsWithMacros_StopsOnRecursiveMacro(t *testing.T) {
	macros := macrosOf(t, "macro.ks",
		"[macro name=\"loop\"]\n[loop]\n[jump storage=\"a.ks\"]\n[endmacro]\n")

	refs := refsWithMacros(t, "scene1.ks", "[loop]\n", macros)

	if len(refs) != 1 || refs[0].Storage != "a.ks" {
		t.Errorf("ExtractRefsWithMacros() = %#v, want one ref and no infinite expansion", refs)
	}
}
