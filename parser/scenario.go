package parser

import (
	"strings"

	"github.com/tyranoscript/tyrano-parser-go/types"
)

// ScenarioParser handles parsing of scenario files
type ScenarioParser struct {
	lexer        *Lexer
	flagScript   bool
	deepIf       int
	currentLine  int
	tagParser    *TagParser
}

// NewScenarioParser creates a new ScenarioParser
func NewScenarioParser() *ScenarioParser {
	return &ScenarioParser{
		flagScript:  false,
		deepIf:      0,
		currentLine: 1,
		tagParser:   NewTagParser("false"),
	}
}

// Parse parses scenario content and returns a ParsedScenario
func (sp *ScenarioParser) Parse(content string) (*types.ParsedScenario, error) {
	scenario := types.NewParsedScenario()
	sp.lexer = NewLexer(content)
	sp.currentLine = 1
	elementIndex := 0
	
	lines := strings.Split(content, "\n")
	inBlockComment := false
	inScriptBlock := false
	scriptContent := []string{}
	scriptStartLine := 0
	
	for lineNum, line := range lines {
		sp.currentLine = lineNum + 1
		originalLine := line
		line = strings.TrimSpace(line)
		
		// Skip empty lines
		if line == "" {
			continue
		}
		
		// Handle block comments
		if inBlockComment {
			if strings.Contains(line, "*/") {
				inBlockComment = false
			}
			continue
		}
		
		if strings.HasPrefix(line, "/*") {
			inBlockComment = true
			if !strings.Contains(line, "*/") {
				continue
			} else {
				inBlockComment = false
				continue
			}
		}
		
		// Skip single line comments
		if strings.HasPrefix(line, ";") {
			continue
		}
		
		// Handle script blocks
		if inScriptBlock {
			if strings.TrimSpace(line) == "[endscript]" {
				// End of script block
				inScriptBlock = false
				
				// Add script content if any
				if len(scriptContent) > 0 {
					contentTag := types.NewParsedTag("script_content", scriptStartLine+1)
					contentTag.Value = strings.Join(scriptContent, "\n")
					scenario.Elements = append(scenario.Elements, *contentTag)
					elementIndex++
				}
				
				// Add endscript tag
				endTag := types.NewParsedTag("endscript", sp.currentLine)
				scenario.Elements = append(scenario.Elements, *endTag)
				elementIndex++
				
				scriptContent = []string{}
				continue
			} else {
				// Collect script content
				scriptContent = append(scriptContent, originalLine)
				continue
			}
		}
		
		if strings.TrimSpace(line) == "[iscript]" {
			// Start of script block
			inScriptBlock = true
			scriptStartLine = sp.currentLine
			
			// Add iscript tag
			scriptTag := types.NewParsedTag("iscript", sp.currentLine)
			scenario.Elements = append(scenario.Elements, *scriptTag)
			elementIndex++
			continue
		}
		
		// Handle different line types
		if strings.HasPrefix(line, "*") {
			// Label line
			tag, labelInfo := sp.parseLabelLine(line)
			if tag != nil {
				tag.Line = sp.currentLine
				scenario.Elements = append(scenario.Elements, *tag)
				
				if labelInfo != nil {
					labelInfo.Line = sp.currentLine
					labelInfo.Index = elementIndex
					scenario.Labels[labelInfo.LabelName] = labelInfo
				}
				elementIndex++
			}
		} else if strings.HasPrefix(line, "#") {
			// Character line
			tag := sp.parseCharacterLine(line)
			if tag != nil {
				tag.Line = sp.currentLine
				scenario.Elements = append(scenario.Elements, *tag)
				elementIndex++
			}
		} else if strings.Contains(line, "[") && strings.Contains(line, "]") {
			// Line with tags (mixed content)
			tags := sp.parseTextLine(line)
			for _, tag := range tags {
				tag.Line = sp.currentLine
				scenario.Elements = append(scenario.Elements, *tag)
				elementIndex++
			}
		} else {
			// Plain text line
			if line != "" {
				tag := types.NewParsedTag("text", sp.currentLine)
				tag.Value = line
				scenario.Elements = append(scenario.Elements, *tag)
				elementIndex++
			}
		}
	}
	
	return scenario, nil
}

// parseCharacterLine parses a character line (#character:expression)
func (sp *ScenarioParser) parseCharacterLine(line string) *types.ParsedTag {
	// Remove the # prefix
	content := strings.TrimPrefix(line, "#")
	content = strings.TrimSpace(content)
	
	if content == "" {
		return nil
	}
	
	// Split by colon to separate character name and expression
	parts := strings.SplitN(content, ":", 2)
	
	characterName := strings.TrimSpace(parts[0])
	expression := ""
	
	if len(parts) > 1 {
		expression = strings.TrimSpace(parts[1])
	}
	
	// Create chara_ptext tag
	tag := types.NewParsedTag("chara_ptext", sp.currentLine)
	tag.Parameters["name"] = characterName
	tag.Parameters["face"] = expression
	
	return tag
}

