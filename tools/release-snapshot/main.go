package main

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
)

const distDir = "dist"

type targetPlatform struct {
	goos   string
	goarch string
}

type buildRunner func(targetPlatform, string) error

func releasePlatforms() []targetPlatform {
	return []targetPlatform{
		{goos: "darwin", goarch: "amd64"},
		{goos: "darwin", goarch: "arm64"},
		{goos: "linux", goarch: "amd64"},
		{goos: "linux", goarch: "arm64"},
		{goos: "windows", goarch: "amd64"},
	}
}

func artifactName(target targetPlatform) string {
	name := fmt.Sprintf("tyrano-parser_%s_%s", target.goos, target.goarch)
	if target.goos == "windows" {
		name += ".exe"
	}
	return name
}

func buildArtifacts(dir string, targets []targetPlatform, runner buildRunner) ([]string, error) {
	artifacts := make([]string, 0, len(targets))
	for _, target := range targets {
		output := filepath.Join(dir, artifactName(target))
		fmt.Printf("==> build %s/%s: %s\n", target.goos, target.goarch, output)
		if err := runner(target, output); err != nil {
			return nil, err
		}
		artifacts = append(artifacts, output)
	}
	return artifacts, nil
}

func runGoBuild(target targetPlatform, output string) error {
	cmd := exec.Command("go", "build", "-trimpath", "-o", output, "./cmd/tyrano-parser")
	cmd.Env = append(os.Environ(), "GOOS="+target.goos, "GOARCH="+target.goarch)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func writeChecksums(dir string, artifacts []string) error {
	lines := make([]string, 0, len(artifacts))
	for _, artifact := range artifacts {
		data, err := os.ReadFile(artifact)
		if err != nil {
			return err
		}
		sum := sha256.Sum256(data)
		lines = append(lines, fmt.Sprintf("%s  %s", hex.EncodeToString(sum[:]), filepath.Base(artifact)))
	}
	sort.Strings(lines)
	content := strings.Join(lines, "\n") + "\n"
	return os.WriteFile(filepath.Join(dir, "checksums.txt"), []byte(content), 0o644)
}

func prepareDist(dir string) error {
	if err := os.RemoveAll(dir); err != nil {
		return err
	}
	return os.MkdirAll(dir, 0o755)
}

func main() {
	if err := prepareDist(distDir); err != nil {
		fmt.Fprintf(os.Stderr, "prepare dist failed: %v\n", err)
		os.Exit(1)
	}
	artifacts, err := buildArtifacts(distDir, releasePlatforms(), runGoBuild)
	if err != nil {
		fmt.Fprintf(os.Stderr, "release snapshot failed: %v\n", err)
		os.Exit(1)
	}
	if err := writeChecksums(distDir, artifacts); err != nil {
		fmt.Fprintf(os.Stderr, "checksum failed: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("wrote %d artifacts and checksums to %s\n", len(artifacts), distDir)
}
