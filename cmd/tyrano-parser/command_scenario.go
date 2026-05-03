package main

import (
	"fmt"
	"os"

	"github.com/botamochi0x12/tyrano-parser-go/loader"
	"github.com/botamochi0x12/tyrano-parser-go/parser"
	"github.com/botamochi0x12/tyrano-parser-go/types"
)

type scenarioOutput struct {
	Kind     string                `json:"kind"`
	Path     string                `json:"path"`
	Scenario *types.ParsedScenario `json:"scenario"`
	Issues   []types.ParseIssue    `json:"issues"`
}

func runScenario(args []string) int {
	g, rest, err := parseGlobalFlags(args)
	if err != nil {
		fmt.Fprintln(os.Stderr, "scenario:", err)
		return 3
	}
	if len(rest) == 0 {
		fmt.Fprintln(os.Stderr, "scenario: file path required")
		return 3
	}
	path := rest[0]
	scenario, result, err := loader.LoadScenarioFile(path, parser.NewDefaultTyranoParser())
	if err != nil {
		fmt.Fprintln(os.Stderr, "scenario:", err)
		return 2
	}
	out := scenarioOutput{
		Kind:     "scenario",
		Path:     path,
		Scenario: scenario,
		Issues:   issuesOrEmpty(result, g.quiet),
	}
	if err := renderScenario(os.Stdout, out, g.format); err != nil {
		fmt.Fprintln(os.Stderr, "scenario:", err)
		return 1
	}
	return exitCodeForResult(result, g.strict)
}

func renderScenario(w *os.File, out scenarioOutput, format string) error {
	if format == "report" {
		return renderScenarioReport(w, out)
	}
	return renderJSON(w, out)
}
