package types

import (
	"encoding/json"
	"testing"
)

func TestScenarioRefRecord_JSONCarriesBranchMetadata(t *testing.T) {
	record := ScenarioRefRecord{
		From:     "scene1.ks",
		Line:     12,
		Tag:      "link",
		Kind:     "choice",
		Storage:  "route_a.ks",
		Target:   "*good_end",
		Label:    "good_end",
		Text:     "森へ行く",
		Macro:    "sel",
		Dynamic:  []string{"storage"},
		Resolved: false,
		UI:       false,
	}

	encoded, err := json.Marshal(record)
	if err != nil {
		t.Fatalf("json.Marshal() error = %v", err)
	}
	var decoded map[string]any
	if err := json.Unmarshal(encoded, &decoded); err != nil {
		t.Fatalf("json.Unmarshal() error = %v", err)
	}

	want := map[string]any{
		"from":     "scene1.ks",
		"line":     float64(12),
		"tag":      "link",
		"kind":     "choice",
		"storage":  "route_a.ks",
		"target":   "*good_end",
		"label":    "good_end",
		"text":     "森へ行く",
		"macro":    "sel",
		"resolved": false,
		"ui":       false,
	}
	for key, value := range want {
		if decoded[key] != value {
			t.Errorf("json field %q = %#v, want %#v", key, decoded[key], value)
		}
	}
	dynamic, ok := decoded["dynamic"].([]any)
	if !ok || len(dynamic) != 1 || dynamic[0] != "storage" {
		t.Errorf("json field \"dynamic\" = %#v, want [storage]", decoded["dynamic"])
	}
}

func TestScenarioRefRecord_JSONOmitsEmptyMacroAndDynamic(t *testing.T) {
	encoded, err := json.Marshal(ScenarioRefRecord{From: "first.ks", Resolved: true})
	if err != nil {
		t.Fatalf("json.Marshal() error = %v", err)
	}
	var decoded map[string]any
	if err := json.Unmarshal(encoded, &decoded); err != nil {
		t.Fatalf("json.Unmarshal() error = %v", err)
	}
	for _, key := range []string{"macro", "dynamic"} {
		if _, present := decoded[key]; present {
			t.Errorf("json field %q present for an empty value, want it omitted", key)
		}
	}
}
