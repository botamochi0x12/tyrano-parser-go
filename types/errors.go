package types

import "fmt"

// ErrorType represents the type of parsing error
type ErrorType int

const (
	SyntaxError ErrorType = iota
	DuplicateLabelError
	UnmatchedIfError
	InvalidConfigError
	MalformedTagError
	UnknownTagError
	MissingParameterError
	InvalidParameterError
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
	case MalformedTagError:
		return "MalformedTagError"
	case UnknownTagError:
		return "UnknownTagError"
	case MissingParameterError:
		return "MissingParameterError"
	case InvalidParameterError:
		return "InvalidParameterError"
	default:
		return "UnknownError"
	}
}

// Severity represents the severity level of an issue
type Severity int

const (
	Error Severity = iota
	Warning
	Info
)

// String returns the string representation of the severity
func (s Severity) String() string {
	switch s {
	case Error:
		return "error"
	case Warning:
		return "warning"
	case Info:
		return "info"
	default:
		return "unknown"
	}
}

// ParseIssue represents a parsing issue (error or warning) with detailed context information
type ParseIssue struct {
	Line     int       `json:"line"`
	Column   int       `json:"column"`
	Message  string    `json:"message"`
	Type     ErrorType `json:"type"`
	Severity Severity  `json:"severity"`
	Context  string    `json:"context,omitempty"` // Additional context like the problematic line content
}

// Error implements the error interface
func (pi *ParseIssue) Error() string {
	return fmt.Sprintf("%s %s at line %d, column %d: %s", pi.Severity.String(), pi.Type.String(), pi.Line, pi.Column, pi.Message)
}

// IsError returns true if this issue is an error (not a warning)
func (pi *ParseIssue) IsError() bool {
	return pi.Severity == Error
}

// IsWarning returns true if this issue is a warning
func (pi *ParseIssue) IsWarning() bool {
	return pi.Severity == Warning
}

// NewParseError creates a new ParseIssue with Error severity
func NewParseError(errorType ErrorType, line, column int, message string) *ParseIssue {
	return &ParseIssue{
		Type:     errorType,
		Line:     line,
		Column:   column,
		Message:  message,
		Severity: Error,
	}
}

// NewParseWarning creates a new ParseIssue with Warning severity
func NewParseWarning(errorType ErrorType, line, column int, message string) *ParseIssue {
	return &ParseIssue{
		Type:     errorType,
		Line:     line,
		Column:   column,
		Message:  message,
		Severity: Warning,
	}
}

// NewParseIssue creates a new ParseIssue with specified severity
func NewParseIssue(errorType ErrorType, severity Severity, line, column int, message string) *ParseIssue {
	return &ParseIssue{
		Type:     errorType,
		Line:     line,
		Column:   column,
		Message:  message,
		Severity: severity,
	}
}

// WithContext adds context information to the issue
func (pi *ParseIssue) WithContext(context string) *ParseIssue {
	pi.Context = context
	return pi
}

// ParseError represents a parsing error with detailed context information
// Deprecated: Use ParseIssue instead
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

// ParseResult represents the result of a parsing operation with errors and warnings
type ParseResult struct {
	Issues []ParseIssue `json:"issues"`
}

// NewParseResult creates a new ParseResult
func NewParseResult() *ParseResult {
	return &ParseResult{
		Issues: make([]ParseIssue, 0),
	}
}

// AddError adds an error to the result
func (pr *ParseResult) AddError(errorType ErrorType, line, column int, message string) {
	issue := NewParseError(errorType, line, column, message)
	pr.Issues = append(pr.Issues, *issue)
}

// AddWarning adds a warning to the result
func (pr *ParseResult) AddWarning(errorType ErrorType, line, column int, message string) {
	issue := NewParseWarning(errorType, line, column, message)
	pr.Issues = append(pr.Issues, *issue)
}

// AddIssue adds an issue to the result
func (pr *ParseResult) AddIssue(issue ParseIssue) {
	pr.Issues = append(pr.Issues, issue)
}

// AddErrorWithContext adds an error with context to the result
func (pr *ParseResult) AddErrorWithContext(errorType ErrorType, line, column int, message, context string) {
	issue := NewParseError(errorType, line, column, message).WithContext(context)
	pr.Issues = append(pr.Issues, *issue)
}

// AddWarningWithContext adds a warning with context to the result
func (pr *ParseResult) AddWarningWithContext(errorType ErrorType, line, column int, message, context string) {
	issue := NewParseWarning(errorType, line, column, message).WithContext(context)
	pr.Issues = append(pr.Issues, *issue)
}

// HasErrors returns true if there are any errors in the result
func (pr *ParseResult) HasErrors() bool {
	for _, issue := range pr.Issues {
		if issue.IsError() {
			return true
		}
	}
	return false
}

// HasWarnings returns true if there are any warnings in the result
func (pr *ParseResult) HasWarnings() bool {
	for _, issue := range pr.Issues {
		if issue.IsWarning() {
			return true
		}
	}
	return false
}

// HasIssues returns true if there are any issues (errors or warnings) in the result
func (pr *ParseResult) HasIssues() bool {
	return len(pr.Issues) > 0
}

// GetErrors returns all errors from the result
func (pr *ParseResult) GetErrors() []ParseIssue {
	var errors []ParseIssue
	for _, issue := range pr.Issues {
		if issue.IsError() {
			errors = append(errors, issue)
		}
	}
	return errors
}

// GetWarnings returns all warnings from the result
func (pr *ParseResult) GetWarnings() []ParseIssue {
	var warnings []ParseIssue
	for _, issue := range pr.Issues {
		if issue.IsWarning() {
			warnings = append(warnings, issue)
		}
	}
	return warnings
}

// GetIssueCount returns the total number of issues
func (pr *ParseResult) GetIssueCount() int {
	return len(pr.Issues)
}

// GetErrorCount returns the number of errors
func (pr *ParseResult) GetErrorCount() int {
	return len(pr.GetErrors())
}

// GetWarningCount returns the number of warnings
func (pr *ParseResult) GetWarningCount() int {
	return len(pr.GetWarnings())
}

// Clear removes all issues from the result
func (pr *ParseResult) Clear() {
	pr.Issues = make([]ParseIssue, 0)
}

// Merge combines issues from another ParseResult into this one
func (pr *ParseResult) Merge(other *ParseResult) {
	if other != nil {
		pr.Issues = append(pr.Issues, other.Issues...)
	}
}

// Error returns the first error message, implementing the error interface
func (pr *ParseResult) Error() string {
	errors := pr.GetErrors()
	if len(errors) > 0 {
		return errors[0].Error()
	}
	return "no errors"
}

// Summary returns a summary string of all issues
func (pr *ParseResult) Summary() string {
	if !pr.HasIssues() {
		return "No issues found"
	}
	
	errorCount := pr.GetErrorCount()
	warningCount := pr.GetWarningCount()
	
	if errorCount > 0 && warningCount > 0 {
		return fmt.Sprintf("Found %d error(s) and %d warning(s)", errorCount, warningCount)
	} else if errorCount > 0 {
		return fmt.Sprintf("Found %d error(s)", errorCount)
	} else {
		return fmt.Sprintf("Found %d warning(s)", warningCount)
	}
}
