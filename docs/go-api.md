# Go API

[← README](../README.md)

The library splits into two layers. `parser/` is pure string-in / struct-out and
never touches the filesystem; `loader/` adds file I/O, project discovery, and
cross-reference scanning on top of it.

## Pure parsers

Use `parser` when the content is already in memory.

```go
package main

import "github.com/botamochi0x12/tyrano-parser-go/parser"

func parseStrings(scenarioContent, configContent string) error {
    p := parser.NewDefaultTyranoParser()

    scenario, scenarioResult := p.ParseScenarioWithResult(scenarioContent)
    _ = scenario
    if scenarioResult.HasErrors() {
        return scenarioResult
    }

    config, configResult := p.ParseConfigWithResult(configContent)
    _ = config
    if configResult.HasErrors() {
        return configResult
    }

    return nil
}
```

## Loader and project scans

Use `loader` to read a project from disk and check its cross-references.

```go
package main

import (
    "github.com/botamochi0x12/tyrano-parser-go/loader"
    "github.com/botamochi0x12/tyrano-parser-go/parser"
)

func scanProject(root string) error {
    layout, err := loader.LayoutFrom(root)
    if err != nil {
        return err
    }

    scan, err := loader.ScanProject(layout, parser.NewDefaultTyranoParser())
    if err != nil {
        return err
    }

    if scan.Issues.HasErrors() {
        return scan.Issues
    }
    return nil
}
```

`loader.DiscoverRoot` walks upward from a starting directory until it finds
`data/scenario/` or `data/system/Config.tjs`. `loader.LayoutFrom` takes an
explicit project root and requires `data/scenario/`. The expected directory
shape is described in [TyranoScript support](tyranoscript-support.md).

A runnable version of this lives in `cmd/tyrano-parser-example/`:

```console
go run ./cmd/tyrano-parser-example
```
