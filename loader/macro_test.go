package loader

import (
	"testing"

	"github.com/botamochi0x12/tyrano-parser-go/parser"
)

func TestCollectMacros_RecordsBodyBetweenMacroAndEndmacro(t *testing.T) {
	src := "[macro name=\"sel\"]\n[link storage=%storage target=%target]%text[endlink]\n[endmacro]\n"
	scenario, _ := parser.NewDefaultTyranoParser().ParseScenarioWithResult(src)

	macros := CollectMacros("macro.ks", scenario)

	def, ok := macros["sel"]
	if !ok {
		t.Fatalf("CollectMacros() = %#v, want a definition named \"sel\"", macros)
	}
	if def.From != "macro.ks" {
		t.Errorf("def.From = %q, want %q", def.From, "macro.ks")
	}
	names := make([]string, 0, len(def.Body))
	for _, tag := range def.Body {
		names = append(names, tag.Name)
	}
	if len(names) != 3 || names[0] != "link" || names[1] != "text" || names[2] != "endlink" {
		t.Errorf("def.Body names = %#v, want [link text endlink]", names)
	}
}

func TestCollectMacros_IgnoresUnnamedMacro(t *testing.T) {
	src := "[macro]\n[jump storage=\"a.ks\"]\n[endmacro]\n"
	scenario, _ := parser.NewDefaultTyranoParser().ParseScenarioWithResult(src)

	if macros := CollectMacros("macro.ks", scenario); len(macros) != 0 {
		t.Errorf("CollectMacros() = %#v, want no definitions for an unnamed macro", macros)
	}
}
