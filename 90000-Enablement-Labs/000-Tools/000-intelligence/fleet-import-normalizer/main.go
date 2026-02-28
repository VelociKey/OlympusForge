package main

import (
	"flag"
	"fmt"
	"io/fs"
	"log/slog"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
)

var importRegex = regexp.MustCompile(`"([^"]+)"`)

func main() {
	root := flag.String("root", ".", "Root directory to traverse")
	prefix := flag.String("strip", "", "Prefix to strip (e.g., )")
	list := flag.Bool("list", false, "Phase 1: List unique web prefixes found in .go files")
	flag.Parse()

	if *list {
		runPhase1(*root)
		return
	}

	if *prefix != "" {
		runPhase2(*root, *prefix)
		return
	}

	fmt.Println("Sovereign Fleet Import Normalizer")
	fmt.Println("Usage:")
	fmt.Println("  Phase 1: fleet-import-normalizer -list -root .")
	fmt.Println("  Phase 2: fleet-import-normalizer -strip <prefix> -root .")
	os.Exit(1)
}

func runPhase1(root string) {
	slog.Info("🔍 Phase 1: Identifying unique web prefixes", "root", root)
	prefixes := make(map[string]bool)
	
	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil { return nil }
		if d.IsDir() {
			if d.Name() == ".git" || d.Name() == "node_modules" || d.Name() == "gen" {
				return filepath.SkipDir
			}
			return nil
		}
		if !strings.HasSuffix(d.Name(), ".go") { return nil }

		content, err := os.ReadFile(path)
		if err != nil { return nil }

		matches := importRegex.FindAllStringSubmatch(string(content), -1)
		for _, m := range matches {
			imp := m[1]
			parts := strings.Split(imp, "/")
			if len(parts) > 0 && strings.Contains(parts[0], ".") {
				if parts[0] == "github.com" && len(parts) > 1 {
					prefixes[parts[0]+"/"+parts[1]+"/"] = true
				} else {
					prefixes[parts[0]+"/"] = true
				}
			}
		}
		return nil
	})

	if err != nil {
		slog.Error("Traversal failed", "error", err)
		return
	}

	var sorted []string
	for p := range prefixes {
		sorted = append(sorted, p)
	}
	sort.Strings(sorted)

	fmt.Println("\n--- Unique Web Prefixes ---")
	for _, p := range sorted {
		fmt.Println(p)
	}
}

func runPhase2(root string, prefix string) {
	slog.Info("🧹 Phase 2: Stripping prefix", "prefix", prefix, "root", root)
	updatedCount := 0

	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil { return nil }
		if d.IsDir() {
			if d.Name() == ".git" || d.Name() == "node_modules" {
				return filepath.SkipDir
			}
			return nil
		}

		name := d.Name()
		ext := filepath.Ext(path)
		if ext != ".go" && name != "go.mod" && name != "go.work" {
			return nil
		}

		content, err := os.ReadFile(path)
		if err != nil { return nil }

		sContent := string(content)
		if strings.Contains(sContent, prefix) {
			newContent := strings.ReplaceAll(sContent, prefix, "")
			
			if newContent != sContent {
				err = os.WriteFile(path, []byte(newContent), 0644)
				if err != nil {
					slog.Error("Failed to write file", "path", path, "error", err)
				} else {
					updatedCount++
					slog.Info("Normalized", "path", path)
				}
			}
		}
		return nil
	})

	if err != nil {
		slog.Error("Normalization failed", "error", err)
	}
	slog.Info("✅ Phase 2 Complete", "filesUpdated", updatedCount)
}
