package parser

import (
	"fmt"
	"strings"
	"unicode"

	"github.com/botamochi0x12/tyrano-parser-go/types"
)

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
	// Validate basic tag format
	if !strings.HasPrefix(tagStr, "[") || !strings.HasSuffix(tagStr, "]") {
		return nil, fmt.Errorf("invalid tag format: missing brackets at line %d", lineNum)
	}

	// Remove brackets
	content := tagStr[1 : len(tagStr)-1]
	content = strings.TrimSpace(content)

	if content == "" {
		return nil, fmt.Errorf("empty tag at line %d", lineNum)
	}

	// Split tag name from parameters
	parts := strings.Fields(content)
	if len(parts) == 0 {
		return nil, fmt.Errorf("empty tag at line %d", lineNum)
	}

	tagName := parts[0]

	// Create parsed tag
	parsedTag := types.NewParsedTag(tagName, lineNum)

	// Extract parameters if present
	if len(parts) > 1 {
		// Rejoin the parameter part (everything after the tag name)
		paramStart := strings.Index(content, tagName) + len(tagName)
		paramStr := strings.TrimSpace(content[paramStart:])

		if paramStr != "" {
			params, err := tp.extractParameters(paramStr)
			if err != nil {
				return nil, fmt.Errorf("error parsing parameters at line %d: %v", lineNum, err)
			}
			parsedTag.Parameters = params
		}
	}

	return parsedTag, nil
}

// extractParameters extracts parameters from a parameter string
func (tp *TagParser) extractParameters(paramStr string) (map[string]string, error) {
	params := make(map[string]string)

	if paramStr == "" {
		return params, nil
	}

	// Use a simple state machine to parse parameters
	i := 0
	runes := []rune(paramStr)

	for i < len(runes) {
		// Skip whitespace
		for i < len(runes) && unicode.IsSpace(runes[i]) {
			i++
		}

		if i >= len(runes) {
			break
		}

		// Read parameter name
		nameStart := i
		for i < len(runes) && runes[i] != '=' && !unicode.IsSpace(runes[i]) {
			i++
		}

		if i >= len(runes) || runes[i] != '=' {
			return nil, fmt.Errorf("invalid parameter format: expected '=' after parameter name")
		}

		paramName := string(runes[nameStart:i])
		i++ // Skip '='

		// Skip whitespace after '='
		for i < len(runes) && unicode.IsSpace(runes[i]) {
			i++
		}

		if i >= len(runes) {
			return nil, fmt.Errorf("parameter '%s' has no value", paramName)
		}

		// Read parameter value
		var paramValue string
		var err error

		if runes[i] == '"' || runes[i] == '\'' {
			// Quoted value
			quote := runes[i]
			i++ // Skip opening quote
			valueStart := i

			// Find closing quote, handling escape sequences
			for i < len(runes) {
				if runes[i] == '\\' && i+1 < len(runes) {
					i += 2 // Skip escape sequence
				} else if runes[i] == quote {
					break
				} else {
					i++
				}
			}

			if i >= len(runes) {
				return nil, fmt.Errorf("unclosed quoted parameter value for '%s'", paramName)
			}

			rawValue := string(runes[valueStart:i])
			paramValue = tp.handleQuotedValue(rawValue, quote)
			i++ // Skip closing quote
		} else {
			// Unquoted value - read until whitespace
			valueStart := i
			for i < len(runes) && !unicode.IsSpace(runes[i]) {
				// Check for invalid characters in unquoted values
				if runes[i] == '=' {
					return nil, fmt.Errorf("invalid parameter format: unexpected '=' in unquoted value for parameter '%s'", paramName)
				}
				i++
			}
			paramValue = string(runes[valueStart:i])
		}

		if err != nil {
			return nil, err
		}

		params[paramName] = paramValue
	}

	return params, nil
}

// handleQuotedValue handles quoted parameter values with escape sequences and space handling
func (tp *TagParser) handleQuotedValue(value string, quote rune) string {
	// Handle escape sequences
	result := strings.ReplaceAll(value, "\\n", "\n")
	result = strings.ReplaceAll(result, "\\t", "\t")
	result = strings.ReplaceAll(result, "\\r", "\r")
	result = strings.ReplaceAll(result, "\\\\", "\\")
	result = strings.ReplaceAll(result, "\\\"", "\"")
	result = strings.ReplaceAll(result, "\\'", "'")

	// Handle space trimming based on configuration
	if tp.keepSpaceConfig != "true" {
		result = strings.TrimSpace(result)
	}

	return result
}
