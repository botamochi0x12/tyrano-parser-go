package main

import (
	"os"

	"github.com/botamochi0x12/tyrano-parser-go/loader"
)

func resolveLayout(g globalFlags) (*loader.ProjectLayout, error) {
	if g.projectRoot != "" {
		return loader.LayoutFrom(g.projectRoot)
	}
	cwd, err := os.Getwd()
	if err != nil {
		return nil, err
	}
	return loader.DiscoverRoot(cwd)
}
