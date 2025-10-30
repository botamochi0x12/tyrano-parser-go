package parser

import (
	"testing"
	"github.com/tyranoscript/tyrano-parser-go/types"
)

func TestNewTagParser(t *testing.T) {
	tests := []struct {
		name            string
		keepSpaceConfig string
	}{
		{
			name:            "default config",
			keepSpaceConfig: "false",
		},
		{
			name:            "keep space config",
			keepSpaceConfig: "true",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parser := NewTagParser(tt.keepSpaceConfig)
			
			if parser == nil {
				t.Fatal("NewTagParser() returned nil")
			}
			
			if parser.keepSpaceConfig != tt.keepSpaceConfig {
				t.Errorf("NewTagParser().keepSpaceConfig = %v, want %v", parser.keepSpaceConfig, tt.keepSpaceConfig)
			}
		})
	}
}

func TestTagParser_ParseTag_BasicTags(t *testing.T) {
	parser := NewTagParser("false")
	
	tests := []struct {
		name     string
		tagStr   string
		lineNum  int
		expected *types.ParsedTag
		wantErr  bool
	}{
		{
			name:    "simple tag without parameters",
			tagStr:  "[cm]",
			lineNum: 1,
			expected: &types.ParsedTag{
				Name:       "cm",
				Line:       1,
				Parameters: map[string]string{},
				Value:      "",
			},
			wantErr: false,
		},
		{
			name:    "tag with single parameter",
			tagStr:  "[bg storage=\"room.jpg\"]",
			lineNum: 2,
			expected: &types.ParsedTag{
				Name:       "bg",
				Line:       2,
				Parameters: map[string]string{
					"storage": "room.jpg",
				},
				Value: "",
			},
			wantErr: false,
		},
		{
			name:    "tag with multiple parameters",
			tagStr:  "[chara_show name=\"akane\" left=\"300\" time=\"1000\"]",
			lineNum: 3,
			expected: &types.ParsedTag{
				Name:       "chara_show",
				Line:       3,
				Parameters: map[string]string{
					"name": "akane",
					"left": "300",
					"time": "1000",
				},
				Value: "",
			},
			wantErr: false,
		},
		{
			name:    "tag with unquoted parameters",
			tagStr:  "[jump storage=scene1.ks target=*start]",
			lineNum: 4,
			expected: &types.ParsedTag{
				Name:       "jump",
				Line:       4,
				Parameters: map[string]string{
					"storage": "scene1.ks",
					"target":  "*start",
				},
				Value: "",
			},
			wantErr: false,
		},
		{
			name:    "tag with mixed quoted and unquoted parameters",
			tagStr:  "[button name=\"save_btn\" graphic=\"save.png\" x=100 y=200]",
			lineNum: 5,
			expected: &types.ParsedTag{
				Name:       "button",
				Line:       5,
				Parameters: map[string]string{
					"name":    "save_btn",
					"graphic": "save.png",
					"x":       "100",
					"y":       "200",
				},
				Value: "",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := parser.ParseTag(tt.tagStr, tt.lineNum)
			
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseTag() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			
			if result == nil {
				t.Fatal("ParseTag() returned nil result")
			}
			
			if result.Name != tt.expected.Name {
				t.Errorf("ParseTag().Name = %v, want %v", result.Name, tt.expected.Name)
			}
			
			if result.Line != tt.expected.Line {
				t.Errorf("ParseTag().Line = %v, want %v", result.Line, tt.expected.Line)
			}
			
			if result.Value != tt.expected.Value {
				t.Errorf("ParseTag().Value = %v, want %v", result.Value, tt.expected.Value)
			}
			
			if len(result.Parameters) != len(tt.expected.Parameters) {
				t.Errorf("ParseTag().Parameters length = %v, want %v", len(result.Parameters), len(tt.expected.Parameters))
			}
			
			for key, expectedValue := range tt.expected.Parameters {
				if actualValue, exists := result.Parameters[key]; !exists {
					t.Errorf("ParseTag().Parameters missing key %v", key)
				} else if actualValue != expectedValue {
					t.Errorf("ParseTag().Parameters[%v] = %v, want %v", key, actualValue, expectedValue)
				}
			}
		})
	}
}

