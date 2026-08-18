package main

import (
	"fmt"
	"io"
	"sort"
	"strings"

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
		writeRef(w, r)
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

// writeRef prints one cross-reference. It shows the label without its leading
// asterisk so the line stays readable when the report is pasted into Markdown,
// falling back to the raw target while that target is still a runtime
// expression.
func writeRef(w io.Writer, r types.ScenarioRefRecord) {
	parts := []string{fmt.Sprintf("%s:%d", r.From, r.Line)}
	if kind := refKind(r); kind != "" {
		parts = append(parts, kind)
	}
	parts = append(parts, "@"+r.Tag)
	if r.Text != "" {
		parts = append(parts, fmt.Sprintf("%q", r.Text))
	}
	if r.Storage != "" {
		parts = append(parts, fmt.Sprintf("storage=%q", r.Storage))
	}
	switch {
	case r.Label != "":
		parts = append(parts, fmt.Sprintf("label=%q", r.Label))
	case r.Target != "":
		parts = append(parts, fmt.Sprintf("target=%q", r.Target))
	}
	if r.Macro != "" {
		parts = append(parts, fmt.Sprintf("(via %s)", r.Macro))
	}
	if len(r.Dynamic) > 0 {
		parts = append(parts, fmt.Sprintf("(unresolved: %s)", strings.Join(r.Dynamic, ", ")))
	}
	fmt.Fprintf(w, "  %s\n", strings.Join(parts, " "))
}

// refKind names what a reference is for: a system control reads as "ui" rather
// than as one more story choice.
func refKind(r types.ScenarioRefRecord) string {
	if r.UI {
		return "ui"
	}
	return r.Kind
}

func writeIssue(w io.Writer, i types.ParseIssue) {
	fmt.Fprintf(w, "  [%s] line %d:%d %s\n", i.Severity.String(), i.Line, i.Column, i.Message)
}
