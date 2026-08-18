package main

import (
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
)

func TestBuildReleaseLeavesOnlyPublishableFiles(t *testing.T) {
	t.Chdir(t.TempDir())

	archives, err := buildRelease("v1.2.3", func(target targetPlatform, output string) error {
		return os.WriteFile(output, []byte("binary for "+target.goos), 0o755)
	})
	if err != nil {
		t.Fatalf("buildRelease() error = %v", err)
	}
	if len(archives) != len(releasePlatforms()) {
		t.Fatalf("buildRelease() produced %d archives, want %d", len(archives), len(releasePlatforms()))
	}

	entries, err := os.ReadDir(distDir)
	if err != nil {
		t.Fatal(err)
	}
	var names []string
	for _, entry := range entries {
		names = append(names, entry.Name())
	}
	slices.Sort(names)
	want := []string{
		"checksums.txt",
		"tyrano-parser_v1.2.3_darwin_amd64.tar.gz",
		"tyrano-parser_v1.2.3_darwin_arm64.tar.gz",
		"tyrano-parser_v1.2.3_linux_amd64.tar.gz",
		"tyrano-parser_v1.2.3_linux_arm64.tar.gz",
		"tyrano-parser_v1.2.3_windows_amd64.zip",
	}
	if !slices.Equal(names, want) {
		t.Fatalf("dist holds %#v, want exactly %#v (the staging directory must not survive)", names, want)
	}
}

func TestBuildReleaseChecksumsCoverEveryArchive(t *testing.T) {
	t.Chdir(t.TempDir())

	archives, err := buildRelease("v1.2.3", func(target targetPlatform, output string) error {
		return os.WriteFile(output, []byte("binary for "+target.goos), 0o755)
	})
	if err != nil {
		t.Fatalf("buildRelease() error = %v", err)
	}

	sums, err := os.ReadFile(filepath.Join(distDir, "checksums.txt"))
	if err != nil {
		t.Fatal(err)
	}
	for _, archive := range archives {
		name := filepath.Base(archive)
		if !strings.Contains(string(sums), "  "+name+"\n") {
			t.Errorf("checksums.txt has no entry for %s:\n%s", name, sums)
		}
	}
}
