package loader

import "github.com/botamochi0x12/tyrano-parser-go/types"

type ScenarioRef struct {
	From    string `json:"from"`
	Line    int    `json:"line"`
	Tag     string `json:"tag"`
	Storage string `json:"storage"`
	Target  string `json:"target"`
}

var refTags = map[string]bool{
	"call": true,
	"jump": true,
	"link": true,
}

func ExtractRefs(from string, scenario *types.ParsedScenario) []ScenarioRef {
	if scenario == nil {
		return nil
	}
	out := make([]ScenarioRef, 0)
	for _, tag := range scenario.Elements {
		if !refTags[tag.Name] {
			continue
		}
		storage := tag.Parameters["storage"]
		target := tag.Parameters["target"]
		if storage == "" && target == "" {
			continue
		}
		out = append(out, ScenarioRef{
			From:    from,
			Line:    tag.Line,
			Tag:     tag.Name,
			Storage: storage,
			Target:  target,
		})
	}
	return out
}
