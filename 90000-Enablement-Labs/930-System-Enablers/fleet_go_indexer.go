package main

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

type GoModule struct {
	Name string
	Path string
}

func main() {
	fmt.Println("Initializing Conflict-Aware Fleet Go Indexing...")
	modules := []GoModule{}
	seen := make(map[string]bool)
	
	workspaces, _ := filepath.Glob("Olympus*")
	for _, ws := range workspaces {
		filepath.Walk(ws, func(path string, info os.FileInfo, err error) error {
			if err != nil || info.IsDir() || info.Name() != "go.mod" {
				return nil
			}
			
			content, _ := os.ReadFile(path)
			re := regexp.MustCompile(`module\s+([^\s\n\r]+)`)
			match := re.FindStringSubmatch(string(content))
			
			if len(match) > 1 {
				modName := match[1]
				if seen[modName] {
					fmt.Printf("Warning: Duplicate module %q found at %q. Skipping to prevent conflicts.\n", modName, path)
					return nil
				}
				seen[modName] = true
				modules = append(modules, GoModule{
					Name: modName,
					Path: filepath.Dir(path),
				})
			}
			return nil
		})
	}

	// Generate FLEET_GO_DNA.jebnf
	var sb strings.Builder
	sb.WriteString("# Olympus Fleet Go Module Registry (jeBNF)\n")
	sb.WriteString(fmt.Sprintf("indexed_at = %q\n\n", time.Now().Format(time.RFC3339)))
	for _, m := range modules {
		sb.WriteString(fmt.Sprintf("Module %q %q\n", m.Name, m.Path))
	}
	os.WriteFile("FLEET_GO_DNA.jebnf", []byte(sb.String()), 0644)

	// Update go.work files
	targets := []string{"go.work", "Olympus2/go.work", "OlympusForge/go.work"}
	for _, t := range targets {
		updateGoWork(t, modules)
	}
}

func updateGoWork(target string, modules []GoModule) {
	if _, err := os.Stat(target); err != nil { return }
	
	// Get target's absolute directory
	targetDir, _ := filepath.Abs(filepath.Dir(target))
	
	var sb strings.Builder
	sb.WriteString("go 1.25.7\n\nuse (\n")
	
	for _, m := range modules {
		modAbsPath, _ := filepath.Abs(m.Path)
		relPath, err := filepath.Rel(targetDir, modAbsPath)
		if err != nil { continue }
		
		relPath = strings.ReplaceAll(relPath, "\\", "/")
		if !strings.HasPrefix(relPath, ".") {
			relPath = "./" + relPath
		}
		sb.WriteString(fmt.Sprintf("\t%s\n", relPath))
	}
	
	sb.WriteString(")\n")
	os.WriteFile(target, []byte(sb.String()), 0644)
	fmt.Printf("Successfully synchronized %s\n", target)
}
