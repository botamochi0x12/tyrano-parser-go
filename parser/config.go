package parser

import (
	"strings"
	"unicode"

	"github.com/botamochi0x12/tyrano-parser-go/types"
)

// ConfigParser handles parsing of Config.tjs files
type ConfigParser struct {
	options ParserOptions
	result  *types.ParseResult
}

// NewConfigParser creates a new ConfigParser with default options
func NewConfigParser(strictMode bool) *ConfigParser {
	return &ConfigParser{
		options: ParserOptions{
			StrictMode:     strictMode,
			EnableWarnings: true,
		},
		result: types.NewParseResult(),
	}
}

// NewConfigParserWithOptions creates a new ConfigParser with specified options
func NewConfigParserWithOptions(options ParserOptions) *ConfigParser {
	return &ConfigParser{
		options: options,
		result:  types.NewParseResult(),
	}
}

// Parse parses Config.tjs content and returns a ConfigMap
func (cp *ConfigParser) Parse(content string) (types.ConfigMap, error) {
	config, result := cp.ParseWithResult(content)

	if cp.options.StrictMode && result.HasErrors() {
		return nil, result
	}

	return config, nil
}

// ParseWithResult parses Config.tjs content and returns both the ConfigMap and ParseResult.
// Real-format grammar:
//
//	;key = value;       (semicolon prefix marks an assignment; trailing ; optional)
//	;key = value; // c  (inline // comment after value, quote-aware)
//	// pure comment
//	/* block */         (single-line only)
//	<other lines>       (TJS code outside our subset; ignored unless looks like assignment)
func (cp *ConfigParser) ParseWithResult(content string) (types.ConfigMap, *types.ParseResult) {
	cp.result = types.NewParseResult()
	config := types.NewConfigMap()

	for lineNum, raw := range strings.Split(content, "\n") {
		trimmed := strings.TrimSpace(raw)
		if trimmed == "" {
			continue
		}
		if strings.HasPrefix(trimmed, "//") {
			continue
		}
		if strings.HasPrefix(trimmed, "/*") && strings.HasSuffix(trimmed, "*/") {
			continue
		}
		if strings.HasPrefix(trimmed, ";") {
			cp.parseConfigLine(strings.TrimSpace(trimmed[1:]), lineNum+1, config)
			if cp.options.StrictMode && cp.result.HasErrors() {
				break
			}
			continue
		}
		if strings.Contains(trimmed, "=") {
			issue := types.NewParseWarning(types.InvalidConfigError, lineNum+1, 1,
				"unrecognized assignment form (missing leading ';' prefix)")
			issue.WithContext(trimmed)
			cp.result.AddIssue(*issue)
		}
	}

	return config, cp.result
}

// parseConfigLine parses a config assignment after the leading ";" has been stripped.
// Handles inline comments, optional trailing semicolon, quoted values.
func (cp *ConfigParser) parseConfigLine(line string, lineNum int, config types.ConfigMap) {
	originalLine := line
	if line == "" {
		return
	}

	line = strings.TrimSpace(stripInlineComment(line))
	if line == "" {
		return
	}

	line = strings.TrimSuffix(line, ";")
	line = strings.TrimSpace(line)

	equalsIndex := strings.Index(line, "=")
	if equalsIndex == -1 {
		issue := types.NewParseError(types.InvalidConfigError, lineNum, 1,
			"missing equals sign in config line")
		issue.WithContext(originalLine)
		cp.result.AddIssue(*issue)
		return
	}

	key := strings.TrimSpace(line[:equalsIndex])
	value := strings.TrimSpace(line[equalsIndex+1:])

	if key == "" {
		issue := types.NewParseError(types.InvalidConfigError, lineNum, 1,
			"empty key in config line")
		issue.WithContext(originalLine)
		cp.result.AddIssue(*issue)
		return
	}

	if err := cp.validateQuotedValue(value, lineNum, originalLine); err != nil {
		cp.result.AddIssue(*err)
		if cp.options.StrictMode {
			return
		}
	}

	value = cp.handleQuotedValue(value)

	if value == "" {
		severity := types.Error
		if !cp.options.StrictMode {
			severity = types.Warning
		}
		issue := types.NewParseIssue(types.InvalidConfigError, severity, lineNum,
			len(originalLine), "empty value in config line")
		issue.WithContext(originalLine)
		cp.result.AddIssue(*issue)
		if cp.options.StrictMode {
			return
		}
	}

	config.Set(key, value)
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
func (cp *ConfigParser) validateQuotedValue(value string, lineNum int, context string) *types.ParseIssue {
	value = strings.TrimSpace(value)

	// Check for unmatched quotes
	if len(value) > 0 {
		if strings.HasPrefix(value, "\"") && !strings.HasSuffix(value, "\"") {
			issue := types.NewParseError(types.InvalidConfigError, lineNum, 1, "unmatched double quote")
			issue.WithContext(context)
			return issue
		}
		if strings.HasPrefix(value, "'") && !strings.HasSuffix(value, "'") {
			issue := types.NewParseError(types.InvalidConfigError, lineNum, 1, "unmatched single quote")
			issue.WithContext(context)
			return issue
		}
		if !strings.HasPrefix(value, "\"") && strings.HasSuffix(value, "\"") {
			issue := types.NewParseError(types.InvalidConfigError, lineNum, 1, "unmatched double quote")
			issue.WithContext(context)
			return issue
		}
		if !strings.HasPrefix(value, "'") && strings.HasSuffix(value, "'") {
			issue := types.NewParseError(types.InvalidConfigError, lineNum, 1, "unmatched single quote")
			issue.WithContext(context)
			return issue
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

// stripInlineComment removes a trailing "// ..." comment from the value side
// of a config line, respecting quoted regions. Backslash-escapes are honored
// inside quotes. Returns the input unchanged if no unquoted "//" exists.
func stripInlineComment(s string) string {
	inSingle := false
	inDouble := false
	i := 0
	for i < len(s) {
		c := s[i]
		if c == '\\' && i+1 < len(s) && (inSingle || inDouble) {
			i += 2
			continue
		}
		if c == '"' && !inSingle {
			inDouble = !inDouble
			i++
			continue
		}
		if c == '\'' && !inDouble {
			inSingle = !inSingle
			i++
			continue
		}
		if !inSingle && !inDouble && c == '/' && i+1 < len(s) && s[i+1] == '/' {
			return strings.TrimRight(s[:i], " \t")
		}
		i++
	}
	return s
}

// isWhitespace checks if a rune is whitespace
func isWhitespace(r rune) bool {
	return unicode.IsSpace(r)
}
