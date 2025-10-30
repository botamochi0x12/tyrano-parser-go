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
