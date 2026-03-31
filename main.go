package main

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/botamochi0x12/tyrano-parser-go/parser"
	"github.com/botamochi0x12/tyrano-parser-go/types"
)

func prettyPrint(v any) {
	b, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		log.Printf("-- prettyPrint error: %v", err)
		return
	}
	fmt.Println(string(b))
}

func main() {
	// Create a new parser with default options
	tyranoParser := parser.NewDefaultTyranoParser()

	// Example scenario content
	scenarioContent := `
;コメント
*start
[cm]
#akane
Hello, world![p]
[s]
`

	// Parse scenario
	scenario, err := tyranoParser.ParseScenario(scenarioContent)
	if err != nil {
		log.Fatalf("-- Error parsing scenario: %v", err)
	}

	fmt.Println("## Parsed scenario:")
	prettyPrint(scenario)

	// Example config content
	configContent := `
;Configuration file
title="My Game";
width=1280;
height=720;
`

	// Parse config
	config, err := tyranoParser.ParseConfig(configContent)
	if err != nil {
		log.Fatalf("-- Error parsing config: %v", err)
	}

	fmt.Println("## Parsed config:")
	prettyPrint(config)

	// Demonstrate error types
	parseErr := types.NewParseError(types.SyntaxError, 10, 5, "Invalid tag syntax")
	fmt.Println("## Example error:")
	prettyPrint(parseErr)
}
