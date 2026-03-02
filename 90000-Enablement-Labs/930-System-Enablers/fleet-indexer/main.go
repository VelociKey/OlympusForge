package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type Module struct {
	Name string
	Path string
}

func main() {
	var rootFlag string
	flag.StringVar(&rootFlag, "root", ".", "Root directory to index")
	flag.Parse()

	root, err := filepath.Abs(rootFlag)
	if err != nil {
		fmt.Println("Error deriving absolute path:", err)
		return
	}

	var modules []Module

	err = filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil || !info.IsDir() {
			return nil
		}
		if info.Name() == ".git" || info.Name() == "node_modules" {
			return filepath.SkipDir
		}

		modPath := filepath.Join(path, "go.mod")
		if _, err := os.Stat(modPath); err == nil {
			modName := extractModuleName(modPath)
			if modName != "" {
				rel, _ := filepath.Rel(root, path)
				modules = append(modules, Module{
					Name: modName,
					Path: "./" + filepath.ToSlash(rel),
				})
			}
		}
		return nil
	})

	if err != nil {
		fmt.Println("Walk error:", err)
		return
	}

	var sb strings.Builder
	sb.WriteString("# Olympus Fleet Go Module Registry (jeBNF)\n")
	sb.WriteString(fmt.Sprintf("indexed_at = %q\n\n", time.Now().Format(time.RFC3339)))
	for _, m := range modules {
		sb.WriteString(fmt.Sprintf("Module %q %q\n", m.Name, m.Path))
	}

	// Write MODULE_INDEX.jebnf to conductor/symbols/
	symbolsDir := filepath.Join(root, "conductor", "symbols")
	os.MkdirAll(symbolsDir, 0755)
	moduleIdxPath := filepath.Join(symbolsDir, "MODULE_INDEX.jebnf")
	err = os.WriteFile(moduleIdxPath, []byte(sb.String()), 0644)
	if err != nil {
		fmt.Println("Error writing MODULE_INDEX.jebnf:", err)
	} else {
		fmt.Printf("Generated %s\n", moduleIdxPath)
	}

	// Write FLEET_GO_DNA.jebnf to root for legacy compatibility
	dnaPath := filepath.Join(root, "FLEET_GO_DNA.jebnf")
	os.WriteFile(dnaPath, []byte(sb.String()), 0644)
	fmt.Printf("Generated %s\n", dnaPath)
}

func extractModuleName(modPath string) string {
	f, err := os.Open(modPath)
	if err != nil {
		return ""
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if strings.HasPrefix(line, "module ") {
			parts := strings.Fields(line)
			if len(parts) >= 2 {
				return parts[1]
			}
		}
	}
	return ""
}
