package main

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

func TestExtractTarGzRestoresContentsAndExecutableBit(t *testing.T) {
	source := t.TempDir()
	binary := filepath.Join(source, "tyrano-parser_linux_amd64")
	readme := filepath.Join(source, "README.md")
	writeFile(t, binary, "binary body")
	writeFile(t, readme, "docs body")
	archive := filepath.Join(source, "release.tar.gz")
	entries := archiveEntries(binary, targetPlatform{goos: "linux", goarch: "amd64"}, []string{readme})
	if err := createArchive(archive, entries); err != nil {
		t.Fatal(err)
	}

	dest := t.TempDir()
	if err := extractTarGz(archive, dest); err != nil {
		t.Fatalf("extractTarGz() error = %v", err)
	}

	body, err := os.ReadFile(filepath.Join(dest, "tyrano-parser"))
	if err != nil {
		t.Fatal(err)
	}
	if string(body) != "binary body" {
		t.Errorf("extracted binary = %q, want %q", body, "binary body")
	}
	info, err := os.Stat(filepath.Join(dest, "tyrano-parser"))
	if err != nil {
		t.Fatal(err)
	}
	if info.Mode()&0o111 == 0 {
		t.Errorf("extracted binary mode = %v, want the executable bit", info.Mode())
	}
	if _, err := os.Stat(filepath.Join(dest, "README.md")); err != nil {
		t.Errorf("README.md was not extracted: %v", err)
	}
}

func TestExtractTarGzRejectsPathTraversal(t *testing.T) {
	source := t.TempDir()
	escaping := filepath.Join(source, "payload")
	writeFile(t, escaping, "payload")
	archive := filepath.Join(source, "evil.tar.gz")
	if err := createArchive(archive, []archiveEntry{
		{name: "../escaped", path: escaping, mode: 0o644},
	}); err != nil {
		t.Fatal(err)
	}

	if err := extractTarGz(archive, t.TempDir()); err == nil {
		t.Fatal("extractTarGz() error = nil, want a rejection of an entry escaping the destination")
	}
}

func TestVerifyHostArchiveAcceptsMatchingVersion(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("the host archive is a zip on windows; the runnable check is unix-only")
	}
	dir := t.TempDir()
	host := targetPlatform{goos: runtime.GOOS, goarch: runtime.GOARCH}
	stubBinary(t, dir, host, "v1.2.3")

	if err := verifyHostArchive(dir, "v1.2.3", host); err != nil {
		t.Fatalf("verifyHostArchive() error = %v", err)
	}
}

func TestVerifyHostArchiveRejectsVersionMismatch(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("the host archive is a zip on windows; the runnable check is unix-only")
	}
	dir := t.TempDir()
	host := targetPlatform{goos: runtime.GOOS, goarch: runtime.GOARCH}
	stubBinary(t, dir, host, "v0.0.1")

	if err := verifyHostArchive(dir, "v1.2.3", host); err == nil {
		t.Fatal("verifyHostArchive() error = nil, want a mismatch between the archive and the release version")
	}
}

func TestVerifyHostArchiveReportsMissingArchive(t *testing.T) {
	host := targetPlatform{goos: runtime.GOOS, goarch: runtime.GOARCH}
	if err := verifyHostArchive(t.TempDir(), "v1.2.3", host); err == nil {
		t.Fatal("verifyHostArchive() error = nil, want a report that the archive is missing")
	}
}

// stubBinary packages a shell script that impersonates the CLI, so the
// verification path can be exercised without a cross-compile.
func stubBinary(t *testing.T, dir string, host targetPlatform, reported string) {
	t.Helper()
	script := filepath.Join(t.TempDir(), "stub")
	writeFile(t, script, "#!/bin/sh\necho \"tyrano-parser "+reported+"\"\n")
	if err := os.Chmod(script, 0o755); err != nil {
		t.Fatal(err)
	}
	archive := filepath.Join(dir, archiveName(host, "v1.2.3"))
	if err := createArchive(archive, archiveEntries(script, host, nil)); err != nil {
		t.Fatal(err)
	}
}
