package main

import (
	"fmt"
	"os"

	"github.com/botamochi0x12/tyrano-parser-go/loader"
	"github.com/botamochi0x12/tyrano-parser-go/parser"
	"github.com/botamochi0x12/tyrano-parser-go/types"
)

type configOutput struct {
	Kind   string             `json:"kind"`
	Path   string             `json:"path"`
	Config types.ConfigMap    `json:"config"`
	Issues []types.ParseIssue `json:"issues"`
}

func runConfig(args []string) int {
	g, rest, err := parseGlobalFlags(args)
	if err != nil {
		fmt.Fprintln(os.Stderr, "config:", err)
		return 3
	}
	cp := parser.NewConfigParser(false)
	var path string
	var config types.ConfigMap
	var result *types.ParseResult
	if len(rest) > 0 {
		path = rest[0]
		content, readErr := os.ReadFile(path)
		if readErr != nil {
			fmt.Fprintln(os.Stderr, "config:", readErr)
			return 2
		}
		config, result = cp.ParseWithResult(string(content))
	} else {
		layout, layoutErr := resolveLayout(g)
		if layoutErr != nil {
			fmt.Fprintln(os.Stderr, "config:", layoutErr)
			return 2
		}
		path = layout.ConfigPath
		var loadErr error
		config, result, loadErr = loader.LoadConfigFile(layout, cp)
		if loadErr != nil {
			fmt.Fprintln(os.Stderr, "config:", loadErr)
			return 2
		}
	}
	out := configOutput{Kind: "config", Path: path, Config: config, Issues: issuesOrEmpty(result, g.quiet)}
	if err := renderConfig(os.Stdout, out, g.format); err != nil {
		fmt.Fprintln(os.Stderr, "config:", err)
		return 1
	}
	return exitCodeForResult(result, g.strict)
}

func renderConfig(w *os.File, out configOutput, format string) error {
	if format == "report" {
		return renderConfigReport(w, out)
	}
	return renderJSON(w, out)
}
