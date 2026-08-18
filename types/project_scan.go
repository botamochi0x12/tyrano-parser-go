package types

type ProjectScan struct {
	Root      string                     `json:"root"`
	Config    ConfigMap                  `json:"config"`
	Scenarios map[string]*ParsedScenario `json:"scenarios"`
	Refs      []ScenarioRefRecord        `json:"refs"`
	Issues    *ParseResult               `json:"issues"`
}

// ScenarioRefRecord is one outgoing reference in the scan output.
//
// Storage and Target keep whatever the script wrote, so a value that is only
// known at runtime such as "&f.next" survives verbatim. Label is Target with
// the leading asterisk removed, and is empty while Target is unresolved. Text
// is the wording a player reads on a branch. Kind separates an authored choice
// from plain control flow, Macro names the macro a ref was expanded from,
// Dynamic lists the fields that are still runtime expressions, Resolved is true
// only when none are, and UI marks a textless button bound to a runtime
// destination — a system menu control rather than a story branch.
type ScenarioRefRecord struct {
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
