package main

import (
	"bytes"
	"io/fs"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"unicode"
)

func main() {
	slog.Info("Sovereign Bazel Cleaner v2: Global Repository Synchronization")

	wd, _ := os.Getwd()
	rootPath := wd
	for {
		if _, err := os.Stat(filepath.Join(rootPath, "go.work")); err == nil {
			break
		}
		parent := filepath.Dir(rootPath)
		if parent == rootPath {
			rootPath = wd
			break
		}
		rootPath = parent
	}

	slog.Info("Root discovered", "path", rootPath)

	targets := []string{"00SDLC", "50RNDF", "60PROX"}
	
	for _, target := range targets {
		absRoot := filepath.Join(rootPath, target)
		if _, err := os.Stat(absRoot); err != nil {
			continue
		}

		err := filepath.WalkDir(absRoot, func(path string, d fs.DirEntry, err error) error {
			if err != nil {
				return err
			}
			if d.IsDir() {
				name := d.Name()
				if strings.HasPrefix(name, "C") && len(name) == 5 {
					return filepath.SkipDir
				}
				// Skip toolchains and bazel internal symlinks
				if name == "81000-Toolchain-External" || name == "81200-Logic-Libraries" || name == ".git" || strings.HasPrefix(name, "bazel-") {
					return filepath.SkipDir
				}
				return nil
			}

			if d.Name() == "BUILD.bazel" {
				return cleanFile(path)
			}
			return nil
		})
		if err != nil {
			slog.Error("Walk error", "target", target, "error", err)
		}
	}

	slog.Info("Bazel Clean Pulse Complete.")
}

func cleanFile(path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	var buf bytes.Buffer
	for _, r := range string(data) {
		if r > 127 {
			continue 
		}
		if unicode.IsPrint(r) || unicode.IsSpace(r) {
			buf.WriteRune(r)
		}
	}

	content := buf.String()
	original := content

	// Aggressive repository synchronization
	content = strings.ReplaceAll(content, "@rules_go//", "@io_bazel_rules_go//")
	content = strings.ReplaceAll(content, "@gazelle//", "@bazel_gazelle//")
	
	// Handle cases where gazelle might be loaded as a string without the // suffix
	content = strings.ReplaceAll(content, "\"@gazelle", "\"@bazel_gazelle")

	content = strings.ReplaceAll(content, "60000-Information-Storage/90200-Logic-Libraries", "90200-Logic-Libraries")

	if content != original || len(data) != buf.Len() {
		slog.Info("Cleaning file", "path", path)
		return os.WriteFile(path, []byte(content), 0644)
	}

	return nil
}
