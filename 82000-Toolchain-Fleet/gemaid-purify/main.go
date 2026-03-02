package main

import (
	"flag"
	"io/fs"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
)

var patterns = []string{
	"testdata",
	"test",
	"examples",
	"samples",
	".github",
	"node_modules",
}

var filePatterns = []string{
	"_test.go",
	"LICENSE",
	"PATENTS",
}

func main() {
	rootPtr := flag.String("root", ".", "Root directory to purify")
	dryRunPtr := flag.Bool("dry-run", false, "Log deletions without removing files")
	flag.Parse()

	root, _ := filepath.Abs(*rootPtr)
	slog.Info("Starting Purification", "root", root, "dryRun", *dryRunPtr)

	reclaimed := int64(0)
	count := 0

	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil
		}

		name := d.Name()
		shouldDelete := false

		if d.IsDir() {
			if name == "internal" {
				return filepath.SkipDir
			}
			for _, p := range patterns {
				if name == p {
					shouldDelete = true
					break
				}
			}
		} else {
			for _, p := range filePatterns {
				if strings.Contains(name, p) {
					shouldDelete = true
					break
				}
			}
		}

		if shouldDelete {
			info, err := d.Info()
			if err == nil {
				reclaimed += info.Size()
			}
			count++

			if *dryRunPtr {
				slog.Info("[DRY-RUN] Purging", "path", path)
			} else {
				err := os.RemoveAll(path)
				if err != nil {
					slog.Warn("Failed to delete", "path", path, "error", err)
				} else {
					slog.Debug("Deleted", "path", path)
				}
			}

			if d.IsDir() {
				return filepath.SkipDir
			}
		}

		return nil
	})

	if err != nil {
		slog.Error("Purification failed", "error", err)
		os.Exit(1)
	}

	slog.Info("Purification Complete", "deletedCount", count, "reclaimedBytes", reclaimed)
}
