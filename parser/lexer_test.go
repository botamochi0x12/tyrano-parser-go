package parser

import (
	"testing"
)

func TestNewLexer(t *testing.T) {
	input := "test input"
	lexer := NewLexer(input)
	
	if lexer == nil {
		t.Fatal("NewLexer() returned nil")
	}
	
	if lexer.input != input {
		t.Errorf("NewLexer().input = %v, want %v", lexer.input, input)
	}
	
	if lexer.position != 0 {
		t.Errorf("NewLexer().position = %v, want 0", lexer.position)
	}
	
	if lexer.line != 1 {
		t.Errorf("NewLexer().line = %v, want 1", lexer.line)
	}
	
	if lexer.column != 1 {
		t.Errorf("NewLexer().column = %v, want 1", lexer.column)
	}
}

func TestLexer_NextToken_EOF(t *testing.T) {
	lexer := NewLexer("")
	token := lexer.NextToken()
	
	if token.Type != TokenEOF {
		t.Errorf("NextToken() on empty input returned %v, want TokenEOF", token.Type)
	}
	
	if token.Line != 1 {
		t.Errorf("NextToken() line = %v, want 1", token.Line)
	}
	
	if token.Column != 1 {
		t.Errorf("NextToken() column = %v, want 1", token.Column)
	}
}

func TestLexer_NextToken_Whitespace(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []Token
	}{
		{
			name:  "single space",
			input: " ",
			expected: []Token{
				{Type: TokenWhitespace, Value: " ", Line: 1, Column: 1, Position: 0},
				{Type: TokenEOF, Value: "", Line: 1, Column: 2, Position: 1},
			},
		},
		{
			name:  "multiple spaces",
			input: "   ",
			expected: []Token{
				{Type: TokenWhitespace, Value: "   ", Line: 1, Column: 1, Position: 0},
				{Type: TokenEOF, Value: "", Line: 1, Column: 4, Position: 3},
			},
		},
		{
			name:  "tab character",
			input: "\t",
			expected: []Token{
				{Type: TokenWhitespace, Value: "\t", Line: 1, Column: 1, Position: 0},
				{Type: TokenEOF, Value: "", Line: 1, Column: 2, Position: 1},
			},
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lexer := NewLexer(tt.input)
			
			for i, expected := range tt.expected {
				token := lexer.NextToken()
				
				if token.Type != expected.Type {
					t.Errorf("NextToken()[%d].Type = %v, want %v", i, token.Type, expected.Type)
				}
				
				if token.Value != expected.Value {
					t.Errorf("NextToken()[%d].Value = %v, want %v", i, token.Value, expected.Value)
				}
				
				if token.Line != expected.Line {
					t.Errorf("NextToken()[%d].Line = %v, want %v", i, token.Line, expected.Line)
				}
				
				if token.Column != expected.Column {
					t.Errorf("NextToken()[%d].Column = %v, want %v", i, token.Column, expected.Column)
				}
				
				if token.Position != expected.Position {
					t.Errorf("NextToken()[%d].Position = %v, want %v", i, token.Position, expected.Position)
				}
			}
		})
	}
}

func TestLexer_NextToken_Newline(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []Token
	}{
		{
			name:  "single newline",
			input: "\n",
			expected: []Token{
				{Type: TokenNewline, Value: "\n", Line: 1, Column: 1, Position: 0},
				{Type: TokenEOF, Value: "", Line: 2, Column: 1, Position: 1},
			},
		},
		{
			name:  "carriage return newline",
			input: "\r\n",
			expected: []Token{
				{Type: TokenNewline, Value: "\r\n", Line: 1, Column: 1, Position: 0},
				{Type: TokenEOF, Value: "", Line: 2, Column: 1, Position: 2},
			},
		},
		{
			name:  "multiple newlines",
			input: "\n\n",
			expected: []Token{
				{Type: TokenNewline, Value: "\n", Line: 1, Column: 1, Position: 0},
				{Type: TokenNewline, Value: "\n", Line: 2, Column: 1, Position: 1},
				{Type: TokenEOF, Value: "", Line: 3, Column: 1, Position: 2},
			},
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lexer := NewLexer(tt.input)
			
			for i, expected := range tt.expected {
				token := lexer.NextToken()
				
				if token.Type != expected.Type {
					t.Errorf("NextToken()[%d].Type = %v, want %v", i, token.Type, expected.Type)
				}
				
				if token.Value != expected.Value {
					t.Errorf("NextToken()[%d].Value = %v, want %v", i, token.Value, expected.Value)
				}
				
				if token.Line != expected.Line {
					t.Errorf("NextToken()[%d].Line = %v, want %v", i, token.Line, expected.Line)
				}
				
				if token.Column != expected.Column {
					t.Errorf("NextToken()[%d].Column = %v, want %v", i, token.Column, expected.Column)
				}
			}
		})
	}
}

