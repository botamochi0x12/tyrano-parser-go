package main

import (
	"fmt"
	"io"
)

// version is stamped at build time with -ldflags "-X main.version=<tag>".
// Local builds keep the "dev" placeholder.
var version = "dev"

func printVersion(w io.Writer) {
	fmt.Fprintf(w, "tyrano-parser %s\n", version)
}
