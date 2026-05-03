package types

type ProjectScan struct {
	Root      string                     `json:"root"`
	Config    ConfigMap                  `json:"config"`
	Scenarios map[string]*ParsedScenario `json:"scenarios"`
	Refs      []ScenarioRefRecord        `json:"refs"`
	Issues    *ParseResult               `json:"issues"`
}

type ScenarioRefRecord struct {
	From    string `json:"from"`
	Line    int    `json:"line"`
	Tag     string `json:"tag"`
	Storage string `json:"storage"`
	Target  string `json:"target"`
}
