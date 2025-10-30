package parser

import "github.com/tyranoscript/tyrano-parser-go/types"

// ConfigParser handles parsing of Config.tjs files
type ConfigParser struct {
	strictMode bool
}

// NewConfigParser creates a new ConfigParser
func NewConfigParser(strictMode bool) *ConfigParser {
	return &ConfigParser{
		strictMode: strictMode,
	}
}

// Parse parses Config.tjs content and returns a ConfigMap
func (cp *ConfigParser) Parse(content string) (types.ConfigMap, error) {
	// Implementation will be added in later tasks
	return types.NewConfigMap(), nil
}
