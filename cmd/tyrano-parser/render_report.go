package main

import (
	"fmt"
	"io"
	"sort"

	"github.com/botamochi0x12/tyrano-parser-go/types"
)

func renderConfigReport(w io.Writer, out configOutput) error {
	fmt.Fprintf(w, "Config: %s\n", out.Path)
	fmt.Fprintf(w, "keys: %d\n\n", len(out.Config))
	keys := make([]string, 0, len(out.Config))
	for k := range out.Config {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		fmt.Fprintf(w, "  %s = %s\n", k, out.Config[k])
	}
	if len(out.Issues) > 0 {
		fmt.Fprintln(w, "\nIssues:")
		for _, i := range out.Issues {
			writeIssue(w, i)
		}
	}
	return nil
}

func renderScenarioReport(w io.Writer, out scenarioOutput) error {
	fmt.Fprintf(w, "Scenario: %s\n", out.Path)
	if out.Scenario != nil {
		fmt.Fprintf(w, "elements: %d\n", len(out.Scenario.Elements))
		fmt.Fprintf(w, "labels: %d\n", len(out.Scenario.Labels))
	}
	if len(out.Issues) > 0 {
		fmt.Fprintln(w, "\nIssues:")
		for _, i := range out.Issues {
			writeIssue(w, i)
		}
	}
	return nil
}

func renderScanReport(w io.Writer, out scanOutput) error {
	fmt.Fprintf(w, "Project: %s\n", out.Root)
	fmt.Fprintf(w, "Config keys: %d\n\n", len(out.Config))
	fmt.Fprintf(w, "Scenarios (%d):\n", len(out.Scenarios))
	scnNames := make([]string, 0, len(out.Scenarios))
	for k := range out.Scenarios {
		scnNames = append(scnNames, k)
	}
	sort.Strings(scnNames)
	for _, name := range scnNames {
		s := out.Scenarios[name]
		fmt.Fprintf(w, "  %s - %d elements, %d labels\n", name, len(s.Elements), len(s.Labels))
	}
	fmt.Fprintf(w, "\nCross-references (%d):\n", len(out.Refs))
	for _, r := range out.Refs {
		switch {
		case r.Storage != "" && r.Target != "":
			fmt.Fprintf(w, "  %s:%d  @%s storage=%q target=%q\n", r.From, r.Line, r.Tag, r.Storage, r.Target)
		case r.Storage != "":
			fmt.Fprintf(w, "  %s:%d  @%s storage=%q\n", r.From, r.Line, r.Tag, r.Storage)
		default:
			fmt.Fprintf(w, "  %s:%d  @%s target=%q\n", r.From, r.Line, r.Tag, r.Target)
		}
	}
	if len(out.Issues) > 0 {
		fmt.Fprintln(w, "\nIssues:")
		for _, i := range out.Issues {
			writeIssue(w, i)
		}
	}
	errors, warnings := 0, 0
	for _, i := range out.Issues {
		if i.IsError() {
			errors++
		} else if i.IsWarning() {
			warnings++
		}
	}
	fmt.Fprintf(w, "\nSummary: %d error(s), %d warning(s)\n", errors, warnings)
	return nil
}

func writeIssue(w io.Writer, i types.ParseIssue) {
	fmt.Fprintf(w, "  [%s] line %d:%d %s\n", i.Severity.String(), i.Line, i.Column, i.Message)
}
