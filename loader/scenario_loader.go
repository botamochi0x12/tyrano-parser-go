package loader

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/botamochi0x12/tyrano-parser-go/parser"
	"github.com/botamochi0x12/tyrano-parser-go/types"
)

func LoadScenarioFile(path string, tp *parser.TyranoParser) (*types.ParsedScenario, *types.ParseResult, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return nil, nil, fmt.Errorf("loader: read scenario %s: %w", path, err)
	}
	scenario, result := tp.ParseScenarioWithResult(string(content))
	return scenario, result, nil
}

func WalkScenarios(layout *ProjectLayout, tp *parser.TyranoParser) (map[string]*types.ParsedScenario, *types.ParseResult, error) {
	out := make(map[string]*types.ParsedScenario)
	combined := types.NewParseResult()
	err := filepath.WalkDir(layout.ScenarioDir, func(path string, d fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if d.IsDir() || !strings.HasSuffix(strings.ToLower(d.Name()), ".ks") {
			return nil
		}
		scenario, result, loadErr := LoadScenarioFile(path, tp)
		if loadErr != nil {
			return loadErr
		}
		rel, relErr := filepath.Rel(layout.ScenarioDir, path)
		if relErr != nil {
			return fmt.Errorf("loader: rel path for %s: %w", path, relErr)
		}
		out[filepath.ToSlash(rel)] = scenario
		if result != nil {
			combined.Merge(result)
		}
		return nil
	})
	if err != nil {
		return nil, nil, fmt.Errorf("loader: walk scenarios %s: %w", layout.ScenarioDir, err)
	}
	return out, combined, nil
}
