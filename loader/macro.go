package loader

import (
	"strings"

	"github.com/botamochi0x12/tyrano-parser-go/types"
)

// MacroDef is a [macro name="..."] ... [endmacro] definition. Its Body holds
// the tags between the two markers, still carrying the unbound %arg and
// &mp.arg placeholders that a call site supplies.
type MacroDef struct {
	Name string
	From string
	Line int
	Body []types.ParsedTag
}

// MacroTable maps a macro name to its definition.
type MacroTable map[string]*MacroDef

// CollectMacros returns the macro definitions declared in one scenario.
func CollectMacros(from string, scenario *types.ParsedScenario) MacroTable {
	out := MacroTable{}
	if scenario == nil {
		return out
	}
	var current *MacroDef
	for _, tag := range scenario.Elements {
		switch {
		case tag.Name == "macro":
			name := tag.Parameters["name"]
			if name == "" {
				current = nil
				continue
			}
			current = &MacroDef{Name: name, From: from, Line: tag.Line}
			out[name] = current
		case tag.Name == "endmacro":
			current = nil
		case current != nil:
			current.Body = append(current.Body, tag)
		}
	}
	return out
}

// bindMacroValue substitutes one macro call argument into a value taken from a
// macro body. TyranoScript spells those placeholders either as %name (with an
// optional |default) or as the &mp.name runtime reference. The second return
// value reports whether the result is a static value; an unbound placeholder or
// any other &expression is returned verbatim and reported as unresolved.
func bindMacroValue(value string, args map[string]string) (string, bool) {
	trimmed := strings.TrimSpace(value)
	switch {
	case strings.HasPrefix(trimmed, "%"):
		name, fallback, hasFallback := strings.Cut(trimmed[1:], "|")
		if bound, ok := args[name]; ok {
			return bound, true
		}
		if hasFallback {
			return fallback, true
		}
		return value, false
	case strings.HasPrefix(trimmed, "&mp."):
		if bound, ok := args[trimmed[len("&mp."):]]; ok {
			return bound, true
		}
		return value, false
	case strings.HasPrefix(trimmed, "&"):
		return value, false
	default:
		return value, true
	}
}

// bindMacroText substitutes macro call arguments into body text, where a
// placeholder such as %name sits inside an ordinary sentence rather than
// standing alone. A '%' not followed by an identifier is a literal percent
// sign. The second return value is false when a placeholder stays unbound.
func bindMacroText(value string, args map[string]string) (string, bool) {
	var out strings.Builder
	resolved := true
	runes := []rune(value)
	for i := 0; i < len(runes); {
		if runes[i] != '%' {
			out.WriteRune(runes[i])
			i++
			continue
		}
		end := i + 1
		for end < len(runes) && isMacroNameRune(runes[end]) {
			end++
		}
		if end == i+1 {
			out.WriteRune(runes[i])
			i++
			continue
		}
		name := string(runes[i+1 : end])
		if bound, ok := args[name]; ok {
			out.WriteString(bound)
		} else {
			out.WriteString(string(runes[i:end]))
			resolved = false
		}
		i = end
	}
	return out.String(), resolved
}

func isMacroNameRune(r rune) bool {
	return r == '_' ||
		(r >= 'a' && r <= 'z') ||
		(r >= 'A' && r <= 'Z') ||
		(r >= '0' && r <= '9')
}
