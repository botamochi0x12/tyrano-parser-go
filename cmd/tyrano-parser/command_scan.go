package main

import (
	"fmt"
	"os"

	"github.com/botamochi0x12/tyrano-parser-go/loader"
	"github.com/botamochi0x12/tyrano-parser-go/parser"
	"github.com/botamochi0x12/tyrano-parser-go/types"
)

type scanOutput struct {
	Kind      string                           `json:"kind"`
	Root      string                           `json:"root"`
	Config    types.ConfigMap                  `json:"config"`
	Scenarios map[string]*types.ParsedScenario `json:"scenarios"`
	Refs      []types.ScenarioRefRecord        `json:"refs"`
	Issues    []types.ParseIssue               `json:"issues"`
}

func runScan(args []string) int {
	g, rest, err := parseGlobalFlags(args)
	if err != nil {
		fmt.Fprintln(os.Stderr, "scan:", err)
		return 3
	}
	layout, err := resolveLayout(g)
	if err != nil {
		fmt.Fprintln(os.Stderr, "scan:", err)
		return 2
	}
	tp := parser.NewDefaultTyranoParser()
	var scan *types.ProjectScan
	if len(rest) > 0 {
		scan, err = loader.ResolveFrom(layout, tp, rest[0])
	} else {
		scan, err = loader.ScanProject(layout, tp)
	}
	if err != nil {
		fmt.Fprintln(os.Stderr, "scan:", err)
		return 2
	}
	out := scanOutput{
		Kind:      "project_scan",
		Root:      scan.Root,
		Config:    scan.Config,
		Scenarios: scan.Scenarios,
		Refs:      scan.Refs,
		Issues:    issuesOrEmpty(scan.Issues, g.quiet),
	}
	if err := renderScan(os.Stdout, out, g.format); err != nil {
		fmt.Fprintln(os.Stderr, "scan:", err)
		return 1
	}
	return exitCodeForResult(scan.Issues, g.strict)
}

func renderScan(w *os.File, out scanOutput, format string) error {
	if format == "report" {
		return renderScanReport(w, out)
	}
	return renderJSON(w, out)
}