// parseLabelLine parses a label line (*label|description)
func (sp *ScenarioParser) parseLabelLine(line string) (*types.ParsedTag, *types.LabelInfo) {
	// Remove the * prefix
	content := strings.TrimPrefix(line, "*")
	content = strings.TrimSpace(content)
	
	if content == "" {
		return nil, nil
	}
	
	// Split by pipe to separate label name and description
	parts := strings.SplitN(content, "|", 2)
	
	labelName := strings.TrimSpace(parts[0])
	description := ""
	
	if len(parts) > 1 {
		description = strings.TrimSpace(parts[1])
	}
	
	// Create label tag
	tag := types.NewParsedTag("label", sp.currentLine)
	tag.Value = labelName
	
	// Create label info
	labelInfo := types.NewLabelInfo(labelName, sp.currentLine, 0, description)
	
	return tag, labelInfo
}

// parseTagLine parses a tag line ([tag param=value])
func (sp *ScenarioParser) parseTagLine(line string) []*types.ParsedTag {
	var tags []*types.ParsedTag
	
	// Find all tags in the line
	i := 0
	runes := []rune(line)
	
	for i < len(runes) {
		// Find opening bracket
		for i < len(runes) && runes[i] != '[' {
			i++
		}
		
		if i >= len(runes) {
			break
		}
		
		// Find matching closing bracket
		start := i
		bracketDepth := 0
		inQuotes := false
		quoteChar := rune(0)
		
		for i < len(runes) {
			ch := runes[i]
			
			if !inQuotes {
				if ch == '[' {
					bracketDepth++
				} else if ch == ']' {
					bracketDepth--
					if bracketDepth == 0 {
						i++
						break
					}
				} else if ch == '"' || ch == '\'' {
					inQuotes = true
					quoteChar = ch
				}
			} else {
				if ch == quoteChar && (i == 0 || runes[i-1] != '\\') {
					inQuotes = false
					quoteChar = 0
				}
			}
			i++
		}
		
		// Extract tag string
		tagStr := string(runes[start:i])
		
		// Parse the tag
		if parsedTag, err := sp.tagParser.ParseTag(tagStr, sp.currentLine); err == nil {
			tags = append(tags, parsedTag)
		}
	}
	
	return tags
}

// parseTextLine parses a text line with mixed content (text and tags)
func (sp *ScenarioParser) parseTextLine(line string) []*types.ParsedTag {
	var tags []*types.ParsedTag
	
	if line == "" {
		return tags
	}
	
	runes := []rune(line)
	i := 0
	
	for i < len(runes) {
		// Find next tag or end of line
		textStart := i
		
		// Read text until we find a tag
		for i < len(runes) && runes[i] != '[' {
			i++
		}
		
		// Add text content if any
		if i > textStart {
			textContent := string(runes[textStart:i])
			if textContent != "" {
				textTag := types.NewParsedTag("text", sp.currentLine)
				textTag.Value = textContent
				tags = append(tags, textTag)
			}
		}
		
		// If we found a tag, parse it
		if i < len(runes) && runes[i] == '[' {
			tagStart := i
			bracketDepth := 0
			inQuotes := false
			quoteChar := rune(0)
			
			// Find matching closing bracket
			for i < len(runes) {
				ch := runes[i]
				
				if !inQuotes {
					if ch == '[' {
						bracketDepth++
					} else if ch == ']' {
						bracketDepth--
						if bracketDepth == 0 {
							i++
							break
						}
					} else if ch == '"' || ch == '\'' {
						inQuotes = true
						quoteChar = ch
					}
				} else {
					if ch == quoteChar && (i == 0 || runes[i-1] != '\\') {
						inQuotes = false
						quoteChar = 0
					}
				}
				i++
			}
			
			// Extract and parse tag
			tagStr := string(runes[tagStart:i])
			if parsedTag, err := sp.tagParser.ParseTag(tagStr, sp.currentLine); err == nil {
				tags = append(tags, parsedTag)
			}
		}
	}
	
	return tags
}


// parseComment checks if a line is a comment and should be ignored
func (sp *ScenarioParser) parseComment(line string) bool {
	trimmed := strings.TrimSpace(line)
	
	// Single line comment
	if strings.HasPrefix(trimmed, ";") {
		return true
	}
	
	// Block comment start
	if strings.HasPrefix(trimmed, "/*") {
		return true
	}
	
	return false
}

// parseScriptBlock checks if a line is a script block tag and returns the parsed tag
func (sp *ScenarioParser) parseScriptBlock(line string) *types.ParsedTag {
	trimmed := strings.TrimSpace(line)
	
	// Check for [iscript] tag
	if trimmed == "[iscript]" {
		tag := types.NewParsedTag("iscript", sp.currentLine)
		return tag
	}
	
	// Check for [endscript] tag
	if trimmed == "[endscript]" {
		tag := types.NewParsedTag("endscript", sp.currentLine)
		return tag
	}
	
	return nil
}
