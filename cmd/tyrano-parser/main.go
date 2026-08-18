package main

import (
	"fmt"
	"os"
)

func main() {
	os.Exit(mainWithArgs(os.Args[1:], os.Stdout, os.Stderr))
}

func mainWithArgs(args []string, stdout, stderr *os.File) int {
	if len(args) < 1 {
		usage(stderr)
		return 3
	}
	cmd := args[0]
	rest := args[1:]
	switch cmd {
	case "scenario":
		return runScenario(rest)
	case "config":
		return runConfig(rest)
	case "scan":
		return runScan(rest)
	case "version", "--version", "-version":
		printVersion(stdout)
		return 0
	case "-h", "--help", "help":
		usage(stdout)
		return 0
	default:
		fmt.Fprintf(stderr, "tyrano-parser: unknown command %q\n", cmd)
		usage(stderr)
		return 3
	}
}

func usage(w *os.File) {
	fmt.Fprint(w, `tyrano-parser <command> [flags] [args]

Commands:
  scenario <file.ks>           Parse one scenario file.
  config   [<Config.tjs>]      Parse one config file (default: auto-discover).
  scan     [<entrypoint.ks>]   Scan project (default: walk all scenarios).
  version                      Print the tyrano-parser version.

Global flags:
  --project-root <dir>   Skip auto-discovery, use this as project root.
  --format json|report   Output format (default: json).
  --strict               Strict mode: any parse issue exits non-zero.
  --quiet                Suppress warnings (errors still emitted).
`)
}
