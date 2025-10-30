package parser

import "github.com/tyranoscript/tyrano-parser-go/types"

// Parser defines the main interface for parsing TyranoScript files
type Parser interface {
	ParseScenario(content string) (*types.ParsedScenario, error)
	ParseConfig(content string) (types.ConfigMap, error)
}

// ParserOptions contains configuration options for the parser
type ParserOptions struct {
	KeepSpaceInParameterValue string
	StrictMode               bool
	EnableWarnings          bool
}

// TyranoParser is the main implementation of the Parser interface
type TyranoParser struct {
	options ParserOptions
}

// NewTyranoParser creates a new TyranoParser with the given options
func NewTyranoParser(options ParserOptions) *TyranoParser {
	return &TyranoParser{
		options: options,
	}
}

// NewDefaultTyranoParser creates a new TyranoParser with default options
func NewDefaultTyranoParser() *TyranoParser {
	return &TyranoParser{
		options: ParserOptions{
			KeepSpaceInParameterValue: "false",
			StrictMode:               false,
			EnableWarnings:          true,
		},
	}
}

// ParseScenario parses a scenario file content and returns a ParsedScenario
func (tp *TyranoParser) ParseScenario(content string) (*types.ParsedScenario, error) {
	// Implementation will be added in later tasks
	return types.NewParsedScenario(), nil
}

// ParseConfig parses a Config.tjs file content and returns a ConfigMap
func (tp *TyranoParser) ParseConfig(content string) (types.ConfigMap, error) {
	configParser := NewConfigParser(tp.options.StrictMode)
	return configParser.Parse(content)
}
