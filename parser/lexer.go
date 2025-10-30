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
	if l.position >= len(l.input) {
		return Token{
			Type:     TokenEOF,
			Value:    "",
			Line:     l.line,
			Column:   l.column,
			Position: l.position,
		}
	}
	
	startPos := l.position
	startLine := l.line
	startColumn := l.column
	
	ch := rune(l.input[l.position])
	
	// Handle newlines
	if ch == '\n' {
		l.position++
		l.line++
		l.column = 1
		return Token{
			Type:     TokenNewline,
			Value:    "\n",
			Line:     startLine,
			Column:   startColumn,
			Position: startPos,
		}
	}
	
	// Handle carriage return + newline
	if ch == '\r' && l.position+1 < len(l.input) && l.input[l.position+1] == '\n' {
		l.position += 2
		l.line++
		l.column = 1
		return Token{
			Type:     TokenNewline,
			Value:    "\r\n",
			Line:     startLine,
			Column:   startColumn,
			Position: startPos,
		}
	}
	
	// Handle whitespace (spaces and tabs)
	if ch == ' ' || ch == '\t' {
		return l.readWhitespace()
	}
	
	// Handle comments
	if ch == ';' {
		return l.readSingleLineComment()
	}
	
	if ch == '/' && l.position+1 < len(l.input) && rune(l.input[l.position+1]) == '*' {
		return l.readBlockComment()
	}
	
	// Handle tags
	if ch == '[' {
		return l.readTag()
	}
	
	// Handle labels
	if ch == '*' {
		return l.readLabel()
	}
	
	// Handle character syntax
	if ch == '#' {
		return l.readCharacter()
	}
	
	// Handle regular text
	return l.readText()
}

// PeekToken returns the next token without advancing the position
func (l *Lexer) PeekToken() Token {
	// Save current state
	savedPos := l.position
	savedLine := l.line
	savedColumn := l.column
	
	// Get next token
	token := l.NextToken()
	
	// Restore state
	l.position = savedPos
	l.line = savedLine
	l.column = savedColumn
	
	return token
}

// SkipWhitespace skips whitespace characters
func (l *Lexer) SkipWhitespace() {
	for l.position < len(l.input) {
		ch := rune(l.input[l.position])
		
		if ch == ' ' || ch == '\t' {
			l.position++
			l.column++
		} else if ch == '\n' {
			l.position++
			l.line++
			l.column = 1
		} else if ch == '\r' && l.position+1 < len(l.input) && l.input[l.position+1] == '\n' {
			l.position += 2
			l.line++
			l.column = 1
		} else {
			break
		}
	}
}

// ReadString reads a string with the given delimiter
func (l *Lexer) ReadString(delimiter rune) string {
	var result []rune
	
	for l.position < len(l.input) {
		// Convert to runes for proper Unicode handling
		runes := []rune(l.input[l.position:])
		if len(runes) == 0 {
			break
		}
		
		r := runes[0]
		runeBytes := []byte(string(r))
		
		// Check for delimiter
		if r == delimiter {
			l.position += len(runeBytes)
			l.column++
			break
		}
		
		// Handle escape sequences
		if r == '\\' && len(runes) > 1 {
			l.position += len(runeBytes)
			l.column++
			
			nextR := runes[1]
			nextRuneBytes := []byte(string(nextR))
			l.position += len(nextRuneBytes)
			l.column++
			
			switch nextR {
			case 'n':
				result = append(result, '\n')
			case 't':
				result = append(result, '\t')
			case 'r':
				result = append(result, '\r')
			case '\\':
				result = append(result, '\\')
			case '"':
				result = append(result, '"')
			case '\'':
				result = append(result, '\'')
			default:
				result = append(result, nextR)
			}
		} else {
			result = append(result, r)
			l.position += len(runeBytes)
			if r == '\n' {
				l.line++
				l.column = 1
			} else {
				l.column++
			}
		}
	}
	
	return string(result)
}

// readWhitespace reads consecutive whitespace characters
func (l *Lexer) readWhitespace() Token {
	startPos := l.position
	startLine := l.line
	startColumn := l.column
	
	for l.position < len(l.input) {
		ch := rune(l.input[l.position])
		if ch != ' ' && ch != '\t' {
			break
		}
		l.position++
		l.column++
	}
	
	return Token{
		Type:     TokenWhitespace,
		Value:    l.input[startPos:l.position],
		Line:     startLine,
		Column:   startColumn,
		Position: startPos,
	}
}

