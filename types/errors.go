package types

import "fmt"

// ErrorType represents the type of parsing error
type ErrorType int

const (
	SyntaxError ErrorType = iota
	DuplicateLabelError
	UnmatchedIfError
	InvalidConfigError
)

// String returns the string representation of the error type
func (e ErrorType) String() string {
	switch e {
	case SyntaxError:
		return "SyntaxError"
	case DuplicateLabelError:
		return "DuplicateLabelError"
	case UnmatchedIfError:
		return "UnmatchedIfError"
	case InvalidConfigError:
		return "InvalidConfigError"
	default:
		return "UnknownError"
	}
}

// ParseError represents a parsing error with detailed context information
type ParseError struct {
	Line    int       `json:"line"`
	Column  int       `json:"column"`
	Message string    `json:"message"`
	Type    ErrorType `json:"type"`
}

// Error implements the error interface
func (pe *ParseError) Error() string {
	return fmt.Sprintf("%s at line %d, column %d: %s", pe.Type.String(), pe.Line, pe.Column, pe.Message)
}

// NewParseError creates a new ParseError
func NewParseError(errorType ErrorType, line, column int, message string) *ParseError {
	return &ParseError{
		Type:    errorType,
		Line:    line,
		Column:  column,
		Message: message,
	}
}
