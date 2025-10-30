package parser

import (
	"strings"
	"unicode"

	"github.com/tyranoscript/tyrano-parser-go/types"
)

// ConfigParser handles parsing of Config.tjs files
type ConfigParser struct {
	strictMode bool
}

// NewConfigParser creates a new ConfigParser
func NewConfigParser(strictMode bool) *ConfigParser {
	return &ConfigParser{
		strictMode: strictMode,
	}
}

// Parse parses Config.tjs content and returns a ConfigMap
func (cp *ConfigParser) Parse(content string) (types.ConfigMap, error) {
	config := types.NewConfigMap()
	
	// Split content into lines for processing
	lines := strings.Split(content, "\n")
	
	for lineNum, line := range lines {
		// Trim whitespace
		trimmedLine := strings.TrimSpace(line)
		
		// Skip empty lines
		if trimmedLine == "" {
			continue
		}
		
		// Skip JavaScript-style comments (// and /* */)
		if strings.HasPrefix(trimmedLine, "//") {
			continue
		}
		
		// Handle block comments (/* ... */) - simple approach for single-line blocks
		if strings.HasPrefix(trimmedLine, "/*") && strings.HasSuffix(trimmedLine, "*/") {
			continue
		}
		
		// Process config lines that start with semicolon
		if strings.HasPrefix(trimmedLine, ";") {
			err := cp.parseConfigLine(trimmedLine, lineNum+1, config)
			if err != nil && cp.strictMode {
				return nil, err
			}
			// In lenient mode, we continue even if there's an error
		}
		
		// Skip any other lines (like regular text, multi-line comments, etc.)
	}
	
	return config, nil
}

// parseConfigLine parses a single configuration line
func (cp *ConfigParser) parseConfigLine(line string, lineNum int, config types.ConfigMap) error {
	// Remove leading semicolon and trim whitespace
	line = strings.TrimPrefix(line, ";")
	line = strings.TrimSpace(line)
	
	// Skip empty lines
	if line == "" {
		return nil
	}
	
	// Find the equals sign
	equalsIndex := strings.Index(line, "=")
	if equalsIndex == -1 {
		if cp.strictMode {
			return types.NewParseError(types.InvalidConfigError, lineNum, 1, "missing equals sign in config line")
		}
		return nil
	}
	
	// Extract key and value
	key := strings.TrimSpace(line[:equalsIndex])
	valueWithSemicolon := strings.TrimSpace(line[equalsIndex+1:])
	
	// Validate key
	if key == "" {
		if cp.strictMode {
			return types.NewParseError(types.InvalidConfigError, lineNum, 1, "empty key in config line")
		}
		return nil
	}
	
	// Remove trailing semicolon if present
	value := valueWithSemicolon
	if strings.HasSuffix(value, ";") {
		value = strings.TrimSuffix(value, ";")
		value = strings.TrimSpace(value)
	} else if cp.strictMode {
		return types.NewParseError(types.InvalidConfigError, lineNum, 1, "missing semicolon at end of config line")
	}
	
	// Validate quoted values
	if err := cp.validateQuotedValue(value, lineNum); err != nil && cp.strictMode {
		return err
	}
	
	// Handle quoted values
	value = cp.handleQuotedValue(value)
	
	// Validate that we have a value
	if value == "" && cp.strictMode {
		return types.NewParseError(types.InvalidConfigError, lineNum, 1, "empty value in config line")
	}
	
	// Store the key-value pair
	config.Set(key, value)
	
	return nil
}

// handleQuotedValue processes quoted and unquoted values
func (cp *ConfigParser) handleQuotedValue(value string) string {
	value = strings.TrimSpace(value)
	
	// Handle quoted strings
	if len(value) >= 2 {
		if (strings.HasPrefix(value, "\"") && strings.HasSuffix(value, "\"")) ||
			(strings.HasPrefix(value, "'") && strings.HasSuffix(value, "'")) {
			// Remove outer quotes
			unquoted := value[1 : len(value)-1]
			// Handle escape sequences
			return cp.processEscapeSequences(unquoted)
		}
	}
	
	// Return unquoted value as-is
	return value
}

// validateQuotedValue checks if quotes are properly matched
func (cp *ConfigParser) validateQuotedValue(value string, lineNum int) error {
	value = strings.TrimSpace(value)
	
	// Check for unmatched quotes
	if len(value) > 0 {
		if strings.HasPrefix(value, "\"") && !strings.HasSuffix(value, "\"") {
			return types.NewParseError(types.InvalidConfigError, lineNum, 1, "unmatched double quote")
		}
		if strings.HasPrefix(value, "'") && !strings.HasSuffix(value, "'") {
			return types.NewParseError(types.InvalidConfigError, lineNum, 1, "unmatched single quote")
		}
		if !strings.HasPrefix(value, "\"") && strings.HasSuffix(value, "\"") {
			return types.NewParseError(types.InvalidConfigError, lineNum, 1, "unmatched double quote")
		}
		if !strings.HasPrefix(value, "'") && strings.HasSuffix(value, "'") {
			return types.NewParseError(types.InvalidConfigError, lineNum, 1, "unmatched single quote")
		}
	}
	
	return nil
}

// processEscapeSequences handles escape sequences in quoted strings
func (cp *ConfigParser) processEscapeSequences(s string) string {
	var result strings.Builder
	runes := []rune(s)
	
	for i := 0; i < len(runes); i++ {
		if runes[i] == '\\' && i+1 < len(runes) {
			switch runes[i+1] {
			case 'n':
				result.WriteRune('\n')
				i++ // Skip the next character
			case 't':
				result.WriteRune('\t')
				i++
			case 'r':
				result.WriteRune('\r')
				i++
			case '\\':
				result.WriteRune('\\')
				i++
			case '"':
				result.WriteRune('"')
				i++
			case '\'':
				result.WriteRune('\'')
				i++
			default:
				// Unknown escape sequence, keep both characters
				result.WriteRune(runes[i])
			}
		} else {
			result.WriteRune(runes[i])
		}
	}
	
	return result.String()
}

// isWhitespace checks if a rune is whitespace
func isWhitespace(r rune) bool {
	return unicode.IsSpace(r)
}
