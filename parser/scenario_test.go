package parser

import (
	"testing"

	"github.com/tyranoscript/tyrano-parser-go/types"
)

func TestScenarioParser_parseCharacterLine(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected *types.ParsedTag
		wantErr  bool
	}{
		{
			name:  "simple character",
			input: "#akane",
			expected: &types.ParsedTag{
				Name: "chara_ptext",
				Parameters: map[string]string{
					"name": "akane",
					"face": "",
				},
				Value: "",
			},
		},
		{
			name:  "character with expression",
			input: "#yamato:happy",
			expected: &types.ParsedTag{
				Name: "chara_ptext",
				Parameters: map[string]string{
					"name": "yamato",
					"face": "happy",
				},
				Value: "",
			},
		},
		{
			name:  "japanese character name",
			input: "#あかね:怒り",
			expected: &types.ParsedTag{
				Name: "chara_ptext",
				Parameters: map[string]string{
					"name": "あかね",
					"face": "怒り",
				},
				Value: "",
			},
		},
		{
			name:  "character with empty expression",
			input: "#akane:",
			expected: &types.ParsedTag{
				Name: "chara_ptext",
				Parameters: map[string]string{
					"name": "akane",
					"face": "",
				},
				Value: "",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sp := NewScenarioParser()
			sp.currentLine = 1
			
			result := sp.parseCharacterLine(tt.input)
			
			if result == nil {
				t.Fatal("parseCharacterLine() returned nil")
			}
			
			if result.Name != tt.expected.Name {
				t.Errorf("parseCharacterLine().Name = %v, want %v", result.Name, tt.expected.Name)
			}
			
			if result.Parameters["name"] != tt.expected.Parameters["name"] {
				t.Errorf("parseCharacterLine().Parameters[name] = %v, want %v", result.Parameters["name"], tt.expected.Parameters["name"])
			}
			
			if result.Parameters["face"] != tt.expected.Parameters["face"] {
				t.Errorf("parseCharacterLine().Parameters[face] = %v, want %v", result.Parameters["face"], tt.expected.Parameters["face"])
			}
		})
	}
}

func TestScenarioParser_parseLabelLine(t *testing.T) {
	tests := []struct {
		name          string
		input         string
		expectedTag   *types.ParsedTag
		expectedLabel *types.LabelInfo
	}{
		{
			name:  "simple label",
			input: "*start",
			expectedTag: &types.ParsedTag{
				Name:       "label",
				Parameters: map[string]string{},
				Value:      "start",
			},
			expectedLabel: &types.LabelInfo{
				LabelName: "start",
				Value:     "",
			},
		},
		{
			name:  "label with description",
			input: "*scene1|First scene begins",
			expectedTag: &types.ParsedTag{
				Name:       "label",
				Parameters: map[string]string{},
				Value:      "scene1",
			},
			expectedLabel: &types.LabelInfo{
				LabelName: "scene1",
				Value:     "First scene begins",
			},
		},
		{
			name:  "japanese label with description",
			input: "*開始|ゲーム開始",
			expectedTag: &types.ParsedTag{
				Name:       "label",
				Parameters: map[string]string{},
				Value:      "開始",
			},
			expectedLabel: &types.LabelInfo{
				LabelName: "開始",
				Value:     "ゲーム開始",
			},
		},
		{
			name:  "label with empty description",
			input: "*start|",
			expectedTag: &types.ParsedTag{
				Name:       "label",
				Parameters: map[string]string{},
				Value:      "start",
			},
			expectedLabel: &types.LabelInfo{
				LabelName: "start",
				Value:     "",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sp := NewScenarioParser()
			sp.currentLine = 1
			
			tag, label := sp.parseLabelLine(tt.input)
			
			if tag == nil {
				t.Fatal("parseLabelLine() returned nil tag")
			}
			
			if label == nil {
				t.Fatal("parseLabelLine() returned nil label")
			}
			
			if tag.Name != tt.expectedTag.Name {
				t.Errorf("parseLabelLine().tag.Name = %v, want %v", tag.Name, tt.expectedTag.Name)
			}
			
			if tag.Value != tt.expectedTag.Value {
				t.Errorf("parseLabelLine().tag.Value = %v, want %v", tag.Value, tt.expectedTag.Value)
			}
			
			if label.LabelName != tt.expectedLabel.LabelName {
				t.Errorf("parseLabelLine().label.LabelName = %v, want %v", label.LabelName, tt.expectedLabel.LabelName)
			}
			
			if label.Value != tt.expectedLabel.Value {
				t.Errorf("parseLabelLine().label.Value = %v, want %v", label.Value, tt.expectedLabel.Value)
			}
		})
	}
}

