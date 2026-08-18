package main

import (
	"archive/tar"
	"archive/zip"
	"compress/gzip"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

const binaryName = "tyrano-parser"

// archiveEntry is one file as it appears inside a release archive.
type archiveEntry struct {
	name string
	path string
	mode fs.FileMode
}

// archiveName is the published file name for one target, e.g.
// tyrano-parser_v1.2.3_linux_amd64.tar.gz.
func archiveName(target targetPlatform, version string) string {
	return fmt.Sprintf("%s_%s_%s_%s%s", binaryName, version, target.goos, target.goarch, archiveExt(target))
}

func archiveExt(target targetPlatform) string {
	if target.goos == "windows" {
		return ".zip"
	}
	return ".tar.gz"
}

// archiveEntries lays out the archive: the binary under its plain name so it
// can be dropped straight onto PATH, followed by whichever docs exist.
func archiveEntries(binaryPath string, target targetPlatform, docs []string) []archiveEntry {
	name := binaryName
	if target.goos == "windows" {
		name += ".exe"
	}
	entries := []archiveEntry{{name: name, path: binaryPath, mode: 0o755}}
	for _, doc := range docs {
		if _, err := os.Stat(doc); err != nil {
			continue
		}
		entries = append(entries, archiveEntry{name: filepath.Base(doc), path: doc, mode: 0o644})
	}
	return entries
}

func createArchive(dest string, entries []archiveEntry) error {
	switch {
	case strings.HasSuffix(dest, ".tar.gz"):
		return createTarGz(dest, entries)
	case strings.HasSuffix(dest, ".zip"):
		return createZip(dest, entries)
	default:
		return fmt.Errorf("unsupported archive format: %s", dest)
	}
}

func createTarGz(dest string, entries []archiveEntry) error {
	out, err := os.Create(dest)
	if err != nil {
		return err
	}
	defer out.Close()
	gz := gzip.NewWriter(out)
	defer gz.Close()
	writer := tar.NewWriter(gz)
	defer writer.Close()

	for _, entry := range entries {
		body, err := os.ReadFile(entry.path)
		if err != nil {
			return err
		}
		info, err := os.Stat(entry.path)
		if err != nil {
			return err
		}
		header := &tar.Header{
			Name:    entry.name,
			Mode:    int64(entry.mode),
			Size:    int64(len(body)),
			ModTime: info.ModTime(),
		}
		if err := writer.WriteHeader(header); err != nil {
			return err
		}
		if _, err := writer.Write(body); err != nil {
			return err
		}
	}
	return writer.Close()
}

func createZip(dest string, entries []archiveEntry) error {
	out, err := os.Create(dest)
	if err != nil {
		return err
	}
	defer out.Close()
	writer := zip.NewWriter(out)
	defer writer.Close()

	for _, entry := range entries {
		info, err := os.Stat(entry.path)
		if err != nil {
			return err
		}
		header := &zip.FileHeader{Name: entry.name, Method: zip.Deflate, Modified: info.ModTime()}
		header.SetMode(entry.mode)
		target, err := writer.CreateHeader(header)
		if err != nil {
			return err
		}
		body, err := os.Open(entry.path)
		if err != nil {
			return err
		}
		if _, err := io.Copy(target, body); err != nil {
			body.Close()
			return err
		}
		if err := body.Close(); err != nil {
			return err
		}
	}
	return writer.Close()
}
