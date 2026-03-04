package main

import (
	"bufio"
	"fmt"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"strings"
)

// Dependency represents a Library or Tool from DEPENDENCIES.jebnf
type Dependency struct {
	Name      string
	LocalPath string
}

func main() {
	rootDir := "."
	if len(os.Args) > 1 {
		rootDir = os.Args[1]
	}

	fmt.Printf("🚀 Starting jeBNF-Driven Pub Overrider in: %s\n", rootDir)

	// Load overrides from the fleet registry
	overrides, err := loadRegistryOverrides("olympus.fleet/00SDLC/OlympusForge/C0100-Configuration-Registry/DEPENDENCIES.jebnf")
	if err != nil {
		log.Fatalf("Fatal: Could not load registry: %v", err)
	}

	err = filepath.WalkDir(rootDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if !d.IsDir() && d.Name() == "pubspec.yaml" {
			return overridePubspec(path, overrides)
		}

		if d.IsDir() && (d.Name() == ".git" || d.Name() == "node_modules" || d.Name() == ".pub-cache") {
			return filepath.SkipDir
		}

		return nil
	})

	if err != nil {
		log.Fatalf("Fatal: Overrider failed: %v", err)
	}

	fmt.Println("✅ jeBNF-Driven Pub Overrides Injected.")
}

func loadRegistryOverrides(path string) (map[string]string, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	overrides := make(map[string]string)
	scanner := bufio.NewScanner(file)

	var currentName string
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		if strings.HasPrefix(line, "Library") {
			currentName = strings.Trim(strings.TrimPrefix(line, "Library"), " {\"")
			continue
		}

		if line == "}" {
			currentName = ""
			continue
		}

		if currentName != "" && strings.Contains(line, "LocalPath") {
			parts := strings.SplitN(line, "=", 2)
			val := strings.Trim(strings.TrimSpace(parts[1]), "\";")
			// Only include dart libraries
			if strings.Contains(val, "/dart/") {
				overrides[currentName] = val
			}
		}
	}

	return overrides, nil
}

func overridePubspec(path string, overrides map[string]string) error {
	content, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	lines := strings.Split(string(content), "\n")
	hasOverrides := false
	for _, line := range lines {
		if strings.HasPrefix(strings.TrimSpace(line), "dependency_overrides:") {
			hasOverrides = true
			break
		}
	}

	if !hasOverrides && len(overrides) > 0 {
		fmt.Printf("  Injecting overrides into: %s\n", path)
		
		lines = append(lines, "", "dependency_overrides:")
		for pkg, relPath := range overrides {
			// Ensure we use the proper relative path from the root
			lines = append(lines, fmt.Sprintf("  %s:", pkg))
			lines = append(lines, fmt.Sprintf("    path: %s", relPath))
		}
		
		return os.WriteFile(path, []byte(strings.Join(lines, "\n")), 0644)
	}

	return nil
}
