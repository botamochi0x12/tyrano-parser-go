package loader

import (
	"strings"

	"github.com/botamochi0x12/tyrano-parser-go/types"
)

// Ref kinds distinguish an authored branch the player picks from plain control
// flow the engine follows on its own.
const (
	RefKindFlow   = "flow"
	RefKindChoice = "choice"
)

// maxMacroDepth bounds macro-in-macro expansion so a recursive definition
// cannot spin forever.
const maxMacroDepth = 8

// ScenarioRef is one outgoing reference found in a scenario file.
//
// Storage and Target keep whatever the script wrote, including unresolved
// runtime expressions; Label is the Target normalised for display, and Dynamic
// names the fields that are still runtime expressions rather than static values.
type ScenarioRef struct {
	From     string   `json:"from"`
	Line     int      `json:"line"`
	Tag      string   `json:"tag"`
	Kind     string   `json:"kind"`
	Storage  string   `json:"storage"`
	Target   string   `json:"target"`
	Label    string   `json:"label"`
	Text     string   `json:"text"`
	Macro    string   `json:"macro,omitempty"`
	Dynamic  []string `json:"dynamic,omitempty"`
	Resolved bool     `json:"resolved"`
	UI       bool     `json:"ui"`
}

var refTags = map[string]string{
	"call":       RefKindFlow,
	"jump":       RefKindFlow,
	"link":       RefKindChoice,
	"glink":      RefKindChoice,
	"button":     RefKindChoice,
	"s_button":   RefKindChoice,
	"showbutton": RefKindChoice,
}

// ExtractRefs returns the references a scenario makes, without expanding macro
// calls. Use ExtractRefsWithMacros when a project-wide macro table is available.
func ExtractRefs(from string, scenario *types.ParsedScenario) []ScenarioRef {
	return ExtractRefsWithMacros(from, scenario, nil)
}

// ExtractRefsWithMacros returns the references a scenario makes, expanding calls
// to the macros in table so that a branch authored as [my_choice text="..."]
// reports the storage, target and text the call site actually binds.
//
// References written inside a [macro] body are not reported at the definition
// site: they are templates, and only become real branches once called.
func ExtractRefsWithMacros(from string, scenario *types.ParsedScenario, macros MacroTable) []ScenarioRef {
	if scenario == nil {
		return nil
	}
	x := &refExtractor{from: from, macros: macros, out: make([]ScenarioRef, 0)}
	x.walk(skipMacroBodies(scenario.Elements), expansion{})
	return x.out
}

// skipMacroBodies drops the tags between [macro] and [endmacro].
func skipMacroBodies(elements []types.ParsedTag) []types.ParsedTag {
	out := make([]types.ParsedTag, 0, len(elements))
	inMacro := false
	for _, tag := range elements {
		switch tag.Name {
		case "macro":
			inMacro = true
		case "endmacro":
			inMacro = false
		default:
			if !inMacro {
				out = append(out, tag)
			}
		}
	}
	return out
}

type refExtractor struct {
	from   string
	macros MacroTable
	out    []ScenarioRef
}

// expansion carries the macro call context a walk is running under. The zero
// value means "not inside a macro".
type expansion struct {
	args   map[string]string
	macro  string
	line   int // call site line, so refs point at the scenario, not the definition
	depth  int
	active map[string]bool
}

// enter returns the context for expanding one macro call made from this one.
func (e expansion) enter(tag types.ParsedTag, name string, args map[string]string) expansion {
	line := e.line
	if line == 0 {
		line = tag.Line
	}
	active := make(map[string]bool, len(e.active)+1)
	for k := range e.active {
		active[k] = true
	}
	active[name] = true
	return expansion{args: args, macro: name, line: line, depth: e.depth + 1, active: active}
}

func (e expansion) lineOf(tag types.ParsedTag) int {
	if e.line != 0 {
		return e.line
	}
	return tag.Line
}

// walk emits a ref per reference tag in elements, expanding macro calls under
// the given context.
func (x *refExtractor) walk(elements []types.ParsedTag, e expansion) {
	for i, tag := range elements {
		if def, ok := x.macros[tag.Name]; ok && e.depth < maxMacroDepth && !e.active[tag.Name] {
			x.walk(def.Body, e.enter(tag, tag.Name, x.callArgs(tag, e.args)))
			continue
		}
		kind, ok := refTags[tag.Name]
		if !ok {
			continue
		}
		if ref, ok := x.newRef(elements, i, kind, e); ok {
			x.out = append(x.out, ref)
		}
	}
}

// callArgs binds the arguments a macro call site passes, resolving any that are
// themselves placeholders from an enclosing macro.
func (x *refExtractor) callArgs(tag types.ParsedTag, outer map[string]string) map[string]string {
	args := make(map[string]string, len(tag.Parameters))
	for name, value := range tag.Parameters {
		bound, _ := bindMacroValue(value, outer)
		args[name] = bound
	}
	return args
}

func (x *refExtractor) newRef(elements []types.ParsedTag, i int, kind string, e expansion) (ScenarioRef, bool) {
	tag := elements[i]
	args := e.args
	storage, storageOK := bindMacroValue(tag.Parameters["storage"], args)
	target, targetOK := bindMacroValue(tag.Parameters["target"], args)
	if storage == "" && target == "" {
		return ScenarioRef{}, false
	}
	text, textOK := bindMacroText(choiceText(elements, i), args)

	dynamic := make([]string, 0, 3)
	if !storageOK {
		dynamic = append(dynamic, "storage")
	}
	if !targetOK {
		dynamic = append(dynamic, "target")
	}
	if !textOK {
		dynamic = append(dynamic, "text")
	}
	label := ""
	if targetOK {
		label = trimTarget(target)
	}
	return ScenarioRef{
		From:     x.from,
		Line:     e.lineOf(tag),
		Tag:      tag.Name,
		Kind:     kind,
		Storage:  storage,
		Target:   target,
		Label:    label,
		Text:     text,
		Macro:    e.macro,
		Dynamic:  dynamic,
		Resolved: len(dynamic) == 0,
		UI:       kind == RefKindChoice && text == "" && (!storageOK || !targetOK),
	}, true
}

// choiceText returns the label a player reads on a branch: the text= parameter
// when the tag carries one, otherwise the body of a [link] ... [endlink] pair.
func choiceText(elements []types.ParsedTag, i int) string {
	if text := elements[i].Parameters["text"]; text != "" {
		return text
	}
	if elements[i].Name != "link" {
		return ""
	}
	var body strings.Builder
	for _, tag := range elements[i+1:] {
		if tag.Name == "endlink" || tag.Name == "link" {
			break
		}
		if tag.Name == "text" {
			body.WriteString(tag.Value)
		}
	}
	return strings.TrimSpace(body.String())
}
