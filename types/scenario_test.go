package types

import (
	"testing"
)

func TestNewParsedScenario(t *testing.T) {
	scenario := NewParsedScenario()
	
	if scenario == nil {
		t.Fatal("NewParsedScenario() returned nil")
	}
	
	if scenario.Elements == nil {
		t.Error("NewParsedScenario().Elements is nil, expected initialized slice")
	}
	
	if len(scenario.Elements) != 0 {
		t.Errorf("NewParsedScenario().Elements length = %v, want 0", len(scenario.Elements))
	}
	
	if scenario.Labels == nil {
		t.Error("NewParsedScenario().Labels is nil, expected initialized map")
	}
	
	if len(scenario.Labels) != 0 {
		t.Errorf("NewParsedScenario().Labels length = %v, want 0", len(scenario.Labels))
	}
}

func TestNewParsedTag(t *testing.T) {
	tests := []struct {
		name     string
		tagName  string
		line     int
	}{
		{
			name:    "basic tag creation",
			tagName: "cm",
			line:    5,
		},
		{
			name:    "character tag creation",
			tagName: "chara_ptext",
			line:    10,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tag := NewParsedTag(tt.tagName, tt.line)
			
			if tag == nil {
				t.Fatal("NewParsedTag() returned nil")
			}
			
			if tag.Name != tt.tagName {
				t.Errorf("NewParsedTag().Name = %v, want %v", tag.Name, tt.tagName)
			}
			
			if tag.Line != tt.line {
				t.Errorf("NewParsedTag().Line = %v, want %v", tag.Line, tt.line)
			}
			
			if tag.Parameters == nil {
				t.Error("NewParsedTag().Parameters is nil, expected initialized map")
			}
			
			if len(tag.Parameters) != 0 {
				t.Errorf("NewParsedTag().Parameters length = %v, want 0", len(tag.Parameters))
			}
			
			if tag.Value != "" {
				t.Errorf("NewParsedTag().Value = %v, want empty string", tag.Value)
			}
			
			if tag.IsEntityDisabled != false {
				t.Errorf("NewParsedTag().IsEntityDisabled = %v, want false", tag.IsEntityDisabled)
			}
		})
	}
}

func TestNewLabelInfo(t *testing.T) {
	tests := []struct {
		name      string
		labelName string
		line      int
		index     int
		value     string
	}{
		{
			name:      "basic label",
			labelName: "start",
			line:      1,
			index:     0,
			value:     "",
		},
		{
			name:      "label with description",
			labelName: "scene1",
			line:      15,
			index:     5,
			value:     "First scene",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			label := NewLabelInfo(tt.labelName, tt.line, tt.index, tt.value)
			
			if label == nil {
				t.Fatal("NewLabelInfo() returned nil")
			}
			
			if label.LabelName != tt.labelName {
				t.Errorf("NewLabelInfo().LabelName = %v, want %v", label.LabelName, tt.labelName)
			}
			
			if label.Line != tt.line {
				t.Errorf("NewLabelInfo().Line = %v, want %v", label.Line, tt.line)
			}
			
			if label.Index != tt.index {
				t.Errorf("NewLabelInfo().Index = %v, want %v", label.Index, tt.index)
			}
			
			if label.Value != tt.value {
				t.Errorf("NewLabelInfo().Value = %v, want %v", label.Value, tt.value)
			}
		})
	}
}

func TestParsedScenario_StructValidation(t *testing.T) {
	scenario := &ParsedScenario{
		Elements: []ParsedTag{
			{
				Line:       1,
				Name:       "cm",
				Parameters: map[string]string{},
				Value:      "",
			},
		},
		Labels: map[string]*LabelInfo{
			"start": {
				Line:      1,
				Index:     0,
				LabelName: "start",
				Value:     "",
			},
		},
	}
	
	if len(scenario.Elements) != 1 {
		t.Errorf("ParsedScenario.Elements length = %v, want 1", len(scenario.Elements))
	}
	
	if len(scenario.Labels) != 1 {
		t.Errorf("ParsedScenario.Labels length = %v, want 1", len(scenario.Labels))
	}
	
	if scenario.Labels["start"] == nil {
		t.Error("ParsedScenario.Labels['start'] is nil")
	}
}