func TestScenarioParser_parseTextLine(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []*types.ParsedTag
	}{
		{
			name:  "simple text",
			input: "Hello world",
			expected: []*types.ParsedTag{
				{
					Name:       "text",
					Parameters: map[string]string{},
					Value:      "Hello world",
				},
			},
		},
		{
			name:  "japanese text",
			input: "こんにちは世界",
			expected: []*types.ParsedTag{
				{
					Name:       "text",
					Parameters: map[string]string{},
					Value:      "こんにちは世界",
				},
			},
		},
		{
			name:  "text with mixed content",
			input: "Hello [p] world",
			expected: []*types.ParsedTag{
				{
					Name:       "text",
					Parameters: map[string]string{},
					Value:      "Hello ",
				},
				{
					Name:       "p",
					Parameters: map[string]string{},
					Value:      "",
				},
				{
					Name:       "text",
					Parameters: map[string]string{},
					Value:      " world",
				},
			},
		},
		{
			name:  "empty text",
			input: "",
			expected: []*types.ParsedTag{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sp := NewScenarioParser()
			sp.currentLine = 1
			
			result := sp.parseTextLine(tt.input)
			
			if len(result) != len(tt.expected) {
				t.Errorf("parseTextLine() returned %d tags, want %d", len(result), len(tt.expected))
				return
			}
			
			for i, tag := range result {
				if tag.Name != tt.expected[i].Name {
					t.Errorf("parseTextLine()[%d].Name = %v, want %v", i, tag.Name, tt.expected[i].Name)
				}
				
				if tag.Value != tt.expected[i].Value {
					t.Errorf("parseTextLine()[%d].Value = %v, want %v", i, tag.Value, tt.expected[i].Value)
				}
			}
		})
	}
}

