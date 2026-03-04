package main

import (
	"bufio"
	"flag"
	"fmt"
	"io/fs"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
)

func main() {
	rootPath := flag.String("root", ".", "Root directory of the fleet")
	dryRun := flag.Bool("dry-run", false, "Scan only without making changes")
	flag.Parse()

	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	goWorkPath := filepath.Join(*rootPath, "go.work")
	targetVersion, err := extractGoVersion(goWorkPath)
	if err != nil {
		slog.Error("failed to extract authoritative go version from go.work", "path", goWorkPath, "error", err)
		os.Exit(1)
	}

	slog.Info("authoritative go version identified", "version", targetVersion)

	err = filepath.WalkDir(*rootPath, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		// Skip directories that are likely to contain third-party or research code not managed by this tool
		if d.IsDir() {
			if d.Name() == ".git" || d.Name() == "node_modules" || d.Name() == "ZC0400-Sovereign-Source" {
				return filepath.SkipDir
			}
			return nil
		}

		if d.Name() == "go.mod" && path != filepath.Join(*rootPath, "go.mod") {
			return syncGoMod(path, targetVersion, *dryRun)
		}

		return nil
	})

	if err != nil {
		slog.Error("failed to walk fleet directory", "error", err)
		os.Exit(1)
	}

	slog.Info("go version synchronization complete")
}

func extractGoVersion(path string) (string, error) {
	file, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if strings.HasPrefix(line, "go ") {
			return strings.TrimPrefix(line, "go "), nil
		}
	}
	return "", fmt.Errorf("no 'go' directive found in %s", path)
}

func syncGoMod(path string, targetVersion string, dryRun bool) error {
	content, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	lines := strings.Split(string(content), "\n")
	modified := false
	for i, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "go ") {
			currentVersion := strings.TrimPrefix(trimmed, "go ")
			if currentVersion != targetVersion {
				slog.Info("mismatch identified", "file", path, "current", currentVersion, "target", targetVersion)
				if !dryRun {
					lines[i] = "go " + targetVersion
					modified = true
				}
			}
			break // Only first 'go' directive matters
		}
	}

	if modified && !dryRun {
		err = os.WriteFile(path, []byte(strings.Join(lines, "\n")), 0644)
		if err != nil {
			return err
		}
		slog.Info("updated go.mod", "file", path)
	}

	return nil
}
