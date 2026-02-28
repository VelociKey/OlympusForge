package main

import (
	"flag"
	"fmt"
	"io/fs"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

func main() {
	root := flag.String("root", "", "Family root to process (e.g., 00SDLC)")
	flag.Parse()

	if *root == "" {
		fmt.Println("Usage: fleet-remod -root <family_dir>")
		os.Exit(1)
	}

	absRoot, err := filepath.Abs(*root)
	if err != nil {
		slog.Error("Failed to get absolute path", "error", err)
		os.Exit(1)
	}

	slog.Info("🚀 Starting Fleet Remodelling", "root", absRoot)

	var modules []string

	err = filepath.WalkDir(absRoot, func(path string, d fs.DirEntry, err error) error {
		if err != nil { return nil }
		if !d.IsDir() { return nil }

		if d.Name() == ".git" || d.Name() == "node_modules" || d.Name() == "gen" || d.Name() == "proto" {
			return filepath.SkipDir
		}

		hasGo := false
		entries, _ := os.ReadDir(path)
		for _, e := range entries {
			if !e.IsDir() && strings.HasSuffix(e.Name(), ".go") {
				hasGo = true
				break
			}
		}

		if hasGo {
			slog.Info("Initializing module", "dir", path)
			os.Remove(filepath.Join(path, "go.mod"))
			os.Remove(filepath.Join(path, "go.sum"))

			moduleName := filepath.Base(path)
			cmd := exec.Command("go", "mod", "init", moduleName)
			cmd.Dir = path
			if out, err := cmd.CombinedOutput(); err != nil {
				slog.Error("go mod init failed", "dir", path, "error", err, "output", string(out))
			} else {
				rel, _ := filepath.Rel(absRoot, path)
				modules = append(modules, "./"+filepath.ToSlash(rel))
			}
		}
		return nil
	})

	if err != nil {
		slog.Error("Traversal failed", "error", err)
		os.Exit(1)
	}

	if len(modules) > 0 {
		workPath := filepath.Join(absRoot, "go.work")
		var sb strings.Builder
		sb.WriteString("go 1.25.7\n\nuse (\n")
		for _, m := range modules {
			sb.WriteString(fmt.Sprintf("\t%s\n", m))
		}
		sb.WriteString(")\n")

		if err := os.WriteFile(workPath, []byte(sb.String()), 0644); err != nil {
			slog.Error("Failed to write go.work", "path", workPath, "error", err)
		} else {
			slog.Info("✅ Generated workspace", "path", workPath, "moduleCount", len(modules))
		}
	}

	slog.Info("✨ Remodelling Complete", "root", *root)
}
