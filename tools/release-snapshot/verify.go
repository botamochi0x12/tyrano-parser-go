package main

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// extractTarGz unpacks archive into dest, refusing any entry that would write
// outside of it.
func extractTarGz(archive, dest string) error {
	file, err := os.Open(archive)
	if err != nil {
		return err
	}
	defer file.Close()
	gz, err := gzip.NewReader(file)
	if err != nil {
		return err
	}
	defer gz.Close()

	reader := tar.NewReader(gz)
	for {
		header, err := reader.Next()
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return err
		}
		target := filepath.Join(dest, filepath.Clean(header.Name))
		if !strings.HasPrefix(target, filepath.Clean(dest)+string(os.PathSeparator)) {
			return fmt.Errorf("archive entry %q escapes the destination directory", header.Name)
		}
		out, err := os.OpenFile(target, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, os.FileMode(header.Mode))
		if err != nil {
			return err
		}
		if _, err := io.Copy(out, reader); err != nil {
			out.Close()
			return err
		}
		if err := out.Close(); err != nil {
			return err
		}
	}
}

// verifyHostArchive proves the release is actually usable: it unpacks the
// archive built for this machine and asks the binary which version it is.
func verifyHostArchive(dir, version string, host targetPlatform) error {
	archive := filepath.Join(dir, archiveName(host, version))
	if _, err := os.Stat(archive); err != nil {
		return fmt.Errorf("host archive %s is missing: %w", filepath.Base(archive), err)
	}
	if host.goos == "windows" {
		fmt.Printf("==> skip runnable check: %s is a zip\n", filepath.Base(archive))
		return nil
	}

	unpacked, err := os.MkdirTemp("", "tyrano-parser-verify")
	if err != nil {
		return err
	}
	defer os.RemoveAll(unpacked)
	if err := extractTarGz(archive, unpacked); err != nil {
		return err
	}

	cmd := exec.Command(filepath.Join(unpacked, binaryName), "version")
	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("extracted binary failed to run: %w", err)
	}

	want := fmt.Sprintf("%s %s", binaryName, version)
	if got := strings.TrimSpace(out.String()); got != want {
		return fmt.Errorf("extracted binary reports %q, want %q", got, want)
	}
	fmt.Printf("==> verified %s reports %s\n", filepath.Base(archive), version)
	return nil
}
