package main

import (
	"os"
	"path/filepath"
	"reflect"
	"slices"
	"strings"
	"testing"
)

func TestReleasePlatforms(t *testing.T) {
	got := releasePlatforms()
	want := []targetPlatform{
		{goos: "darwin", goarch: "amd64"},
		{goos: "darwin", goarch: "arm64"},
		{goos: "linux", goarch: "amd64"},
		{goos: "linux", goarch: "arm64"},
		{goos: "windows", goarch: "amd64"},
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("releasePlatforms() = %#v, want %#v", got, want)
	}
}

func TestArtifactName(t *testing.T) {
	tests := []struct {
		name string
		in   targetPlatform
		want string
	}{
		{name: "linux", in: targetPlatform{goos: "linux", goarch: "amd64"}, want: "tyrano-parser_linux_amd64"},
		{name: "windows", in: targetPlatform{goos: "windows", goarch: "amd64"}, want: "tyrano-parser_windows_amd64.exe"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := artifactName(tt.in); got != tt.want {
				t.Fatalf("artifactName() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestWriteChecksums(t *testing.T) {
	dir := t.TempDir()
	first := filepath.Join(dir, "tyrano-parser_linux_amd64")
	second := filepath.Join(dir, "tyrano-parser_windows_amd64.exe")
	if err := os.WriteFile(first, []byte("linux"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(second, []byte("windows"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := writeChecksums(dir, []string{second, first}); err != nil {
		t.Fatalf("writeChecksums() error = %v", err)
	}
	data, err := os.ReadFile(filepath.Join(dir, "checksums.txt"))
	if err != nil {
		t.Fatal(err)
	}
	lines := strings.Split(strings.TrimSpace(string(data)), "\n")
	if !slices.IsSorted(lines) {
		t.Fatalf("checksum lines are not sorted: %#v", lines)
	}
	for _, wantName := range []string{"tyrano-parser_linux_amd64", "tyrano-parser_windows_amd64.exe"} {
		if !strings.Contains(string(data), "  "+wantName+"\n") {
			t.Fatalf("checksums missing %q:\n%s", wantName, data)
		}
	}
}

func TestBuildArtifactsStopsOnFailure(t *testing.T) {
	dir := t.TempDir()
	targets := []targetPlatform{
		{goos: "linux", goarch: "amd64"},
		{goos: "windows", goarch: "amd64"},
	}
	var ran []targetPlatform
	_, err := buildArtifacts(dir, targets, func(target targetPlatform, output string) error {
		ran = append(ran, target)
		return os.ErrPermission
	})
	if err == nil {
		t.Fatal("buildArtifacts() error = nil, want failure")
	}
	if len(ran) != 1 {
		t.Fatalf("ran %d builds, want 1", len(ran))
	}
}