func TestScenarioParser_Parse(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected *types.ParsedScenario
		wantErr  bool
	}{
		{
			name: "simple scenario",
			input: `*start
#akane
Hello world[p]
[cm]`,
			expected: &types.ParsedScenario{
				Elements: []types.ParsedTag{
					{
						Name:       "label",
						Parameters: map[string]string{},
						Value:      "start",
						Line:       1,
					},
					{
						Name: "chara_ptext",
						Parameters: map[string]string{
							"name": "akane",
							"face": "",
						},
						Value: "",
						Line:  2,
					},
					{
						Name:       "text",
						Parameters: map[string]string{},
						Value:      "Hello world",
						Line:       3,
					},
					{
						Name:       "p",
						Parameters: map[string]string{},
						Value:      "",
						Line:       3,
					},
					{
						Name:       "cm",
						Parameters: map[string]string{},
						Value:      "",
						Line:       4,
					},
				},
				Labels: map[string]*types.LabelInfo{
					"start": {
						LabelName: "start",
						Line:      1,
						Index:     0,
						Value:     "",
					},
				},
			},
		},
		{
			name: "scenario with comments",
			input: `;This is a comment
*start
;Another comment
Hello world`,
			expected: &types.ParsedScenario{
				Elements: []types.ParsedTag{
					{
						Name:       "label",
						Parameters: map[string]string{},
						Value:      "start",
						Line:       2,
					},
					{
						Name:       "text",
						Parameters: map[string]string{},
						Value:      "Hello world",
						Line:       4,
					},
				},
				Labels: map[string]*types.LabelInfo{
					"start": {
						LabelName: "start",
						Line:      2,
						Index:     0,
						Value:     "",
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sp := NewScenarioParser()
			
			result, err := sp.Parse(tt.input)
			
			if (err != nil) != tt.wantErr {
				t.Errorf("Parse() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			
			if result == nil {
				t.Fatal("Parse() returned nil")
			}
			
			if len(result.Elements) != len(tt.expected.Elements) {
				t.Errorf("Parse().Elements length = %d, want %d", len(result.Elements), len(tt.expected.Elements))
				return
			}
			
			for i, element := range result.Elements {
				if element.Name != tt.expected.Elements[i].Name {
					t.Errorf("Parse().Elements[%d].Name = %v, want %v", i, element.Name, tt.expected.Elements[i].Name)
				}
				
				if element.Line != tt.expected.Elements[i].Line {
					t.Errorf("Parse().Elements[%d].Line = %v, want %v", i, element.Line, tt.expected.Elements[i].Line)
				}
				
				if element.Value != tt.expected.Elements[i].Value {
					t.Errorf("Parse().Elements[%d].Value = %v, want %v", i, element.Value, tt.expected.Elements[i].Value)
				}
			}
			
			if len(result.Labels) != len(tt.expected.Labels) {
				t.Errorf("Parse().Labels length = %d, want %d", len(result.Labels), len(tt.expected.Labels))
			}
			
			for labelName, labelInfo := range tt.expected.Labels {
				if result.Labels[labelName] == nil {
					t.Errorf("Parse().Labels[%s] is nil", labelName)
					continue
				}
				
				if result.Labels[labelName].LabelName != labelInfo.LabelName {
					t.Errorf("Parse().Labels[%s].LabelName = %v, want %v", labelName, result.Labels[labelName].LabelName, labelInfo.LabelName)
				}
				
				if result.Labels[labelName].Line != labelInfo.Line {
					t.Errorf("Parse().Labels[%s].Line = %v, want %v", labelName, result.Labels[labelName].Line, labelInfo.Line)
				}
			}
		})
	}
}

func TestScenarioParser_parseComment(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected bool // true if comment should be ignored
	}{
		{
			name:     "single line comment",
			input:    ";This is a comment",
			expected: true,
		},
		{
			name:     "japanese comment",
			input:    ";これはコメントです",
			expected: true,
		},
		{
			name:     "block comment start",
			input:    "/* This is a block comment",
			expected: true,
		},
		{
			name:     "not a comment",
			input:    "Hello world",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sp := NewScenarioParser()
			
			isComment := sp.parseComment(tt.input)
			
			if isComment != tt.expected {
				t.Errorf("parseComment() = %v, want %v", isComment, tt.expected)
			}
		})
	}
}

func TestScenarioParser_parseScriptBlock(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected *types.ParsedTag
		wantErr  bool
	}{
		{
			name:  "iscript block start",
			input: "[iscript]",
			expected: &types.ParsedTag{
				Name:       "iscript",
				Parameters: map[string]string{},
				Value:      "",
			},
		},
		{
			name:  "endscript block",
			input: "[endscript]",
			expected: &types.ParsedTag{
				Name:       "endscript",
				Parameters: map[string]string{},
				Value:      "",
			},
		},
		{
			name:     "not a script block",
			input:    "regular text",
			expected: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sp := NewScenarioParser()
			sp.currentLine = 1
			
			result := sp.parseScriptBlock(tt.input)
			
			if tt.expected == nil {
				if result != nil {
					t.Errorf("parseScriptBlock() = %v, want nil", result)
				}
				return
			}
			
			if result == nil {
				t.Fatal("parseScriptBlock() returned nil")
			}
			
			if result.Name != tt.expected.Name {
				t.Errorf("parseScriptBlock().Name = %v, want %v", result.Name, tt.expected.Name)
			}
		})
	}
}

