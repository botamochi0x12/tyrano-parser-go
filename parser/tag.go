package parser

import "github.com/tyranoscript/tyrano-parser-go/types"

// TagParser handles parsing of individual tags and their parameters
type TagParser struct {
	keepSpaceConfig string
}

// NewTagParser creates a new TagParser
func NewTagParser(keepSpaceConfig string) *TagParser {
	return &TagParser{
		keepSpaceConfig: keepSpaceConfig,
	}
}

// ParseTag parses a tag string and returns a ParsedTag
func (tp *TagParser) ParseTag(tagStr string, lineNum int) (*types.ParsedTag, error) {
	// Implementation will be added in later tasks
	return types.NewParsedTag("", lineNum), nil
}

// extractParameters extracts parameters from a parameter string
func (tp *TagParser) extractParameters(paramStr string) (map[string]string, error) {
	// Implementation will be added in later tasks
	return make(map[string]string), nil
}

// handleQuotedValue handles quoted parameter values
func (tp *TagParser) handleQuotedValue(value string, quote rune) string {
	// Implementation will be added in later tasks
	return value
}
