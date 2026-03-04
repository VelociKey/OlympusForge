package main

import (
	"flag"
	"fmt"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

type renameRule struct {
	Old string
	New string
}

func main() {
	rootDir := flag.String("root", ".", "Root directory to start refinement")
	renameMap := flag.String("map", "", "Comma-separated old:new rename rules")
	dryRun := flag.Bool("dry-run", false, "Show changes without applying them")
	flag.Parse()

	if *renameMap == "" {
		log.Fatal("Error: -map is required (e.g., 'P0000-pkg:P0100-Identity,000-Tools:930-System-Enablers')")
	}

	rules := parseRules(*renameMap)

	fmt.Printf("🚀 Starting Fleet Taxonomy Refinement in: %s\n", *rootDir)
	if *dryRun {
		fmt.Println("⚠️  DRY RUN MODE: No changes will be applied.")
	}

	// 1. Rename Directories (Bottom-up to avoid path invalidation)
	err := refineDirectories(*rootDir, rules, *dryRun)
	if err != nil {
		log.Fatalf("Fatal: Directory refinement failed: %v", err)
	}

	// 2. Update File Contents (Imports/Paths)
	err = refineContents(*rootDir, rules, *dryRun)
	if err != nil {
		log.Fatalf("Fatal: Content refinement failed: %v", err)
	}

	fmt.Println("✅ Fleet Taxonomy Refinement Complete.")
}

func parseRules(raw string) []renameRule {
	var rules []renameRule
	parts := strings.Split(raw, ",")
	for _, p := range parts {
		kv := strings.Split(p, ":")
		if len(kv) == 2 {
			rules = append(rules, renameRule{Old: kv[0], New: kv[1]})
		}
	}
	return rules
}

func refineDirectories(root string, rules []renameRule, dryRun bool) error {
	var dirsToRename []string

	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			name := d.Name()
			for _, rule := range rules {
				if name == rule.Old {
					dirsToRename = append(dirsToRename, path)
					break
				}
			}
		}
		return nil
	})

	if err != nil {
		return err
	}

	// Sort by length descending to process deepest directories first
	sort.Slice(dirsToRename, func(i, j int) bool {
		return len(dirsToRename[i]) > len(dirsToRename[j])
	})

	for _, oldPath := range dirsToRename {
		base := filepath.Base(oldPath)
		var newName string
		for _, rule := range rules {
			if base == rule.Old {
				newName = rule.New
				break
			}
		}

		newPath := filepath.Join(filepath.Dir(oldPath), newName)
		fmt.Printf("  [DIR] %s -> %s\n", oldPath, newPath)

		if !dryRun {
			err := os.Rename(oldPath, newPath)
			if err != nil {
				return fmt.Errorf("failed to rename %s: %w", oldPath, err)
			}
		}
	}

	return nil
}

func refineContents(root string, rules []renameRule, dryRun bool) error {
	return filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			if d.Name() == ".git" || d.Name() == "node_modules" {
				return filepath.SkipDir
			}
			return nil
		}

		// Only process text files (Go, Dart, JEBNF, MD, Proto, YAML, JSON)
		ext := filepath.Ext(path)
		switch ext {
		case ".go", ".dart", ".jebnf", ".md", ".proto", ".yaml", ".yml", ".json", ".mod", ".work":
			return processFile(path, rules, dryRun)
		}

		return nil
	})
}

func processFile(path string, rules []renameRule, dryRun bool) error {
	content, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	s := string(content)
	modified := false

	for _, rule := range rules {
		if strings.Contains(s, rule.Old) {
			s = strings.ReplaceAll(s, rule.Old, rule.New)
			modified = true
		}
	}

	if modified {
		fmt.Printf("  [FILE] %s updated\n", path)
		if !dryRun {
			return os.WriteFile(path, []byte(s), 0644)
		}
	}

	return nil
}
