package main

import (
	"encoding/json"
	"fmt"
	"io"

	"github.com/botamochi0x12/tyrano-parser-go/types"
)

func renderJSON(w io.Writer, v any) error {
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	if err := enc.Encode(v); err != nil {
		return fmt.Errorf("renderJSON: %w", err)
	}
	return nil
}

func issuesOrEmpty(result *types.ParseResult, quiet bool) []types.ParseIssue {
	if result == nil {
		return []types.ParseIssue{}
	}
	if !quiet {
		out := make([]types.ParseIssue, len(result.Issues))
		copy(out, result.Issues)
		return out
	}
	out := make([]types.ParseIssue, 0)
	for _, i := range result.Issues {
		if i.IsError() {
			out = append(out, i)
		}
	}
	return out
}

func exitCodeForResult(result *types.ParseResult, strict bool) int {
	if result == nil {
		return 0
	}
	if strict && result.HasIssues() {
		return 1
	}
	if result.HasErrors() {
		return 1
	}
	return 0
}
