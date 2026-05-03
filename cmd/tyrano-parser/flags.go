package main

import (
	"flag"
	"fmt"
)

type globalFlags struct {
	projectRoot string
	format      string
	strict      bool
	quiet       bool
}

func parseGlobalFlags(args []string) (globalFlags, []string, error) {
	fs := flag.NewFlagSet("global", flag.ContinueOnError)
	g := globalFlags{format: "json"}
	fs.StringVar(&g.projectRoot, "project-root", "", "explicit project root")
	fs.StringVar(&g.format, "format", "json", "output format: json or report")
	fs.BoolVar(&g.strict, "strict", false, "strict: any parse issue exits non-zero")
	fs.BoolVar(&g.quiet, "quiet", false, "suppress warnings")
	flagArgs, rest := splitFlags(args)
	if err := fs.Parse(flagArgs); err != nil {
		return g, nil, err
	}
	if g.format != "json" && g.format != "report" {
		return g, nil, fmt.Errorf("--format must be 'json' or 'report', got %q", g.format)
	}
	return g, rest, nil
}

func splitFlags(args []string) ([]string, []string) {
	flagArgs := make([]string, 0, len(args))
	rest := make([]string, 0, len(args))
	for i := 0; i < len(args); i++ {
		arg := args[i]
		switch arg {
		case "--":
			rest = append(rest, args[i+1:]...)
			return flagArgs, rest
		case "--strict", "-strict", "--quiet", "-quiet":
			flagArgs = append(flagArgs, arg)
		case "--project-root", "-project-root", "--format", "-format":
			flagArgs = append(flagArgs, arg)
			if i+1 < len(args) {
				i++
				flagArgs = append(flagArgs, args[i])
			}
		default:
			if len(arg) > 0 && arg[0] == '-' {
				flagArgs = append(flagArgs, arg)
			} else {
				rest = append(rest, arg)
			}
		}
	}
	return flagArgs, rest
}
