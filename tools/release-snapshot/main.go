package main

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"slices"
	"sort"
	"strings"
)

const (
	distDir = "dist"
	// stageDirName holds the raw binaries until they are packaged; only the
	// archives and their checksums are meant to be published.
	stageDirName = ".stage"
)

type targetPlatform struct {
	goos   string
	goarch string
}

// builtArtifact is a compiled binary together with the target it was built for.
type builtArtifact struct {
	target targetPlatform
	path   string
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

// releaseDocs travel inside every archive so a downloaded binary carries its
// licence and usage notes with it.
func releaseDocs() []string {
	return []string{"README.md", "LICENSE"}
}

func artifactName(target targetPlatform) string {
	name := fmt.Sprintf("tyrano-parser_%s_%s", target.goos, target.goarch)
	if target.goos == "windows" {
		name += ".exe"
	}
	return name
}

func buildArtifacts(dir string, targets []targetPlatform, runner buildRunner) ([]builtArtifact, error) {
	artifacts := make([]builtArtifact, 0, len(targets))
	for _, target := range targets {
		output := filepath.Join(dir, artifactName(target))
		fmt.Printf("==> build %s/%s: %s\n", target.goos, target.goarch, output)
		if err := runner(target, output); err != nil {
			return nil, err
		}
		artifacts = append(artifacts, builtArtifact{target: target, path: output})
	}
	return artifacts, nil
}

func packageArtifacts(dir string, built []builtArtifact, version string, docs []string) ([]string, error) {
	archives := make([]string, 0, len(built))
	for _, artifact := range built {
		dest := filepath.Join(dir, archiveName(artifact.target, version))
		fmt.Printf("==> package %s\n", dest)
		if err := createArchive(dest, archiveEntries(artifact.path, artifact.target, docs)); err != nil {
			return nil, err
		}
		archives = append(archives, dest)
	}
	return archives, nil
}

func goBuilder(version string) buildRunner {
	return func(target targetPlatform, output string) error {
		cmd := exec.Command("go", goBuildArgs(output, version)...)
		cmd.Env = append(os.Environ(),
			"GOOS="+target.goos,
			"GOARCH="+target.goarch,
			// Static binaries keep the download runnable on any host of that
			// platform, without a matching libc.
			"CGO_ENABLED=0",
		)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		return cmd.Run()
	}
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
	if err := run(os.Args[1:]); err != nil {
		fmt.Fprintf(os.Stderr, "release-snapshot failed: %v\n", err)
		os.Exit(1)
	}
}

func run(args []string) error {
	version, err := resolveVersion(args, os.Getenv)
	if err != nil {
		return err
	}
	fmt.Printf("==> release version: %s\n", version)

	stage := filepath.Join(distDir, stageDirName)
	if err := prepareDist(distDir); err != nil {
		return err
	}
	if err := os.MkdirAll(stage, 0o755); err != nil {
		return err
	}

	built, err := buildArtifacts(stage, releasePlatforms(), goBuilder(version))
	if err != nil {
		return err
	}
	archives, err := packageArtifacts(distDir, built, version, releaseDocs())
	if err != nil {
		return err
	}
	if err := os.RemoveAll(stage); err != nil {
		return err
	}
	if err := writeChecksums(distDir, archives); err != nil {
		return err
	}

	host := targetPlatform{goos: runtime.GOOS, goarch: runtime.GOARCH}
	if slices.Contains(releasePlatforms(), host) {
		if err := verifyHostArchive(distDir, version, host); err != nil {
			return err
		}
	} else {
		fmt.Printf("==> skip runnable check: %s/%s is not a release target\n", host.goos, host.goarch)
	}

	fmt.Printf("wrote %d archives and checksums for %s to %s\n", len(archives), version, distDir)
	return nil
}