func TestLexer_NextToken_Comment(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []Token
	}{
		{
			name:  "single line comment",
			input: ";this is a comment",
			expected: []Token{
				{Type: TokenComment, Value: ";this is a comment", Line: 1, Column: 1, Position: 0},
				{Type: TokenEOF, Value: "", Line: 1, Column: 19, Position: 18},
			},
		},
		{
			name:  "block comment single line",
			input: "/* comment */",
			expected: []Token{
				{Type: TokenComment, Value: "/* comment */", Line: 1, Column: 1, Position: 0},
				{Type: TokenEOF, Value: "", Line: 1, Column: 14, Position: 13},
			},
		},
		{
			name:  "block comment multiline",
			input: "/* line 1\nline 2 */",
			expected: []Token{
				{Type: TokenComment, Value: "/* line 1\nline 2 */", Line: 1, Column: 1, Position: 0},
				{Type: TokenEOF, Value: "", Line: 2, Column: 10, Position: 19},
			},
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lexer := NewLexer(tt.input)
			
			for i, expected := range tt.expected {
				token := lexer.NextToken()
				
				if token.Type != expected.Type {
					t.Errorf("NextToken()[%d].Type = %v, want %v", i, token.Type, expected.Type)
				}
				
				if token.Value != expected.Value {
					t.Errorf("NextToken()[%d].Value = %v, want %v", i, token.Value, expected.Value)
				}
				
				if token.Line != expected.Line {
					t.Errorf("NextToken()[%d].Line = %v, want %v", i, token.Line, expected.Line)
				}
			}
		})
	}
}

func TestLexer_NextToken_Tag(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []Token
	}{
		{
			name:  "simple tag",
			input: "[cm]",
			expected: []Token{
				{Type: TokenTag, Value: "[cm]", Line: 1, Column: 1, Position: 0},
				{Type: TokenEOF, Value: "", Line: 1, Column: 5, Position: 4},
			},
		},
		{
			name:  "tag with parameters",
			input: "[bg storage=\"room.jpg\"]",
			expected: []Token{
				{Type: TokenTag, Value: "[bg storage=\"room.jpg\"]", Line: 1, Column: 1, Position: 0},
				{Type: TokenEOF, Value: "", Line: 1, Column: 23, Position: 22},
			},
		},
		{
			name:  "tag with nested brackets",
			input: "[eval exp=\"f.test = [1,2,3]\"]",
			expected: []Token{
				{Type: TokenTag, Value: "[eval exp=\"f.test = [1,2,3]\"]", Line: 1, Column: 1, Position: 0},
				{Type: TokenEOF, Value: "", Line: 1, Column: 30, Position: 29},
			},
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lexer := NewLexer(tt.input)
			
			for i, expected := range tt.expected {
				token := lexer.NextToken()
				
				if token.Type != expected.Type {
					t.Errorf("NextToken()[%d].Type = %v, want %v", i, token.Type, expected.Type)
				}
				
				if token.Value != expected.Value {
					t.Errorf("NextToken()[%d].Value = %v, want %v", i, token.Value, expected.Value)
				}
			}
		})
	}
}

