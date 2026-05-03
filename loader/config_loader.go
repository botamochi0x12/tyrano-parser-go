package loader

import (
	"errors"
	"fmt"
	"io/fs"
	"os"

	"github.com/botamochi0x12/tyrano-parser-go/parser"
	"github.com/botamochi0x12/tyrano-parser-go/types"
)

func LoadConfigFile(layout *ProjectLayout, cp *parser.ConfigParser) (types.ConfigMap, *types.ParseResult, error) {
	content, err := os.ReadFile(layout.ConfigPath)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return nil, nil, nil
		}
		return nil, nil, fmt.Errorf("loader: read config %s: %w", layout.ConfigPath, err)
	}
	config, result := cp.ParseWithResult(string(content))
	return config, result, nil
}
