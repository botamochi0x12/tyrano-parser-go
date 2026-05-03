package main

import (
	"bytes"
	"strings"
	"testing"

	"github.com/botamochi0x12/tyrano-parser-go/types"
)

func TestRenderConfigReport_HappyPath(t *testing.T) {
	out := configOutput{
		Kind:   "config",
		Path:   "/p/data/system/Config.tjs",
		Config: types.ConfigMap{"System.title": "Foo", "scWidth": "800"},
	}
	var buf bytes.Buffer
	if err := renderConfigReport(&buf, out); err != nil {
		t.Fatal(err)
	}
	got := buf.String()
	for _, want := range []string{"Config: /p/data/system/Config.tjs", "keys: 2", "System.title", "scWidth"} {
		if !strings.Contains(got, want) {
			t.Errorf("missing %q in:\n%s", want, got)
		}
	}
}

func TestRenderConfigReport_WithErrors(t *testing.T) {
	issue := types.NewParseError(types.InvalidConfigError, 5, 1, "bad line")
	out := configOutput{Kind: "config", Path: "x", Config: types.ConfigMap{}, Issues: []types.ParseIssue{*issue}}
	var buf bytes.Buffer
	if err := renderConfigReport(&buf, out); err != nil {
		t.Fatal(err)
	}
	got := buf.String()
	if !strings.Contains(got, "[error]") || !strings.Contains(got, "bad line") {
		t.Errorf("expected error marker and message in:\n%s", got)
	}
}

func TestRenderScenarioReport_HappyPath(t *testing.T) {
	scenario := types.NewParsedScenario()
	scenario.Labels["start"] = types.NewLabelInfo("start", 1, 0, "")
	scenario.Elements = []types.ParsedTag{{Name: "cm", Line: 2}}
	out := scenarioOutput{Kind: "scenario", Path: "first.ks", Scenario: scenario}
	var buf bytes.Buffer
	if err := renderScenarioReport(&buf, out); err != nil {
		t.Fatal(err)
	}
	got := buf.String()
	for _, want := range []string{"Scenario: first.ks", "elements: 1", "labels: 1"} {
		if !strings.Contains(got, want) {
			t.Errorf("missing %q in:\n%s", want, got)
		}
	}
}

func TestRenderScanReport_HappyPath(t *testing.T) {
	out := scanOutput{
		Kind:      "project_scan",
		Root:      "/p",
		Config:    types.ConfigMap{"System.title": "X"},
		Scenarios: map[string]*types.ParsedScenario{"first.ks": types.NewParsedScenario()},
		Refs:      []types.ScenarioRefRecord{{From: "first.ks", Line: 2, Tag: "call", Storage: "tyrano.ks"}},
	}
	var buf bytes.Buffer
	if err := renderScanReport(&buf, out); err != nil {
		t.Fatal(err)
	}
	got := buf.String()
	for _, want := range []string{"Project: /p", "Scenarios (1)", "Cross-references (1)", "first.ks:2"} {
		if !strings.Contains(got, want) {
			t.Errorf("missing %q in:\n%s", want, got)
		}
	}
}