func TestLexer_NextToken_Label(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []Token
	}{
		{
			name:  "simple label",
			input: "*start",
			expected: []Token{
				{Type: TokenLabel, Value: "*start", Line: 1, Column: 1, Position: 0},
				{Type: TokenEOF, Value: "", Line: 1, Column: 7, Position: 6},
			},
		},
		{
			name:  "label with description",
			input: "*start|Game Start",
			expected: []Token{
				{Type: TokenLabel, Value: "*start|Game Start", Line: 1, Column: 1, Position: 0},
				{Type: TokenEOF, Value: "", Line: 1, Column: 18, Position: 17},
			},
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lexer := NewLexer(tt.input)
			
			for i, expected := range tt.expected {
				token := lexer.NextToken()
				
				if token.Type != expected.Type {
					t.Errorf("NextToken()[%d].Type = %v, want %v", i, token.Type, expected.Type)
				}
				
				if token.Value != expected.Value {
					t.Errorf("NextToken()[%d].Value = %v, want %v", i, token.Value, expected.Value)
				}
			}
		})
	}
}

func TestLexer_NextToken_Character(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []Token
	}{
		{
			name:  "simple character",
			input: "#akane",
			expected: []Token{
				{Type: TokenCharacter, Value: "#akane", Line: 1, Column: 1, Position: 0},
				{Type: TokenEOF, Value: "", Line: 1, Column: 7, Position: 6},
			},
		},
		{
			name:  "character with expression",
			input: "#akane:happy",
			expected: []Token{
				{Type: TokenCharacter, Value: "#akane:happy", Line: 1, Column: 1, Position: 0},
				{Type: TokenEOF, Value: "", Line: 1, Column: 13, Position: 12},
			},
		},
		{
			name:  "character with Japanese name",
			input: "#あかね:嬉しい",
			expected: []Token{
				{Type: TokenCharacter, Value: "#あかね:嬉しい", Line: 1, Column: 1, Position: 0},
				{Type: TokenEOF, Value: "", Line: 1, Column: 8, Position: 13}, // Note: Unicode characters affect position differently
			},
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lexer := NewLexer(tt.input)
			
			for i, expected := range tt.expected {
				token := lexer.NextToken()
				
				if token.Type != expected.Type {
					t.Errorf("NextToken()[%d].Type = %v, want %v", i, token.Type, expected.Type)
				}
				
				if token.Value != expected.Value {
					t.Errorf("NextToken()[%d].Value = %v, want %v", i, token.Value, expected.Value)
				}
			}
		})
	}
}

func TestLexer_NextToken_Text(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []Token
	}{
		{
			name:  "simple text",
			input: "Hello world",
			expected: []Token{
				{Type: TokenText, Value: "Hello world", Line: 1, Column: 1, Position: 0},
				{Type: TokenEOF, Value: "", Line: 1, Column: 12, Position: 11},
			},
		},
		{
			name:  "text with Japanese characters",
			input: "こんにちは世界",
			expected: []Token{
				{Type: TokenText, Value: "こんにちは世界", Line: 1, Column: 1, Position: 0},
				{Type: TokenEOF, Value: "", Line: 1, Column: 8, Position: 21}, // Unicode affects position
			},
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lexer := NewLexer(tt.input)
			
			for i, expected := range tt.expected {
				token := lexer.NextToken()
				
				if token.Type != expected.Type {
					t.Errorf("NextToken()[%d].Type = %v, want %v", i, token.Type, expected.Type)
				}
				
				if token.Value != expected.Value {
					t.Errorf("NextToken()[%d].Value = %v, want %v", i, token.Value, expected.Value)
				}
			}
		})
	}
}

func TestLexer_PeekToken(t *testing.T) {
	lexer := NewLexer("hello world")
	
	// Peek should return the same token multiple times
	token1 := lexer.PeekToken()
	token2 := lexer.PeekToken()
	
	if token1.Type != token2.Type || token1.Value != token2.Value {
		t.Errorf("PeekToken() returned different tokens: %v vs %v", token1, token2)
	}
	
	// Next should return the same token as peek
	token3 := lexer.NextToken()
	
	if token1.Type != token3.Type || token1.Value != token3.Value {
		t.Errorf("PeekToken() and NextToken() returned different tokens: %v vs %v", token1, token3)
	}
	
	// Position should have advanced after NextToken
	if lexer.position == 0 {
		t.Error("Lexer position should have advanced after NextToken()")
	}
}

