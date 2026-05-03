package parser

import (
	"testing"
)

func TestNewConfigParser(t *testing.T) {
	tests := []struct {
		name       string
		strictMode bool
	}{
		{
			name:       "strict mode enabled",
			strictMode: true,
		},
		{
			name:       "strict mode disabled",
			strictMode: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parser := NewConfigParser(tt.strictMode)

			if parser == nil {
				t.Fatal("NewConfigParser() returned nil")
			}

			if parser.options.StrictMode != tt.strictMode {
				t.Errorf("NewConfigParser().options.StrictMode = %v, want %v", parser.options.StrictMode, tt.strictMode)
			}
		})
	}
}

func TestConfigParser_Parse_ValidConfig(t *testing.T) {
	parser := NewConfigParser(false)

	tests := []struct {
		name     string
		content  string
		expected map[string]string
		wantErr  bool
	}{
		{
			name:     "empty content",
			content:  "",
			expected: map[string]string{},
			wantErr:  false,
		},
		{
			name:    "single quoted value",
			content: `;System.title = "My Game";`,
			expected: map[string]string{
				"System.title": "My Game",
			},
			wantErr: false,
		},
		{
			name:    "single unquoted value",
			content: `;scWidth = 1280;`,
			expected: map[string]string{
				"scWidth": "1280",
			},
			wantErr: false,
		},
		{
			name: "multiple values",
			content: `;System.title = "My Game";
;scWidth = 1280;
;scHeight = 720;`,
			expected: map[string]string{
				"System.title": "My Game",
				"scWidth":      "1280",
				"scHeight":     "720",
			},
			wantErr: false,
		},
		{
			name: "values with spaces",
			content: `;userFace = Quicksand, 游ゴシック体, "Yu Gothic";
;projectID = tyranoproject;`,
			expected: map[string]string{
				"userFace":  "Quicksand, 游ゴシック体, \"Yu Gothic\"",
				"projectID": "tyranoproject",
			},
			wantErr: false,
		},
		{
			name: "boolean values",
			content: `;configVisible = true;
;useCamera = false;`,
			expected: map[string]string{
				"configVisible": "true",
				"useCamera":     "false",
			},
			wantErr: false,
		},
		{
			name: "decimal values",
			content: `;game_version = 0.0;
;configThumbnailScale = 0.125;`,
			expected: map[string]string{
				"game_version":         "0.0",
				"configThumbnailScale": "0.125",
			},
			wantErr: false,
		},
		{
			name: "hex color values",
			content: `;frameColor = 0x000000;
;defaultChColor = 0xffffff;`,
			expected: map[string]string{
				"frameColor":     "0x000000",
				"defaultChColor": "0xffffff",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := parser.Parse(tt.content)

			if (err != nil) != tt.wantErr {
				t.Errorf("ConfigParser.Parse() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if len(result) != len(tt.expected) {
				t.Errorf("ConfigParser.Parse() returned %d items, want %d", len(result), len(tt.expected))
				return
			}

			for key, expectedValue := range tt.expected {
				if actualValue := result.Get(key); actualValue != expectedValue {
					t.Errorf("ConfigParser.Parse() result[%s] = %v, want %v", key, actualValue, expectedValue)
				}
			}
		})
	}
}

func TestConfigParser_Parse_CommentHandling(t *testing.T) {
	parser := NewConfigParser(false)

	tests := []struct {
		name     string
		content  string
		expected map[string]string
		wantErr  bool
	}{
		{
			name: "single line comments",
			content: `// This is a comment
;System.title = "My Game";
// Another comment`,
			expected: map[string]string{
				"System.title": "My Game",
			},
			wantErr: false,
		},
		{
			name: "block comments",
			content: `/* This is a 
   block comment */
;scWidth = 1280;
/* Another block comment */`,
			expected: map[string]string{
				"scWidth": "1280",
			},
			wantErr: false,
		},
		{
			name: "mixed comments",
			content: `// Single line comment
/* Block comment */
;System.title = "My Game";
// Another single line
;scWidth = 1280;
/* Another block */`,
			expected: map[string]string{
				"System.title": "My Game",
				"scWidth":      "1280",
			},
			wantErr: false,
		},
		{
			name: "comments with Japanese text",
			content: `// ティラノスクリプトの基本設定
;System.title = "ティラノスクリプト";
/* 画面サイズの設定 */
;scWidth = 1280;`,
			expected: map[string]string{
				"System.title": "ティラノスクリプト",
				"scWidth":      "1280",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := parser.Parse(tt.content)

			if (err != nil) != tt.wantErr {
				t.Errorf("ConfigParser.Parse() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if len(result) != len(tt.expected) {
				t.Errorf("ConfigParser.Parse() returned %d items, want %d", len(result), len(tt.expected))
				return
			}

			for key, expectedValue := range tt.expected {
				if actualValue := result.Get(key); actualValue != expectedValue {
					t.Errorf("ConfigParser.Parse() result[%s] = %v, want %v", key, actualValue, expectedValue)
				}
			}
		})
	}
}

func TestConfigParser_Parse_ErrorCases(t *testing.T) {
	parser := NewConfigParser(true) // Use strict mode for error testing

	tests := []struct {
		name    string
		content string
		wantErr bool
	}{
		{
			name:    "missing semicolon is allowed",
			content: `;System.title = "My Game"`,
			wantErr: false,
		},
		{
			name:    "missing equals sign",
			content: `;System.title "My Game";`,
			wantErr: true,
		},
		{
			name:    "unmatched quotes",
			content: `;System.title = "My Game;`,
			wantErr: true,
		},
		{
			name:    "empty key",
			content: `; = "value";`,
			wantErr: true,
		},
		{
			name:    "invalid line format",
			content: `;invalid line without proper format;`,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := parser.Parse(tt.content)

			if (err != nil) != tt.wantErr {
				t.Errorf("ConfigParser.Parse() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestConfigParser_Parse_LenientMode(t *testing.T) {
	parser := NewConfigParser(false) // Use lenient mode

	tests := []struct {
		name     string
		content  string
		expected map[string]string
		wantErr  bool
	}{
		{
			name:    "missing semicolon - lenient",
			content: `;System.title = "My Game"`,
			expected: map[string]string{
				"System.title": "My Game",
			},
			wantErr: false,
		},
		{
			name: "mixed valid and invalid lines",
			content: `;System.title = "My Game";
invalid line
;scWidth = 1280;`,
			expected: map[string]string{
				"System.title": "My Game",
				"scWidth":      "1280",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := parser.Parse(tt.content)

			if (err != nil) != tt.wantErr {
				t.Errorf("ConfigParser.Parse() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if len(result) != len(tt.expected) {
				t.Errorf("ConfigParser.Parse() returned %d items, want %d", len(result), len(tt.expected))
				return
			}

			for key, expectedValue := range tt.expected {
				if actualValue := result.Get(key); actualValue != expectedValue {
					t.Errorf("ConfigParser.Parse() result[%s] = %v, want %v", key, actualValue, expectedValue)
				}
			}
		})
	}
}

func TestConfigParser_Parse_WhitespaceHandling(t *testing.T) {
	parser := NewConfigParser(false)

	tests := []struct {
		name     string
		content  string
		expected map[string]string
		wantErr  bool
	}{
		{
			name: "extra whitespace",
			content: `  ;  System.title  =  "My Game"  ;  
			;   scWidth   =   1280   ;   `,
			expected: map[string]string{
				"System.title": "My Game",
				"scWidth":      "1280",
			},
			wantErr: false,
		},
		{
			name:    "tabs and spaces",
			content: "\t;\tSystem.title\t=\t\"My Game\"\t;\t\n\t;\tscWidth\t=\t1280\t;\t",
			expected: map[string]string{
				"System.title": "My Game",
				"scWidth":      "1280",
			},
			wantErr: false,
		},
		{
			name: "empty lines",
			content: `;System.title = "My Game";

;scWidth = 1280;


;scHeight = 720;`,
			expected: map[string]string{
				"System.title": "My Game",
				"scWidth":      "1280",
				"scHeight":     "720",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := parser.Parse(tt.content)

			if (err != nil) != tt.wantErr {
				t.Errorf("ConfigParser.Parse() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if len(result) != len(tt.expected) {
				t.Errorf("ConfigParser.Parse() returned %d items, want %d", len(result), len(tt.expected))
				return
			}

			for key, expectedValue := range tt.expected {
				if actualValue := result.Get(key); actualValue != expectedValue {
					t.Errorf("ConfigParser.Parse() result[%s] = %v, want %v", key, actualValue, expectedValue)
				}
			}
		})
	}
}

func TestConfigParser_RealFormat_HappyPath(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantKey string
		wantVal string
	}{
		{
			name:    "leading semicolon prefix with trailing semicolon",
			input:   `;System.title = "StarGazers";`,
			wantKey: "System.title",
			wantVal: "StarGazers",
		},
		{
			name:    "leading semicolon prefix without trailing semicolon",
			input:   `;scWidth = 1280`,
			wantKey: "scWidth",
			wantVal: "1280",
		},
		{
			name:    "unquoted identifier value",
			input:   `;ScreenRatio = fix;`,
			wantKey: "ScreenRatio",
			wantVal: "fix",
		},
		{
			name:    "boolean value",
			input:   `;use3D = false;`,
			wantKey: "use3D",
			wantVal: "false",
		},
		{
			name:    "dotted key",
			input:   `;scPositionX.left = 160;`,
			wantKey: "scPositionX.left",
			wantVal: "160",
		},
		{
			name:    "hex literal",
			input:   `;defaultChColor = 0xffffff;`,
			wantKey: "defaultChColor",
			wantVal: "0xffffff",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cp := NewConfigParser(false)
			config, err := cp.Parse(tt.input)
			if err != nil {
				t.Fatalf("Parse returned error: %v", err)
			}
			got, ok := config[tt.wantKey]
			if !ok {
				t.Fatalf("key %q missing from config; got: %#v", tt.wantKey, config)
			}
			if got != tt.wantVal {
				t.Errorf("config[%q] = %q, want %q", tt.wantKey, got, tt.wantVal)
			}
		})
	}
}

func TestConfigParser_RealFormat_InlineComments(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantKey string
		wantVal string
	}{
		{
			name:    "inline comment after value",
			input:   `;cursorDefault = default; // 通常のマウスカーソル`,
			wantKey: "cursorDefault",
			wantVal: "default",
		},
		{
			name:    "inline comment without trailing semicolon",
			input:   `;configLeft    = -1     //コンフィグアイコンの左位置を指定`,
			wantKey: "configLeft",
			wantVal: "-1",
		},
		{
			name:    "double-slash inside quoted value is preserved",
			input:   `;url = "http://example.com/path";`,
			wantKey: "url",
			wantVal: "http://example.com/path",
		},
		{
			name:    "comma-separated quoted value with embedded quotes",
			input:   `;userFace = Quicksand, "Yu Gothic", sans-serif;`,
			wantKey: "userFace",
			wantVal: `Quicksand, "Yu Gothic", sans-serif`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cp := NewConfigParser(false)
			config, err := cp.Parse(tt.input)
			if err != nil {
				t.Fatalf("Parse returned error: %v", err)
			}
			got, ok := config[tt.wantKey]
			if !ok {
				t.Fatalf("key %q missing; got: %#v", tt.wantKey, config)
			}
			if got != tt.wantVal {
				t.Errorf("config[%q] = %q, want %q", tt.wantKey, got, tt.wantVal)
			}
		})
	}
}

func TestConfigParser_RealFormat_SkipComments(t *testing.T) {
	input := `// pure comment line
// another comment
;System.title = "StarGazers";
// trailing comment`
	cp := NewConfigParser(false)
	config, err := cp.Parse(input)
	if err != nil {
		t.Fatalf("Parse returned error: %v", err)
	}
	if len(config) != 1 {
		t.Errorf("expected 1 key, got %d: %#v", len(config), config)
	}
	if config["System.title"] != "StarGazers" {
		t.Errorf("System.title = %q, want %q", config["System.title"], "StarGazers")
	}
}

func TestStripInlineComment(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want string
	}{
		{"no comment", `default`, `default`},
		{"trailing comment", `default // note`, `default`},
		{"comment after semicolon-stripped value", `default ; // note`, `default ;`},
		{"// inside double quotes", `"http://example.com"`, `"http://example.com"`},
		{"// inside single quotes", `'a//b'`, `'a//b'`},
		{"escaped quote then //", `"a\"b" // c`, `"a\"b"`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := stripInlineComment(tt.in)
			if got != tt.want {
				t.Errorf("stripInlineComment(%q) = %q, want %q", tt.in, got, tt.want)
			}
		})
	}
}
