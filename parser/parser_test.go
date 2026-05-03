package parser

import (
	"strings"
	"testing"

	"github.com/botamochi0x12/tyrano-parser-go/types"
)

func TestNewTyranoParser(t *testing.T) {
	options := ParserOptions{
		KeepSpaceInParameterValue: "true",
		StrictMode:                true,
		EnableWarnings:            false,
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
			content: ";title=\"My Game\";\n;width=1280;\n;height=720;",
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

func TestTyranoParser_StrictMode(t *testing.T) {
	strictParser := NewTyranoParser(ParserOptions{
		StrictMode:     true,
		EnableWarnings: true,
	})

	lenientParser := NewTyranoParser(ParserOptions{
		StrictMode:     false,
		EnableWarnings: true,
	})

	// Test scenario with errors
	invalidScenario := "*\n#\n[invalid_tag"

	// Strict mode should return error
	_, err := strictParser.ParseScenario(invalidScenario)
	if err == nil {
		t.Error("Strict mode should return error for invalid scenario")
	}

	// Lenient mode should not return error but collect issues
	scenario, err := lenientParser.ParseScenario(invalidScenario)
	if err != nil {
		t.Errorf("Lenient mode should not return error: %v", err)
	}
	if scenario == nil {
		t.Error("Lenient mode should return scenario even with issues")
	}

	// Test config with errors
	invalidConfig := ";title=;\n;width=invalid"

	// Strict mode should return error
	_, err = strictParser.ParseConfig(invalidConfig)
	if err == nil {
		t.Error("Strict mode should return error for invalid config")
	}

	// Lenient mode should not return error
	config, err := lenientParser.ParseConfig(invalidConfig)
	if err != nil {
		t.Errorf("Lenient mode should not return error: %v", err)
	}
	if config == nil {
		t.Error("Lenient mode should return config even with issues")
	}
}

func TestTyranoParser_WithResult_Methods(t *testing.T) {
	parser := NewDefaultTyranoParser()

	// Test ParseScenarioWithResult
	scenario, result := parser.ParseScenarioWithResult("*start\n#akane\nHello![p]\n[s]")

	if scenario == nil {
		t.Error("ParseScenarioWithResult() returned nil scenario")
	}
	if result == nil {
		t.Error("ParseScenarioWithResult() returned nil result")
	}
	if result.HasErrors() {
		t.Errorf("Valid scenario should not have errors: %s", result.Summary())
	}

	// Test ParseConfigWithResult
	config, result := parser.ParseConfigWithResult(";title=\"Test\";\n;width=800;")

	if config == nil {
		t.Error("ParseConfigWithResult() returned nil config")
	}
	if result == nil {
		t.Error("ParseConfigWithResult() returned nil result")
	}
	if result.HasErrors() {
		t.Errorf("Valid config should not have errors: %s", result.Summary())
	}
}

func TestTyranoParser_ErrorCollection(t *testing.T) {
	parser := NewDefaultTyranoParser()

	// Test scenario with multiple issues
	invalidScenario := "*\n*start\n*start\n[invalid_tag"

	scenario, result := parser.ParseScenarioWithResult(invalidScenario)

	if scenario == nil {
		t.Error("ParseScenarioWithResult() returned nil scenario")
	}
	if result == nil {
		t.Error("ParseScenarioWithResult() returned nil result")
	}

	if !result.HasErrors() {
		t.Error("Invalid scenario should have errors")
	}

	errors := result.GetErrors()
	if len(errors) == 0 {
		t.Error("Should have collected multiple errors")
	}

	// Check for specific error types
	hasEmptyLabelError := false
	hasDuplicateLabelError := false

	for _, err := range errors {
		switch err.Type {
		case types.SyntaxError:
			if strings.Contains(err.Message, "empty label") {
				hasEmptyLabelError = true
			}
		case types.DuplicateLabelError:
			hasDuplicateLabelError = true
		}
	}

	if !hasEmptyLabelError {
		t.Error("Should have detected empty label error")
	}
	if !hasDuplicateLabelError {
		t.Error("Should have detected duplicate label error")
	}
}

func TestParserOptions_StructValidation(t *testing.T) {
	options := ParserOptions{
		KeepSpaceInParameterValue: "custom",
		StrictMode:                true,
		EnableWarnings:            false,
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
