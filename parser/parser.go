package parser

import "github.com/tyranoscript/tyrano-parser-go/types"

// Parser defines the main interface for parsing TyranoScript files
type Parser interface {
	ParseScenario(content string) (*types.ParsedScenario, error)
	ParseConfig(content string) (types.ConfigMap, error)
	ParseScenarioWithResult(content string) (*types.ParsedScenario, *types.ParseResult)
	ParseConfigWithResult(content string) (types.ConfigMap, *types.ParseResult)
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
// In strict mode, returns error on first error. In lenient mode, collects warnings.
func (tp *TyranoParser) ParseScenario(content string) (*types.ParsedScenario, error) {
	scenario, result := tp.ParseScenarioWithResult(content)
	
	if tp.options.StrictMode && result.HasErrors() {
		return nil, result
	}
	
	return scenario, nil
}

// ParseConfig parses a Config.tjs file content and returns a ConfigMap
// In strict mode, returns error on first error. In lenient mode, collects warnings.
func (tp *TyranoParser) ParseConfig(content string) (types.ConfigMap, error) {
	config, result := tp.ParseConfigWithResult(content)
	
	if tp.options.StrictMode && result.HasErrors() {
		return nil, result
	}
	
	return config, nil
}

// ParseScenarioWithResult parses a scenario file and returns both the result and all issues
func (tp *TyranoParser) ParseScenarioWithResult(content string) (*types.ParsedScenario, *types.ParseResult) {
	scenarioParser := NewScenarioParserWithOptions(tp.options)
	return scenarioParser.ParseWithResult(content)
}

// ParseConfigWithResult parses a config file and returns both the result and all issues
func (tp *TyranoParser) ParseConfigWithResult(content string) (types.ConfigMap, *types.ParseResult) {
	configParser := NewConfigParserWithOptions(tp.options)
	return configParser.ParseWithResult(content)
}