// readSingleLineComment reads a single line comment starting with ;
func (l *Lexer) readSingleLineComment() Token {
	startPos := l.position
	startLine := l.line
	startColumn := l.column
	
	// Read until end of line or end of input
	for l.position < len(l.input) {
		ch := rune(l.input[l.position])
		if ch == '\n' || ch == '\r' {
			break
		}
		l.position++
		l.column++
	}
	
	return Token{
		Type:     TokenComment,
		Value:    l.input[startPos:l.position],
		Line:     startLine,
		Column:   startColumn,
		Position: startPos,
	}
}

// readBlockComment reads a block comment /* ... */
func (l *Lexer) readBlockComment() Token {
	startPos := l.position
	startLine := l.line
	startColumn := l.column
	
	// Skip the opening /*
	l.position += 2
	l.column += 2
	
	for l.position < len(l.input)-1 {
		ch := rune(l.input[l.position])
		
		if ch == '*' && rune(l.input[l.position+1]) == '/' {
			// Found closing */
			l.position += 2
			l.column += 2
			break
		}
		
		l.position++
		if ch == '\n' {
			l.line++
			l.column = 1
		} else {
			l.column++
		}
	}
	
	return Token{
		Type:     TokenComment,
		Value:    l.input[startPos:l.position],
		Line:     startLine,
		Column:   startColumn,
		Position: startPos,
	}
}

// readTag reads a tag [...]
func (l *Lexer) readTag() Token {
	startPos := l.position
	startLine := l.line
	startColumn := l.column
	
	bracketDepth := 0
	inQuotes := false
	quoteChar := rune(0)
	
	for l.position < len(l.input) {
		ch := rune(l.input[l.position])
		
		if !inQuotes {
			if ch == '[' {
				bracketDepth++
			} else if ch == ']' {
				bracketDepth--
				if bracketDepth == 0 {
					l.position++
					l.column++
					break
				}
			} else if ch == '"' || ch == '\'' {
				inQuotes = true
				quoteChar = ch
			}
		} else {
			if ch == quoteChar && (l.position == 0 || rune(l.input[l.position-1]) != '\\') {
				inQuotes = false
				quoteChar = 0
			}
		}
		
		l.position++
		if ch == '\n' {
			l.line++
			l.column = 1
		} else {
			l.column++
		}
	}
	
	return Token{
		Type:     TokenTag,
		Value:    l.input[startPos:l.position],
		Line:     startLine,
		Column:   startColumn,
		Position: startPos,
	}
}

// readLabel reads a label *label|description
func (l *Lexer) readLabel() Token {
	startPos := l.position
	startLine := l.line
	startColumn := l.column
	
	// Read until end of line (labels can contain spaces in description)
	for l.position < len(l.input) {
		ch := rune(l.input[l.position])
		if ch == '\n' || ch == '\r' {
			break
		}
		l.position++
		l.column++
	}
	
	return Token{
		Type:     TokenLabel,
		Value:    l.input[startPos:l.position],
		Line:     startLine,
		Column:   startColumn,
		Position: startPos,
	}
}

// readCharacter reads character syntax #character:expression
func (l *Lexer) readCharacter() Token {
	startPos := l.position
	startLine := l.line
	startColumn := l.column
	
	// Read until end of line or whitespace
	for l.position < len(l.input) {
		ch := rune(l.input[l.position])
		if ch == '\n' || ch == '\r' || ch == ' ' || ch == '\t' {
			break
		}
		l.position++
		l.column++
	}
	
	return Token{
		Type:     TokenCharacter,
		Value:    l.input[startPos:l.position],
		Line:     startLine,
		Column:   startColumn,
		Position: startPos,
	}
}

// readText reads regular text content
func (l *Lexer) readText() Token {
	startPos := l.position
	startLine := l.line
	startColumn := l.column
	
	// Read until end of line or special character
	for l.position < len(l.input) {
		ch := rune(l.input[l.position])
		if ch == '\n' || ch == '\r' || ch == '[' || ch == '*' || ch == '#' || ch == ';' {
			break
		}
		// Check for block comment start
		if ch == '/' && l.position+1 < len(l.input) && rune(l.input[l.position+1]) == '*' {
			break
		}
		l.position++
		l.column++
	}
	
	return Token{
		Type:     TokenText,
		Value:    l.input[startPos:l.position],
		Line:     startLine,
		Column:   startColumn,
		Position: startPos,
	}
}
