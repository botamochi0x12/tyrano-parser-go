package main

import (
	"archive/tar"
	"archive/zip"
	"compress/gzip"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"slices"
	"testing"
	"time"
)

func TestArchiveName(t *testing.T) {
	tests := []struct {
		name    string
		target  targetPlatform
		version string
		want    string
	}{
		{
			name:    "unix targets ship a tarball",
			target:  targetPlatform{goos: "linux", goarch: "amd64"},
			version: "v1.2.3",
			want:    "tyrano-parser_v1.2.3_linux_amd64.tar.gz",
		},
		{
			name:    "windows ships a zip",
			target:  targetPlatform{goos: "windows", goarch: "amd64"},
			version: "v1.2.3",
			want:    "tyrano-parser_v1.2.3_windows_amd64.zip",
		},
		{
			name:    "dev snapshots are labelled too",
			target:  targetPlatform{goos: "darwin", goarch: "arm64"},
			version: devVersion,
			want:    "tyrano-parser_dev_darwin_arm64.tar.gz",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := archiveName(tt.target, tt.version); got != tt.want {
				t.Fatalf("archiveName() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestArchiveEntriesCarryBinaryAndAvailableDocs(t *testing.T) {
	dir := t.TempDir()
	binary := filepath.Join(dir, "tyrano-parser_linux_amd64")
	readme := filepath.Join(dir, "README.md")
	writeFile(t, binary, "binary")
	writeFile(t, readme, "docs")
	missing := filepath.Join(dir, "LICENSE")

	entries := archiveEntries(binary, targetPlatform{goos: "linux", goarch: "amd64"}, []string{readme, missing})

	var names []string
	for _, entry := range entries {
		names = append(names, entry.name)
	}
	want := []string{"tyrano-parser", "README.md"}
	if !slices.Equal(names, want) {
		t.Fatalf("archiveEntries() names = %#v, want %#v (missing docs are skipped)", names, want)
	}
	if entries[0].mode&0o111 == 0 {
		t.Fatalf("binary mode = %v, want the executable bit set", entries[0].mode)
	}
}

func TestArchiveEntriesUseWindowsBinaryName(t *testing.T) {
	dir := t.TempDir()
	binary := filepath.Join(dir, "tyrano-parser_windows_amd64.exe")
	writeFile(t, binary, "binary")

	entries := archiveEntries(binary, targetPlatform{goos: "windows", goarch: "amd64"}, nil)

	if len(entries) != 1 || entries[0].name != "tyrano-parser.exe" {
		t.Fatalf("archiveEntries() = %#v, want a single tyrano-parser.exe entry", entries)
	}
}

func TestCreateArchiveWritesTarGz(t *testing.T) {
	dir := t.TempDir()
	binary := filepath.Join(dir, "tyrano-parser_linux_amd64")
	readme := filepath.Join(dir, "README.md")
	writeFile(t, binary, "binary body")
	writeFile(t, readme, "docs body")
	dest := filepath.Join(dir, "tyrano-parser_v1.2.3_linux_amd64.tar.gz")

	entries := archiveEntries(binary, targetPlatform{goos: "linux", goarch: "amd64"}, []string{readme})
	if err := createArchive(dest, entries); err != nil {
		t.Fatalf("createArchive() error = %v", err)
	}

	got := readTarGz(t, dest)
	if got["tyrano-parser"] != "binary body" {
		t.Errorf("binary content = %q, want %q", got["tyrano-parser"], "binary body")
	}
	if got["README.md"] != "docs body" {
		t.Errorf("README content = %q, want %q", got["README.md"], "docs body")
	}
}

func TestCreateArchiveTarGzKeepsBinaryExecutable(t *testing.T) {
	dir := t.TempDir()
	binary := filepath.Join(dir, "tyrano-parser_linux_amd64")
	writeFile(t, binary, "binary body")
	dest := filepath.Join(dir, "tyrano-parser_v1.2.3_linux_amd64.tar.gz")

	entries := archiveEntries(binary, targetPlatform{goos: "linux", goarch: "amd64"}, nil)
	if err := createArchive(dest, entries); err != nil {
		t.Fatalf("createArchive() error = %v", err)
	}

	if mode := tarEntryMode(t, dest, "tyrano-parser"); mode&0o111 == 0 {
		t.Fatalf("archived binary mode = %v, want the executable bit preserved", mode)
	}
}

func TestCreateArchiveStampsModificationTimes(t *testing.T) {
	dir := t.TempDir()
	binary := filepath.Join(dir, "tyrano-parser_linux_amd64")
	writeFile(t, binary, "binary body")
	dest := filepath.Join(dir, "tyrano-parser_v1.2.3_linux_amd64.tar.gz")

	entries := archiveEntries(binary, targetPlatform{goos: "linux", goarch: "amd64"}, nil)
	if err := createArchive(dest, entries); err != nil {
		t.Fatalf("createArchive() error = %v", err)
	}

	info, err := os.Stat(binary)
	if err != nil {
		t.Fatal(err)
	}
	want := info.ModTime()
	got := tarEntryModTime(t, dest, "tyrano-parser")
	if diff := got.Sub(want); diff > time.Second || diff < -time.Second {
		t.Fatalf("archived binary mtime = %s, want the source timestamp %s", got, want)
	}
}

func TestCreateArchiveWritesZip(t *testing.T) {
	dir := t.TempDir()
	binary := filepath.Join(dir, "tyrano-parser_windows_amd64.exe")
	readme := filepath.Join(dir, "README.md")
	writeFile(t, binary, "binary body")
	writeFile(t, readme, "docs body")
	dest := filepath.Join(dir, "tyrano-parser_v1.2.3_windows_amd64.zip")

	entries := archiveEntries(binary, targetPlatform{goos: "windows", goarch: "amd64"}, []string{readme})
	if err := createArchive(dest, entries); err != nil {
		t.Fatalf("createArchive() error = %v", err)
	}

	got := readZip(t, dest)
	if got["tyrano-parser.exe"] != "binary body" {
		t.Errorf("binary content = %q, want %q", got["tyrano-parser.exe"], "binary body")
	}
	if got["README.md"] != "docs body" {
		t.Errorf("README content = %q, want %q", got["README.md"], "docs body")
	}
}

func TestCreateArchiveRejectsUnknownExtension(t *testing.T) {
	dir := t.TempDir()
	binary := filepath.Join(dir, "tyrano-parser_linux_amd64")
	writeFile(t, binary, "binary")
	entries := archiveEntries(binary, targetPlatform{goos: "linux", goarch: "amd64"}, nil)

	if err := createArchive(filepath.Join(dir, "artifact.rar"), entries); err == nil {
		t.Fatal("createArchive() error = nil, want a rejection for an unsupported format")
	}
}

func TestPackageArtifactsWritesOneArchivePerTarget(t *testing.T) {
	stage := t.TempDir()
	dist := t.TempDir()
	linuxTarget := targetPlatform{goos: "linux", goarch: "amd64"}
	windowsTarget := targetPlatform{goos: "windows", goarch: "amd64"}
	linuxBinary := filepath.Join(stage, artifactName(linuxTarget))
	windowsBinary := filepath.Join(stage, artifactName(windowsTarget))
	writeFile(t, linuxBinary, "linux body")
	writeFile(t, windowsBinary, "windows body")

	archives, err := packageArtifacts(dist, []builtArtifact{
		{target: linuxTarget, path: linuxBinary},
		{target: windowsTarget, path: windowsBinary},
	}, "v1.2.3", nil)
	if err != nil {
		t.Fatalf("packageArtifacts() error = %v", err)
	}

	var names []string
	for _, archive := range archives {
		names = append(names, filepath.Base(archive))
	}
	want := []string{
		"tyrano-parser_v1.2.3_linux_amd64.tar.gz",
		"tyrano-parser_v1.2.3_windows_amd64.zip",
	}
	if !slices.Equal(names, want) {
		t.Fatalf("packageArtifacts() = %#v, want %#v", names, want)
	}
	if got := readTarGz(t, archives[0])["tyrano-parser"]; got != "linux body" {
		t.Fatalf("archived linux binary = %q, want %q", got, "linux body")
	}
}

func TestReleaseDocsShipLicenseAndReadme(t *testing.T) {
	want := []string{"README.md", "LICENSE"}
	if got := releaseDocs(); !slices.Equal(got, want) {
		t.Fatalf("releaseDocs() = %#v, want %#v", got, want)
	}
}

func writeFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

func readTarGz(t *testing.T, path string) map[string]string {
	t.Helper()
	file, err := os.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()
	gz, err := gzip.NewReader(file)
	if err != nil {
		t.Fatal(err)
	}
	defer gz.Close()
	contents := map[string]string{}
	reader := tar.NewReader(gz)
	for {
		header, err := reader.Next()
		if err == io.EOF {
			return contents
		}
		if err != nil {
			t.Fatal(err)
		}
		body, err := io.ReadAll(reader)
		if err != nil {
			t.Fatal(err)
		}
		contents[header.Name] = string(body)
	}
}

func tarEntryMode(t *testing.T, path, name string) fs.FileMode {
	t.Helper()
	return fs.FileMode(tarEntryHeader(t, path, name).Mode)
}

func tarEntryModTime(t *testing.T, path, name string) time.Time {
	t.Helper()
	return tarEntryHeader(t, path, name).ModTime
}

func tarEntryHeader(t *testing.T, path, name string) *tar.Header {
	t.Helper()
	file, err := os.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()
	gz, err := gzip.NewReader(file)
	if err != nil {
		t.Fatal(err)
	}
	defer gz.Close()
	reader := tar.NewReader(gz)
	for {
		header, err := reader.Next()
		if err == io.EOF {
			t.Fatalf("archive %s has no entry named %q", path, name)
		}
		if err != nil {
			t.Fatal(err)
		}
		if header.Name == name {
			return header
		}
	}
}

func readZip(t *testing.T, path string) map[string]string {
	t.Helper()
	reader, err := zip.OpenReader(path)
	if err != nil {
		t.Fatal(err)
	}
	defer reader.Close()
	contents := map[string]string{}
	for _, file := range reader.File {
		rc, err := file.Open()
		if err != nil {
			t.Fatal(err)
		}
		body, err := io.ReadAll(rc)
		rc.Close()
		if err != nil {
			t.Fatal(err)
		}
		contents[file.Name] = string(body)
	}
	return contents
}
