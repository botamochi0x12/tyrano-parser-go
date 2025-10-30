# TyranoScript Parser - Go Implementation

A Go implementation of the TyranoScript KAG parser for parsing scenario files (.ks format) and configuration files (Config.tjs).

## Project Structure

```
tyrano-parser-go/
├── go.mod                 # Go module definition
├── main.go               # Example usage and demonstration
├── README.md             # This file
├── parser/               # Parser implementation package
│   ├── parser.go         # Main parser interface and TyranoParser
│   ├── scenario.go       # Scenario file parsing logic
│   ├── config.go         # Config.tjs parsing logic
│   ├── tag.go           # Tag parsing and parameter extraction
│   └── lexer.go         # Low-level tokenization
└── types/               # Data structures and types
    ├── scenario.go      # Scenario-related data structures
    ├── config.go        # Configuration data structures
    └── errors.go        # Custom error types
```

## Core Components

### Types Package
- **ParsedScenario**: Represents a parsed scenario with elements and labels
- **ParsedTag**: Represents individual parsed tags with parameters
- **LabelInfo**: Contains information about scenario labels
- **ConfigMap**: Key-value pairs from configuration files
- **ParseError**: Custom error type with detailed context

### Parser Package
- **Parser Interface**: Main parsing interface
- **TyranoParser**: Main parser implementation
- **ScenarioParser**: Handles .ks file parsing
- **ConfigParser**: Handles Config.tjs file parsing
- **TagParser**: Specialized tag and parameter parsing
- **Lexer**: Low-level tokenization

## Usage

```go
package main

import (
    "github.com/tyranoscript/tyrano-parser-go/parser"
)

func main() {
    // Create parser with default options
    tyranoParser := parser.NewDefaultTyranoParser()
    
    // Parse scenario file
    scenario, err := tyranoParser.ParseScenario(scenarioContent)
    if err != nil {
        // Handle error
    }
    
    // Parse config file
    config, err := tyranoParser.ParseConfig(configContent)
    if err != nil {
        // Handle error
    }
}
```

## Development Status

This is the foundation implementation with core data structures and interfaces defined. The actual parsing logic will be implemented in subsequent development phases following Test-Driven Development (TDD) methodology.

## Requirements

- Go 1.25.2 or later
- Compatible with KAG3/Kirikiri syntax
- Maintains compatibility with original JavaScript parser output format

## Building

```bash
go build -v
```

## Running

```bash
go run main.go
```