func TestLexer_SkipWhitespace(t *testing.T) {
	tests := []struct {
		name           string
		input          string
		expectedPos    int
		expectedLine   int
		expectedColumn int
	}{
		{
			name:           "no whitespace",
			input:          "hello",
			expectedPos:    0,
			expectedLine:   1,
			expectedColumn: 1,
		},
		{
			name:           "spaces only",
			input:          "   hello",
			expectedPos:    3,
			expectedLine:   1,
			expectedColumn: 4,
		},
		{
			name:           "tabs and spaces",
			input:          "\t  hello",
			expectedPos:    3,
			expectedLine:   1,
			expectedColumn: 4,
		},
		{
			name:           "newlines",
			input:          "\n\nhello",
			expectedPos:    2,
			expectedLine:   3,
			expectedColumn: 1,
		},
		{
			name:           "mixed whitespace",
			input:          " \t\n hello",
			expectedPos:    4,
			expectedLine:   2,
			expectedColumn: 2,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lexer := NewLexer(tt.input)
			lexer.SkipWhitespace()
			
			if lexer.position != tt.expectedPos {
				t.Errorf("SkipWhitespace() position = %v, want %v", lexer.position, tt.expectedPos)
			}
			
			if lexer.line != tt.expectedLine {
				t.Errorf("SkipWhitespace() line = %v, want %v", lexer.line, tt.expectedLine)
			}
			
			if lexer.column != tt.expectedColumn {
				t.Errorf("SkipWhitespace() column = %v, want %v", lexer.column, tt.expectedColumn)
			}
		})
	}
}

func TestLexer_ReadString(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		delimiter rune
		expected  string
		wantErr   bool
	}{
		{
			name:      "double quoted string",
			input:     "\"hello world\"",
			delimiter: '"',
			expected:  "hello world",
			wantErr:   false,
		},
		{
			name:      "single quoted string",
			input:     "'hello world'",
			delimiter: '\'',
			expected:  "hello world",
			wantErr:   false,
		},
		{
			name:      "string with escape sequences",
			input:     "\"hello\\nworld\\t!\"",
			delimiter: '"',
			expected:  "hello\nworld\t!",
			wantErr:   false,
		},
		{
			name:      "string with escaped quotes",
			input:     "\"hello \\\"world\\\"\"",
			delimiter: '"',
			expected:  "hello \"world\"",
			wantErr:   false,
		},
		{
			name:      "Japanese string",
			input:     "\"こんにちは世界\"",
			delimiter: '"',
			expected:  "こんにちは世界",
			wantErr:   false,
		},
		{
			name:      "empty string",
			input:     "\"\"",
			delimiter: '"',
			expected:  "",
			wantErr:   false,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lexer := NewLexer(tt.input)
			// Skip the opening delimiter
			lexer.position = 1
			lexer.column = 2
			
			result := lexer.ReadString(tt.delimiter)
			
			if result != tt.expected {
				t.Errorf("ReadString() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestLexer_ComplexScenario(t *testing.T) {
	input := `*start|Game Start
[cm]
#akane:happy
Hello world!
; This is a comment
[bg storage="room.jpg"]
[s]`
	
	lexer := NewLexer(input)
	
	expectedTokens := []TokenType{
		TokenLabel,     // *start|Game Start
		TokenNewline,   // \n
		TokenTag,       // [cm]
		TokenNewline,   // \n
		TokenCharacter, // #akane:happy
		TokenNewline,   // \n
		TokenText,      // Hello world!
		TokenNewline,   // \n
		TokenComment,   // ; This is a comment
		TokenNewline,   // \n
		TokenTag,       // [bg storage="room.jpg"]
		TokenNewline,   // \n
		TokenTag,       // [s]
		TokenEOF,
	}
	
	for i, expectedType := range expectedTokens {
		token := lexer.NextToken()
		if token.Type != expectedType {
			t.Errorf("Token[%d] type = %v, want %v (value: %q)", i, token.Type, expectedType, token.Value)
		}
	}
}
