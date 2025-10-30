package parser

// Token represents a lexical token
type Token struct {
	Type     TokenType
	Value    string
	Line     int
	Column   int
	Position int
}

// TokenType represents the type of a token
type TokenType int

const (
	TokenEOF TokenType = iota
	TokenText
	TokenTag
	TokenLabel
	TokenComment
	TokenCharacter
	TokenNewline
	TokenWhitespace
)

// String returns the string representation of the token type
func (tt TokenType) String() string {
	switch tt {
	case TokenEOF:
		return "EOF"
	case TokenText:
		return "TEXT"
	case TokenTag:
		return "TAG"
	case TokenLabel:
		return "LABEL"
	case TokenComment:
		return "COMMENT"
	case TokenCharacter:
		return "CHARACTER"
	case TokenNewline:
		return "NEWLINE"
	case TokenWhitespace:
		return "WHITESPACE"
	default:
		return "UNKNOWN"
	}
}

// Lexer handles tokenization of input text
type Lexer struct {
	input    string
	position int
	line     int
	column   int
}

// NewLexer creates a new Lexer
func NewLexer(input string) *Lexer {
	return &Lexer{
		input:    input,
		position: 0,
		line:     1,
		column:   1,
	}
}

// NextToken returns the next token from the input
func (l *Lexer) NextToken() Token {
	// Implementation will be added in later tasks
	return Token{Type: TokenEOF}
}

// PeekToken returns the next token without advancing the position
func (l *Lexer) PeekToken() Token {
	// Implementation will be added in later tasks
	return Token{Type: TokenEOF}
}

// SkipWhitespace skips whitespace characters
func (l *Lexer) SkipWhitespace() {
	// Implementation will be added in later tasks
}

// ReadString reads a string with the given delimiter
func (l *Lexer) ReadString(delimiter rune) string {
	// Implementation will be added in later tasks
	return ""
}
