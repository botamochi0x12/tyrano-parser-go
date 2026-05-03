package loader

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/botamochi0x12/tyrano-parser-go/parser"
	"github.com/botamochi0x12/tyrano-parser-go/types"
)

func ScanProject(layout *ProjectLayout, tp *parser.TyranoParser) (*types.ProjectScan, error) {
	scan := newProjectScan(layout)
	config, configResult, err := LoadConfigFile(layout, parser.NewConfigParser(false))
	if err != nil {
		return nil, fmt.Errorf("loader: scan config: %w", err)
	}
	scan.Config = config
	if configResult != nil {
		scan.Issues.Merge(configResult)
	}
	scenarios, walkResult, err := WalkScenarios(layout, tp)
	if err != nil {
		return nil, fmt.Errorf("loader: scan scenarios: %w", err)
	}
	scan.Scenarios = scenarios
	scan.Issues.Merge(walkResult)

	allRefs := make([]ScenarioRef, 0)
	for relPath, scenario := range scenarios {
		allRefs = append(allRefs, ExtractRefs(relPath, scenario)...)
	}
	for _, r := range allRefs {
		resolveRef(r, scenarios, scan.Issues)
	}
	scan.Refs = refRecords(allRefs)
	return scan, nil
}

func ResolveFrom(layout *ProjectLayout, tp *parser.TyranoParser, entrypoint string) (*types.ProjectScan, error) {
	scan := newProjectScan(layout)
	scan.Scenarios = make(map[string]*types.ParsedScenario)
	config, configResult, err := LoadConfigFile(layout, parser.NewConfigParser(false))
	if err != nil {
		return nil, fmt.Errorf("loader: resolve config: %w", err)
	}
	scan.Config = config
	if configResult != nil {
		scan.Issues.Merge(configResult)
	}
	startRel, err := relativizeScenario(layout, entrypoint)
	if err != nil {
		return nil, err
	}

	queue := []string{startRel}
	visited := map[string]bool{}
	allRefs := make([]ScenarioRef, 0)
	for len(queue) > 0 {
		rel := queue[0]
		queue = queue[1:]
		if visited[rel] {
			continue
		}
		visited[rel] = true
		scenario, result, loadErr := LoadScenarioFile(filepath.Join(layout.ScenarioDir, filepath.FromSlash(rel)), tp)
		if loadErr != nil {
			issue := types.NewParseError(types.MissingStorageError, 0, 0,
				fmt.Sprintf("referenced storage %q not found", rel))
			issue.WithContext("entrypoint or call chain")
			scan.Issues.AddIssue(*issue)
			continue
		}
		scan.Scenarios[rel] = scenario
		if result != nil {
			scan.Issues.Merge(result)
		}
		refs := ExtractRefs(rel, scenario)
		allRefs = append(allRefs, refs...)
		for _, r := range refs {
			storage := filepath.ToSlash(r.Storage)
			if storage != "" && !isDynamicRefValue(storage) && !visited[storage] {
				queue = append(queue, storage)
			}
		}
	}
	for _, r := range allRefs {
		resolveRefVisited(r, scan.Scenarios, scan.Issues)
	}
	scan.Refs = refRecords(allRefs)
	return scan, nil
}

func newProjectScan(layout *ProjectLayout) *types.ProjectScan {
	return &types.ProjectScan{
		Root:      layout.Root,
		Scenarios: nil,
		Refs:      nil,
		Issues:    types.NewParseResult(),
	}
}

func resolveRef(r ScenarioRef, scenarios map[string]*types.ParsedScenario, issues *types.ParseResult) {
	var labelHost *types.ParsedScenario
	if r.Storage != "" {
		if isDynamicRefValue(r.Storage) {
			return
		}
		host, ok := scenarios[filepath.ToSlash(r.Storage)]
		if !ok {
			issue := types.NewParseError(types.MissingStorageError, r.Line, 1,
				fmt.Sprintf("referenced storage %q not found", r.Storage))
			issue.WithContext(fmt.Sprintf("from=%s tag=%s", r.From, r.Tag))
			issues.AddIssue(*issue)
			return
		}
		labelHost = host
	} else {
		labelHost = scenarios[r.From]
	}
	if r.Target == "" || labelHost == nil || isDynamicRefValue(r.Target) {
		return
	}
	if _, ok := labelHost.Labels[trimTarget(r.Target)]; !ok {
		issue := types.NewParseError(types.MissingLabelError, r.Line, 1,
			fmt.Sprintf("referenced label %q not found", r.Target))
		issue.WithContext(fmt.Sprintf("from=%s tag=%s storage=%s", r.From, r.Tag, r.Storage))
		issues.AddIssue(*issue)
	}
}

func resolveRefVisited(r ScenarioRef, scenarios map[string]*types.ParsedScenario, issues *types.ParseResult) {
	if r.Target == "" {
		return
	}
	host := r.From
	if r.Storage != "" {
		if isDynamicRefValue(r.Storage) {
			return
		}
		host = filepath.ToSlash(r.Storage)
	}
	scenario, ok := scenarios[host]
	if !ok {
		return
	}
	if isDynamicRefValue(r.Target) {
		return
	}
	if _, ok := scenario.Labels[trimTarget(r.Target)]; !ok {
		issue := types.NewParseError(types.MissingLabelError, r.Line, 1,
			fmt.Sprintf("referenced label %q not found", r.Target))
		issue.WithContext(fmt.Sprintf("from=%s tag=%s storage=%s", r.From, r.Tag, r.Storage))
		issues.AddIssue(*issue)
	}
}

func relativizeScenario(layout *ProjectLayout, entrypoint string) (string, error) {
	if filepath.IsAbs(entrypoint) {
		rel, err := filepath.Rel(layout.ScenarioDir, entrypoint)
		if err != nil {
			return "", fmt.Errorf("loader: relativize %s: %w", entrypoint, err)
		}
		return filepath.ToSlash(rel), nil
	}
	abs, err := filepath.Abs(entrypoint)
	if err != nil {
		return "", fmt.Errorf("loader: abs %s: %w", entrypoint, err)
	}
	rel, err := filepath.Rel(layout.ScenarioDir, abs)
	if err == nil && !startsWithDotDot(rel) {
		return filepath.ToSlash(rel), nil
	}
	return filepath.ToSlash(entrypoint), nil
}

func startsWithDotDot(p string) bool {
	return p == ".." || strings.HasPrefix(filepath.ToSlash(p), "../")
}

func trimTarget(target string) string {
	return strings.TrimPrefix(target, "*")
}

func isDynamicRefValue(value string) bool {
	return strings.HasPrefix(strings.TrimSpace(value), "&")
}

func refRecords(refs []ScenarioRef) []types.ScenarioRefRecord {
	out := make([]types.ScenarioRefRecord, 0, len(refs))
	for _, r := range refs {
		out = append(out, types.ScenarioRefRecord{
			From: r.From, Line: r.Line, Tag: r.Tag,
			Storage: r.Storage, Target: r.Target,
		})
	}
	return out
}