func TestScenarioParser_Parse_WithComments(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected *types.ParsedScenario
	}{
		{
			name: "scenario with single line comments",
			input: `;Header comment
*start
;Character comment
#akane
;Text comment
Hello world
;End comment`,
			expected: &types.ParsedScenario{
				Elements: []types.ParsedTag{
					{
						Name:       "label",
						Parameters: map[string]string{},
						Value:      "start",
						Line:       2,
					},
					{
						Name: "chara_ptext",
						Parameters: map[string]string{
							"name": "akane",
							"face": "",
						},
						Value: "",
						Line:  4,
					},
					{
						Name:       "text",
						Parameters: map[string]string{},
						Value:      "Hello world",
						Line:       6,
					},
				},
				Labels: map[string]*types.LabelInfo{
					"start": {
						LabelName: "start",
						Line:      2,
						Index:     0,
						Value:     "",
					},
				},
			},
		},
		{
			name: "scenario with block comments",
			input: `/* Block comment
   spanning multiple lines */
*start
/* Another block comment */
Hello world`,
			expected: &types.ParsedScenario{
				Elements: []types.ParsedTag{
					{
						Name:       "label",
						Parameters: map[string]string{},
						Value:      "start",
						Line:       3,
					},
					{
						Name:       "text",
						Parameters: map[string]string{},
						Value:      "Hello world",
						Line:       5,
					},
				},
				Labels: map[string]*types.LabelInfo{
					"start": {
						LabelName: "start",
						Line:      3,
						Index:     0,
						Value:     "",
					},
				},
			},
		},
		{
			name: "scenario with iscript blocks",
			input: `*start
[iscript]
var x = 10;
console.log(x);
[endscript]
Hello world`,
			expected: &types.ParsedScenario{
				Elements: []types.ParsedTag{
					{
						Name:       "label",
						Parameters: map[string]string{},
						Value:      "start",
						Line:       1,
					},
					{
						Name:       "iscript",
						Parameters: map[string]string{},
						Value:      "",
						Line:       2,
					},
					{
						Name:       "script_content",
						Parameters: map[string]string{},
						Value:      "var x = 10;\nconsole.log(x);",
						Line:       3,
					},
					{
						Name:       "endscript",
						Parameters: map[string]string{},
						Value:      "",
						Line:       5,
					},
					{
						Name:       "text",
						Parameters: map[string]string{},
						Value:      "Hello world",
						Line:       6,
					},
				},
				Labels: map[string]*types.LabelInfo{
					"start": {
						LabelName: "start",
						Line:      1,
						Index:     0,
						Value:     "",
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sp := NewScenarioParser()
			
			result, err := sp.Parse(tt.input)
			
			if err != nil {
				t.Errorf("Parse() error = %v", err)
				return
			}
			
			if result == nil {
				t.Fatal("Parse() returned nil")
			}
			
			if len(result.Elements) != len(tt.expected.Elements) {
				t.Errorf("Parse().Elements length = %d, want %d", len(result.Elements), len(tt.expected.Elements))
				return
			}
			
			for i, element := range result.Elements {
				if element.Name != tt.expected.Elements[i].Name {
					t.Errorf("Parse().Elements[%d].Name = %v, want %v", i, element.Name, tt.expected.Elements[i].Name)
				}
				
				if element.Line != tt.expected.Elements[i].Line {
					t.Errorf("Parse().Elements[%d].Line = %v, want %v", i, element.Line, tt.expected.Elements[i].Line)
				}
				
				if element.Value != tt.expected.Elements[i].Value {
					t.Errorf("Parse().Elements[%d].Value = %v, want %v", i, element.Value, tt.expected.Elements[i].Value)
				}
			}
		})
	}
}