func TestTagParser_ParseTag_QuotedStrings(t *testing.T) {
	parser := NewTagParser("false")
	
	tests := []struct {
		name     string
		tagStr   string
		lineNum  int
		expected *types.ParsedTag
		wantErr  bool
	}{
		{
			name:    "double quoted string",
			tagStr:  "[text value=\"Hello world!\"]",
			lineNum: 1,
			expected: &types.ParsedTag{
				Name:       "text",
				Line:       1,
				Parameters: map[string]string{
					"value": "Hello world!",
				},
				Value: "",
			},
			wantErr: false,
		},
		{
			name:    "single quoted string",
			tagStr:  "[text value='Hello world!']",
			lineNum: 1,
			expected: &types.ParsedTag{
				Name:       "text",
				Line:       1,
				Parameters: map[string]string{
					"value": "Hello world!",
				},
				Value: "",
			},
			wantErr: false,
		},
		{
			name:    "string with escape sequences",
			tagStr:  "[text value=\"Hello\\nworld\\t!\"]",
			lineNum: 1,
			expected: &types.ParsedTag{
				Name:       "text",
				Line:       1,
				Parameters: map[string]string{
					"value": "Hello\nworld\t!",
				},
				Value: "",
			},
			wantErr: false,
		},
		{
			name:    "string with escaped quotes",
			tagStr:  "[text value=\"Say \\\"Hello\\\" to the world\"]",
			lineNum: 1,
			expected: &types.ParsedTag{
				Name:       "text",
				Line:       1,
				Parameters: map[string]string{
					"value": "Say \"Hello\" to the world",
				},
				Value: "",
			},
			wantErr: false,
		},
		{
			name:    "Japanese string",
			tagStr:  "[text value=\"こんにちは世界！\"]",
			lineNum: 1,
			expected: &types.ParsedTag{
				Name:       "text",
				Line:       1,
				Parameters: map[string]string{
					"value": "こんにちは世界！",
				},
				Value: "",
			},
			wantErr: false,
		},
		{
			name:    "empty string",
			tagStr:  "[text value=\"\"]",
			lineNum: 1,
			expected: &types.ParsedTag{
				Name:       "text",
				Line:       1,
				Parameters: map[string]string{
					"value": "",
				},
				Value: "",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := parser.ParseTag(tt.tagStr, tt.lineNum)
			
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseTag() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			
			if result == nil {
				t.Fatal("ParseTag() returned nil result")
			}
			
			if result.Name != tt.expected.Name {
				t.Errorf("ParseTag().Name = %v, want %v", result.Name, tt.expected.Name)
			}
			
			for key, expectedValue := range tt.expected.Parameters {
				if actualValue, exists := result.Parameters[key]; !exists {
					t.Errorf("ParseTag().Parameters missing key %v", key)
				} else if actualValue != expectedValue {
					t.Errorf("ParseTag().Parameters[%v] = %v, want %v", key, actualValue, expectedValue)
				}
			}
		})
	}
}

func TestTagParser_ParseTag_NestedBrackets(t *testing.T) {
	parser := NewTagParser("false")
	
	tests := []struct {
		name     string
		tagStr   string
		lineNum  int
		expected *types.ParsedTag
		wantErr  bool
	}{
		{
			name:    "nested brackets in parameter value",
			tagStr:  "[eval exp=\"f.test = [1,2,3]\"]",
			lineNum: 1,
			expected: &types.ParsedTag{
				Name:       "eval",
				Line:       1,
				Parameters: map[string]string{
					"exp": "f.test = [1,2,3]",
				},
				Value: "",
			},
			wantErr: false,
		},
		{
			name:    "complex nested brackets",
			tagStr:  "[eval exp=\"f.data = {items: [1, 2, {nested: [3, 4]}]}\"]",
			lineNum: 1,
			expected: &types.ParsedTag{
				Name:       "eval",
				Line:       1,
				Parameters: map[string]string{
					"exp": "f.data = {items: [1, 2, {nested: [3, 4]}]}",
				},
				Value: "",
			},
			wantErr: false,
		},
		{
			name:    "brackets in multiple parameters",
			tagStr:  "[eval exp1=\"arr1 = [1,2]\" exp2=\"arr2 = [3,4]\"]",
			lineNum: 1,
			expected: &types.ParsedTag{
				Name:       "eval",
				Line:       1,
				Parameters: map[string]string{
					"exp1": "arr1 = [1,2]",
					"exp2": "arr2 = [3,4]",
				},
				Value: "",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := parser.ParseTag(tt.tagStr, tt.lineNum)
			
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseTag() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			
			if result == nil {
				t.Fatal("ParseTag() returned nil result")
			}
			
			if result.Name != tt.expected.Name {
				t.Errorf("ParseTag().Name = %v, want %v", result.Name, tt.expected.Name)
			}
			
			for key, expectedValue := range tt.expected.Parameters {
				if actualValue, exists := result.Parameters[key]; !exists {
					t.Errorf("ParseTag().Parameters missing key %v", key)
				} else if actualValue != expectedValue {
					t.Errorf("ParseTag().Parameters[%v] = %v, want %v", key, actualValue, expectedValue)
				}
			}
		})
	}
}

func TestTagParser_ParseTag_KeepSpaceInParameterValue(t *testing.T) {
	tests := []struct {
		name            string
		keepSpaceConfig string
		tagStr          string
		expected        string
	}{
		{
			name:            "keep space disabled",
			keepSpaceConfig: "false",
			tagStr:          "[text value=\" hello world \"]",
			expected:        "hello world",
		},
		{
			name:            "keep space enabled",
			keepSpaceConfig: "true",
			tagStr:          "[text value=\" hello world \"]",
			expected:        " hello world ",
		},
		{
			name:            "keep space with tabs",
			keepSpaceConfig: "true",
			tagStr:          "[text value=\"\thello\tworld\t\"]",
			expected:        "\thello\tworld\t",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parser := NewTagParser(tt.keepSpaceConfig)
			result, err := parser.ParseTag(tt.tagStr, 1)
			
			if err != nil {
				t.Errorf("ParseTag() error = %v", err)
				return
			}
			
			if result == nil {
				t.Fatal("ParseTag() returned nil result")
			}
			
			if actualValue, exists := result.Parameters["value"]; !exists {
				t.Error("ParseTag().Parameters missing 'value' key")
			} else if actualValue != tt.expected {
				t.Errorf("ParseTag().Parameters[value] = %q, want %q", actualValue, tt.expected)
			}
		})
	}
}

