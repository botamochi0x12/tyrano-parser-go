package parser

import (
	"testing"
)

func TestNewTyranoParser(t *testing.T) {
	options := ParserOptions{
		KeepSpaceInParameterValue: "true",
		StrictMode:               true,
		EnableWarnings:          false,
	}
	
	parser := NewTyranoParser(options)
	
	if parser == nil {
		t.Fatal("NewTyranoParser() returned nil")
	}
	
	if parser.options.KeepSpaceInParameterValue != "true" {
		t.Errorf("NewTyranoParser().options.KeepSpaceInParameterValue = %v, want true", parser.options.KeepSpaceInParameterValue)
	}
	
	if parser.options.StrictMode != true {
		t.Errorf("NewTyranoParser().options.StrictMode = %v, want true", parser.options.StrictMode)
	}
	
	if parser.options.EnableWarnings != false {
		t.Errorf("NewTyranoParser().options.EnableWarnings = %v, want false", parser.options.EnableWarnings)
	}
}

func TestNewDefaultTyranoParser(t *testing.T) {
	parser := NewDefaultTyranoParser()
	
	if parser == nil {
		t.Fatal("NewDefaultTyranoParser() returned nil")
	}
	
	if parser.options.KeepSpaceInParameterValue != "false" {
		t.Errorf("NewDefaultTyranoParser().options.KeepSpaceInParameterValue = %v, want false", parser.options.KeepSpaceInParameterValue)
	}
	
	if parser.options.StrictMode != false {
		t.Errorf("NewDefaultTyranoParser().options.StrictMode = %v, want false", parser.options.StrictMode)
	}
	
	if parser.options.EnableWarnings != true {
		t.Errorf("NewDefaultTyranoParser().options.EnableWarnings = %v, want true", parser.options.EnableWarnings)
	}
}

func TestTyranoParser_InterfaceCompliance(t *testing.T) {
	// Test that TyranoParser implements the Parser interface
	var _ Parser = (*TyranoParser)(nil)
	
	parser := NewDefaultTyranoParser()
	
	// Test ParseScenario method exists and returns expected types
	scenario, err := parser.ParseScenario("")
	if err != nil {
		t.Errorf("ParseScenario() returned error: %v", err)
	}
	if scenario == nil {
		t.Error("ParseScenario() returned nil scenario")
	}
	
	// Test ParseConfig method exists and returns expected types
	config, err := parser.ParseConfig("")
	if err != nil {
		t.Errorf("ParseConfig() returned error: %v", err)
	}
	if config == nil {
		t.Error("ParseConfig() returned nil config")
	}
}

func TestTyranoParser_ParseScenario_BasicFunctionality(t *testing.T) {
	parser := NewDefaultTyranoParser()
	
	tests := []struct {
		name    string
		content string
		wantErr bool
	}{
		{
			name:    "empty content",
			content: "",
			wantErr: false,
		},
		{
			name:    "basic scenario content",
			content: "*start\n[cm]\n#akane\nHello![p]\n[s]",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			scenario, err := parser.ParseScenario(tt.content)
			
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseScenario() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			
			if scenario == nil {
				t.Error("ParseScenario() returned nil scenario")
				return
			}
			
			// Verify basic structure
			if scenario.Elements == nil {
				t.Error("ParseScenario() returned scenario with nil Elements")
			}
			
			if scenario.Labels == nil {
				t.Error("ParseScenario() returned scenario with nil Labels")
			}
		})
	}
}

func TestTyranoParser_ParseConfig_BasicFunctionality(t *testing.T) {
	parser := NewDefaultTyranoParser()
	
	tests := []struct {
		name    string
		content string
		wantErr bool
	}{
		{
			name:    "empty content",
			content: "",
			wantErr: false,
		},
		{
			name:    "basic config content",
			content: "title=\"My Game\";\nwidth=1280;\nheight=720;",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config, err := parser.ParseConfig(tt.content)
			
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseConfig() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			
			if config == nil {
				t.Error("ParseConfig() returned nil config")
				return
			}
		})
	}
}

func TestParserOptions_StructValidation(t *testing.T) {
	options := ParserOptions{
		KeepSpaceInParameterValue: "custom",
		StrictMode:               true,
		EnableWarnings:          false,
	}
	
	if options.KeepSpaceInParameterValue != "custom" {
		t.Errorf("ParserOptions.KeepSpaceInParameterValue = %v, want custom", options.KeepSpaceInParameterValue)
	}
	
	if options.StrictMode != true {
		t.Errorf("ParserOptions.StrictMode = %v, want true", options.StrictMode)
	}
	
	if options.EnableWarnings != false {
		t.Errorf("ParserOptions.EnableWarnings = %v, want false", options.EnableWarnings)
	}
}
