package types

import (
	"testing"
)

func TestErrorType_String(t *testing.T) {
	tests := []struct {
		name     string
		errorType ErrorType
		expected string
	}{
		{
			name:     "SyntaxError",
			errorType: SyntaxError,
			expected: "SyntaxError",
		},
		{
			name:     "DuplicateLabelError",
			errorType: DuplicateLabelError,
			expected: "DuplicateLabelError",
		},
		{
			name:     "UnmatchedIfError",
			errorType: UnmatchedIfError,
			expected: "UnmatchedIfError",
		},
		{
			name:     "InvalidConfigError",
			errorType: InvalidConfigError,
			expected: "InvalidConfigError",
		},
		{
			name:     "UnknownError",
			errorType: ErrorType(999),
			expected: "UnknownError",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.errorType.String()
			if result != tt.expected {
				t.Errorf("ErrorType.String() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestNewParseError(t *testing.T) {
	tests := []struct {
		name      string
		errorType ErrorType
		line      int
		column    int
		message   string
	}{
		{
			name:      "basic syntax error",
			errorType: SyntaxError,
			line:      10,
			column:    5,
			message:   "Invalid tag syntax",
		},
		{
			name:      "duplicate label error",
			errorType: DuplicateLabelError,
			line:      25,
			column:    1,
			message:   "Label 'start' already defined",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := NewParseError(tt.errorType, tt.line, tt.column, tt.message)
			
			if err.Type != tt.errorType {
				t.Errorf("NewParseError().Type = %v, want %v", err.Type, tt.errorType)
			}
			if err.Line != tt.line {
				t.Errorf("NewParseError().Line = %v, want %v", err.Line, tt.line)
			}
			if err.Column != tt.column {
				t.Errorf("NewParseError().Column = %v, want %v", err.Column, tt.column)
			}
			if err.Message != tt.message {
				t.Errorf("NewParseError().Message = %v, want %v", err.Message, tt.message)
			}
		})
	}
}

func TestParseError_Error(t *testing.T) {
	tests := []struct {
		name     string
		parseErr *ParseError
		expected string
	}{
		{
			name: "syntax error formatting",
			parseErr: &ParseError{
				Type:    SyntaxError,
				Line:    10,
				Column:  5,
				Message: "Invalid tag syntax",
			},
			expected: "SyntaxError at line 10, column 5: Invalid tag syntax",
		},
		{
			name: "duplicate label error formatting",
			parseErr: &ParseError{
				Type:    DuplicateLabelError,
				Line:    25,
				Column:  1,
				Message: "Label 'start' already defined",
			},
			expected: "DuplicateLabelError at line 25, column 1: Label 'start' already defined",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.parseErr.Error()
			if result != tt.expected {
				t.Errorf("ParseError.Error() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestParseIssue_Error(t *testing.T) {
	tests := []struct {
		name     string
		issue    *ParseIssue
		expected string
	}{
		{
			name: "error issue formatting",
			issue: &ParseIssue{
				Type:     SyntaxError,
				Severity: Error,
				Line:     10,
				Column:   5,
				Message:  "Invalid tag syntax",
			},
			expected: "error SyntaxError at line 10, column 5: Invalid tag syntax",
		},
		{
			name: "warning issue formatting",
			issue: &ParseIssue{
				Type:     InvalidConfigError,
				Severity: Warning,
				Line:     15,
				Column:   3,
				Message:  "Missing semicolon",
			},
			expected: "warning InvalidConfigError at line 15, column 3: Missing semicolon",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.issue.Error()
			if result != tt.expected {
				t.Errorf("ParseIssue.Error() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestParseResult_ErrorHandling(t *testing.T) {
	result := NewParseResult()
	
	// Test empty result
	if result.HasErrors() {
		t.Error("NewParseResult() should not have errors initially")
	}
	if result.HasWarnings() {
		t.Error("NewParseResult() should not have warnings initially")
	}
	if result.HasIssues() {
		t.Error("NewParseResult() should not have issues initially")
	}
	
	// Add an error
	result.AddError(SyntaxError, 10, 5, "Test error")
	
	if !result.HasErrors() {
		t.Error("ParseResult should have errors after adding error")
	}
	if result.HasWarnings() {
		t.Error("ParseResult should not have warnings after adding only error")
	}
	if !result.HasIssues() {
		t.Error("ParseResult should have issues after adding error")
	}
	
	// Add a warning
	result.AddWarning(InvalidConfigError, 15, 3, "Test warning")
	
	if !result.HasErrors() {
		t.Error("ParseResult should still have errors")
	}
	if !result.HasWarnings() {
		t.Error("ParseResult should have warnings after adding warning")
	}
	if !result.HasIssues() {
		t.Error("ParseResult should have issues after adding warning")
	}
	
	// Test counts
	if result.GetErrorCount() != 1 {
		t.Errorf("ParseResult.GetErrorCount() = %d, want 1", result.GetErrorCount())
	}
	if result.GetWarningCount() != 1 {
		t.Errorf("ParseResult.GetWarningCount() = %d, want 1", result.GetWarningCount())
	}
	if result.GetIssueCount() != 2 {
		t.Errorf("ParseResult.GetIssueCount() = %d, want 2", result.GetIssueCount())
	}
}

func TestParseResult_WithContext(t *testing.T) {
	result := NewParseResult()
	
	result.AddErrorWithContext(SyntaxError, 10, 5, "Test error", "*invalid_label")
	
	errors := result.GetErrors()
	if len(errors) != 1 {
		t.Fatalf("Expected 1 error, got %d", len(errors))
	}
	
	if errors[0].Context != "*invalid_label" {
		t.Errorf("Expected context '*invalid_label', got '%s'", errors[0].Context)
	}
}

func TestParseResult_Merge(t *testing.T) {
	result1 := NewParseResult()
	result1.AddError(SyntaxError, 10, 5, "Error 1")
	
	result2 := NewParseResult()
	result2.AddWarning(InvalidConfigError, 15, 3, "Warning 1")
	
	result1.Merge(result2)
	
	if result1.GetErrorCount() != 1 {
		t.Errorf("Expected 1 error after merge, got %d", result1.GetErrorCount())
	}
	if result1.GetWarningCount() != 1 {
		t.Errorf("Expected 1 warning after merge, got %d", result1.GetWarningCount())
	}
	if result1.GetIssueCount() != 2 {
		t.Errorf("Expected 2 total issues after merge, got %d", result1.GetIssueCount())
	}
}

func TestParseResult_Summary(t *testing.T) {
	tests := []struct {
		name     string
		setup    func(*ParseResult)
		expected string
	}{
		{
			name:     "no issues",
			setup:    func(pr *ParseResult) {},
			expected: "No issues found",
		},
		{
			name: "errors only",
			setup: func(pr *ParseResult) {
				pr.AddError(SyntaxError, 10, 5, "Error 1")
				pr.AddError(DuplicateLabelError, 15, 1, "Error 2")
			},
			expected: "Found 2 error(s)",
		},
		{
			name: "warnings only",
			setup: func(pr *ParseResult) {
				pr.AddWarning(InvalidConfigError, 10, 5, "Warning 1")
			},
			expected: "Found 1 warning(s)",
		},
		{
			name: "errors and warnings",
			setup: func(pr *ParseResult) {
				pr.AddError(SyntaxError, 10, 5, "Error 1")
				pr.AddWarning(InvalidConfigError, 15, 3, "Warning 1")
			},
			expected: "Found 1 error(s) and 1 warning(s)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := NewParseResult()
			tt.setup(result)
			
			summary := result.Summary()
			if summary != tt.expected {
				t.Errorf("ParseResult.Summary() = %v, want %v", summary, tt.expected)
			}
		})
	}
}
