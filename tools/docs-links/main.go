// Command docs-links reports relative links in the repository's Markdown files
// whose target does not exist, so splitting or moving a page cannot silently
// strand a reader.
package main

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// brokenLink is one Markdown link pointing at a file that is not there.
type brokenLink struct {
	source string
	target string
	line   int
}

// markdownLink matches both [text](target) and ![alt](target).
var markdownLink = regexp.MustCompile(`\[[^\]]*\]\(([^)]+)\)`)

func brokenLinks(root string) ([]brokenLink, error) {
	var broken []brokenLink
	err := filepath.WalkDir(root, func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.IsDir() {
			if entry.Name() == ".git" {
				return fs.SkipDir
			}
			return nil
		}
		if strings.ToLower(filepath.Ext(path)) != ".md" {
			return nil
		}
		found, err := brokenLinksIn(root, path)
		if err != nil {
			return err
		}
		broken = append(broken, found...)
		return nil
	})
	if err != nil {
		return nil, err
	}
	return broken, nil
}

func brokenLinksIn(root, path string) ([]brokenLink, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	source, err := filepath.Rel(root, path)
	if err != nil {
		return nil, err
	}

	var broken []brokenLink
	inFence := false
	for i, line := range strings.Split(string(data), "\n") {
		if strings.HasPrefix(strings.TrimSpace(line), "```") {
			inFence = !inFence
			continue
		}
		if inFence {
			continue
		}
		for _, match := range markdownLink.FindAllStringSubmatch(line, -1) {
			target := strings.TrimSpace(match[1])
			if !isLocalTarget(target) {
				continue
			}
			resolved := filepath.Join(filepath.Dir(path), filePart(target))
			if _, err := os.Stat(resolved); err != nil {
				broken = append(broken, brokenLink{source: source, target: target, line: i + 1})
			}
		}
	}
	return broken, nil
}

// isLocalTarget keeps out anything that does not name a file in the repository:
// absolute URLs, mail links, and same-page anchors.
func isLocalTarget(target string) bool {
	if target == "" || strings.HasPrefix(target, "#") {
		return false
	}
	if strings.Contains(target, "://") || strings.HasPrefix(target, "mailto:") {
		return false
	}
	return true
}

// filePart drops the "#anchor" suffix, leaving the path the link points at.
func filePart(target string) string {
	if index := strings.Index(target, "#"); index >= 0 {
		return target[:index]
	}
	return target
}

func main() {
	root, err := os.Getwd()
	if err != nil {
		fmt.Fprintf(os.Stderr, "docs-links failed: %v\n", err)
		os.Exit(1)
	}
	broken, err := brokenLinks(root)
	if err != nil {
		fmt.Fprintf(os.Stderr, "docs-links failed: %v\n", err)
		os.Exit(1)
	}
	for _, link := range broken {
		fmt.Fprintf(os.Stderr, "%s:%d: broken link to %q\n", link.source, link.line, link.target)
	}
	if len(broken) > 0 {
		os.Exit(1)
	}
	fmt.Println("all markdown links resolve")
}
