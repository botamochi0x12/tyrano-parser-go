package types

// ParsedScenario represents a parsed scenario file with elements and labels
type ParsedScenario struct {
	Elements []ParsedTag            `json:"array_s"`
	Labels   map[string]*LabelInfo  `json:"map_label"`
}

// ParsedTag represents a parsed tag with name, parameters, and metadata
type ParsedTag struct {
	Line              int               `json:"line"`
	Name              string            `json:"name"`
	Parameters        map[string]string `json:"pm"`
	Value             string            `json:"val"`
	IsEntityDisabled  bool              `json:"is_entity_disabled,omitempty"`
}

// LabelInfo represents information about a scenario label
type LabelInfo struct {
	Line      int    `json:"line"`
	Index     int    `json:"index"`
	LabelName string `json:"label_name"`
	Value     string `json:"val"`
}

// NewParsedScenario creates a new ParsedScenario with initialized maps and slices
func NewParsedScenario() *ParsedScenario {
	return &ParsedScenario{
		Elements: make([]ParsedTag, 0),
		Labels:   make(map[string]*LabelInfo),
	}
}

// NewParsedTag creates a new ParsedTag with initialized parameters map
func NewParsedTag(name string, line int) *ParsedTag {
	return &ParsedTag{
		Name:       name,
		Line:       line,
		Parameters: make(map[string]string),
		Value:      "",
	}
}

// NewLabelInfo creates a new LabelInfo
func NewLabelInfo(labelName string, line, index int, value string) *LabelInfo {
	return &LabelInfo{
		LabelName: labelName,
		Line:      line,
		Index:     index,
		Value:     value,
	}
}
