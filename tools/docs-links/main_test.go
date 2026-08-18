package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestBrokenLinksReportsAMissingTarget(t *testing.T) {
	root := t.TempDir()
	writeDoc(t, root, "README.md", "See the [guide](docs/guide.md).\n")

	broken := findBrokenLinks(t, root)
	if len(broken) != 1 {
		t.Fatalf("brokenLinks() = %#v, want exactly one missing target", broken)
	}
	if broken[0].target != "docs/guide.md" || broken[0].source != "README.md" {
		t.Fatalf("brokenLinks() = %#v, want docs/guide.md reported from README.md", broken[0])
	}
	if broken[0].line != 1 {
		t.Errorf("broken link line = %d, want 1", broken[0].line)
	}
}

func TestBrokenLinksAcceptsAnExistingTarget(t *testing.T) {
	root := t.TempDir()
	writeDoc(t, root, "README.md", "See the [guide](docs/guide.md).\n")
	writeDoc(t, root, filepath.Join("docs", "guide.md"), "# Guide\n")

	if broken := findBrokenLinks(t, root); len(broken) != 0 {
		t.Fatalf("brokenLinks() = %#v, want none", broken)
	}
}

func TestBrokenLinksResolvesRelativeToTheLinkingFile(t *testing.T) {
	root := t.TempDir()
	writeDoc(t, root, "README.md", "# Readme\n")
	writeDoc(t, root, filepath.Join("docs", "guide.md"), "Back to the [readme](../README.md).\n")

	if broken := findBrokenLinks(t, root); len(broken) != 0 {
		t.Fatalf("brokenLinks() = %#v, want none", broken)
	}
}

func TestBrokenLinksIgnoresExternalAndAnchorOnlyLinks(t *testing.T) {
	root := t.TempDir()
	writeDoc(t, root, "README.md", "[web](https://example.com) [mail](mailto:a@example.com) [here](#features)\n")

	if broken := findBrokenLinks(t, root); len(broken) != 0 {
		t.Fatalf("brokenLinks() = %#v, want none", broken)
	}
}

func TestBrokenLinksChecksTheFilePartOfAnAnchoredLink(t *testing.T) {
	root := t.TempDir()
	writeDoc(t, root, "README.md", "See [releasing](docs/development.md#releases).\n")

	broken := findBrokenLinks(t, root)
	if len(broken) != 1 {
		t.Fatalf("brokenLinks() = %#v, want the missing file behind the anchor", broken)
	}
	if broken[0].target != "docs/development.md#releases" {
		t.Fatalf("broken link target = %q, want the link as written", broken[0].target)
	}
}

func TestBrokenLinksIgnoresFencedCodeBlocks(t *testing.T) {
	root := t.TempDir()
	writeDoc(t, root, "README.md", "```md\n[example](nowhere.md)\n```\n")

	if broken := findBrokenLinks(t, root); len(broken) != 0 {
		t.Fatalf("brokenLinks() = %#v, want fenced samples left alone", broken)
	}
}

// TestRepositoryMarkdownLinksResolve guards the documentation set itself, so
// splitting or moving a page cannot silently strand a link.
func TestRepositoryMarkdownLinksResolve(t *testing.T) {
	broken, err := brokenLinks(repositoryRoot(t))
	if err != nil {
		t.Fatalf("brokenLinks() error = %v", err)
	}
	for _, link := range broken {
		t.Errorf("%s:%d links to %q, which does not exist", link.source, link.line, link.target)
	}
}

func repositoryRoot(t *testing.T) string {
	t.Helper()
	root, err := filepath.Abs(filepath.Join("..", ".."))
	if err != nil {
		t.Fatal(err)
	}
	// Without this the guard above would pass vacuously if the walk ever
	// started somewhere that holds no documentation.
	if _, err := os.Stat(filepath.Join(root, "README.md")); err != nil {
		t.Fatalf("repository root %s has no README.md: %v", root, err)
	}
	return root
}

func findBrokenLinks(t *testing.T, root string) []brokenLink {
	t.Helper()
	broken, err := brokenLinks(root)
	if err != nil {
		t.Fatalf("brokenLinks() error = %v", err)
	}
	return broken
}

func writeDoc(t *testing.T, root, name, content string) {
	t.Helper()
	path := filepath.Join(root, name)
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}