func TestTagParser_ParseTag_MalformedTags(t *testing.T) {
	parser := NewTagParser("false")
	
	tests := []struct {
		name    string
		tagStr  string
		lineNum int
		wantErr bool
	}{
		{
			name:    "missing opening bracket",
			tagStr:  "cm]",
			lineNum: 1,
			wantErr: true,
		},
		{
			name:    "missing closing bracket",
			tagStr:  "[cm",
			lineNum: 1,
			wantErr: true,
		},
		{
			name:    "empty tag",
			tagStr:  "[]",
			lineNum: 1,
			wantErr: true,
		},
		{
			name:    "unclosed quoted parameter",
			tagStr:  "[text value=\"hello world]",
			lineNum: 1,
			wantErr: true,
		},
		{
			name:    "invalid parameter format",
			tagStr:  "[text value=hello=world]",
			lineNum: 1,
			wantErr: true,
		},
		{
			name:    "parameter without value",
			tagStr:  "[text value=]",
			lineNum: 1,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := parser.ParseTag(tt.tagStr, tt.lineNum)
			
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseTag() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			
			if !tt.wantErr && result == nil {
				t.Error("ParseTag() returned nil result when expecting success")
			}
		})
	}
}

func TestTagParser_extractParameters(t *testing.T) {
	parser := NewTagParser("false")
	
	tests := []struct {
		name     string
		paramStr string
		expected map[string]string
		wantErr  bool
	}{
		{
			name:     "empty parameters",
			paramStr: "",
			expected: map[string]string{},
			wantErr:  false,
		},
		{
			name:     "single parameter",
			paramStr: "storage=\"room.jpg\"",
			expected: map[string]string{
				"storage": "room.jpg",
			},
			wantErr: false,
		},
		{
			name:     "multiple parameters",
			paramStr: "name=\"akane\" left=\"300\" time=\"1000\"",
			expected: map[string]string{
				"name": "akane",
				"left": "300",
				"time": "1000",
			},
			wantErr: false,
		},
		{
			name:     "unquoted parameters",
			paramStr: "storage=scene1.ks target=*start",
			expected: map[string]string{
				"storage": "scene1.ks",
				"target":  "*start",
			},
			wantErr: false,
		},
		{
			name:     "mixed quoted and unquoted",
			paramStr: "name=\"save_btn\" x=100 y=200",
			expected: map[string]string{
				"name": "save_btn",
				"x":    "100",
				"y":    "200",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := parser.extractParameters(tt.paramStr)
			
			if (err != nil) != tt.wantErr {
				t.Errorf("extractParameters() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			
			if len(result) != len(tt.expected) {
				t.Errorf("extractParameters() length = %v, want %v", len(result), len(tt.expected))
			}
			
			for key, expectedValue := range tt.expected {
				if actualValue, exists := result[key]; !exists {
					t.Errorf("extractParameters() missing key %v", key)
				} else if actualValue != expectedValue {
					t.Errorf("extractParameters()[%v] = %v, want %v", key, actualValue, expectedValue)
				}
			}
		})
	}
}

func TestTagParser_handleQuotedValue(t *testing.T) {
	tests := []struct {
		name            string
		keepSpaceConfig string
		value           string
		quote           rune
		expected        string
	}{
		{
			name:            "double quoted value",
			keepSpaceConfig: "false",
			value:           "hello world",
			quote:           '"',
			expected:        "hello world",
		},
		{
			name:            "single quoted value",
			keepSpaceConfig: "false",
			value:           "hello world",
			quote:           '\'',
			expected:        "hello world",
		},
		{
			name:            "value with leading/trailing spaces - keep space disabled",
			keepSpaceConfig: "false",
			value:           " hello world ",
			quote:           '"',
			expected:        "hello world",
		},
		{
			name:            "value with leading/trailing spaces - keep space enabled",
			keepSpaceConfig: "true",
			value:           " hello world ",
			quote:           '"',
			expected:        " hello world ",
		},
		{
			name:            "value with escape sequences",
			keepSpaceConfig: "false",
			value:           "hello\\nworld\\t!",
			quote:           '"',
			expected:        "hello\nworld\t!",
		},
		{
			name:            "value with escaped quotes",
			keepSpaceConfig: "false",
			value:           "Say \\\"Hello\\\" world",
			quote:           '"',
			expected:        "Say \"Hello\" world",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parser := NewTagParser(tt.keepSpaceConfig)
			result := parser.handleQuotedValue(tt.value, tt.quote)
			
			if result != tt.expected {
				t.Errorf("handleQuotedValue() = %q, want %q", result, tt.expected)
			}
		})
	}
}
