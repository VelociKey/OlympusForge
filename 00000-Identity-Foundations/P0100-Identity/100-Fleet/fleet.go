package fleet

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

type Module struct {
	Name string
	Path string
}

func FindGoModules(root string) ([]Module, error) {
	var modules []Module
	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil || !info.IsDir() {
			return nil
		}
		name := info.Name()
		// We allow ZC directories because they contain sovereign source modules (like MCP)
		if strings.HasPrefix(name, ".") || name == "node_modules" || name == "data" || name == "dist" {
			return filepath.SkipDir
		}
		if _, err := os.Stat(filepath.Join(path, "go.mod")); err == nil {
			rel, _ := filepath.Rel(root, path)
			modName := filepath.ToSlash(rel)
			if modName == "." {
				// We still want to include the root module if it exists
				// but let's check its name from go.mod first
				return nil
			}

			// Detect module name from go.mod file
			modContent, err := os.ReadFile(filepath.Join(path, "go.mod"))
			if err == nil {
				lines := strings.Split(string(modContent), "\n")
				for _, line := range lines {
					if strings.HasPrefix(line, "module ") {
						modName = strings.TrimSpace(strings.TrimPrefix(line, "module "))
						break
					}
				}
			}

			modules = append(modules, Module{
				Name: modName,
				Path: path,
			})
		}
		return nil
	})
	return modules, err
}

func SyncGoMod(mod Module, allModules []Module, root string) error {
	modPath := filepath.Join(mod.Path, "go.mod")
	fmt.Printf("DEBUG: Syncing %s\n", modPath)

	content, err := os.ReadFile(modPath)
	if err != nil {
		return fmt.Errorf("failed to read go.mod: %v", err)
	}

	lines := strings.Split(string(content), "\n")
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("module %s\n\ngo 1.25.7\n", mod.Name))

	// Project Mandate: Local cross-module dependency resolution MUST be handled exclusively via the root go.work file.
	// We explicitly DO NOT add 'replace' directives here.

	inRequire := false
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "require") {
			inRequire = true
			sb.WriteString("\n" + line + "\n")
			continue
		}
		if inRequire {
			sb.WriteString(line + "\n")
			if trimmed == ")" {
				inRequire = false
			}
		}
	}

	return os.WriteFile(modPath, []byte(sb.String()), 0644)
}

func SyncGoWork(root string, modules []Module) error {
	var sb strings.Builder
	sb.WriteString("go 1.25.7\n\nuse (\n")

	// Sort by path for deterministic output
	sort.Slice(modules, func(i, j int) bool {
		return modules[i].Path < modules[j].Path
	})

	for _, m := range modules {
		rel, err := filepath.Rel(root, m.Path)
		if err != nil {
			continue
		}
		sb.WriteString(fmt.Sprintf("\t./%s\n", filepath.ToSlash(rel)))
	}
	sb.WriteString(")\n")

	// Ensure we write to the Fleet Root
	return os.WriteFile(filepath.Join(root, "go.work"), []byte(sb.String()), 0644)
}

func MigrateImports(root, oldPath, newPath string) error {
	return filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			name := info.Name()
			if strings.HasPrefix(name, ".") || name == "node_modules" || name == "data" || name == "dist" || name == "vendor" {
				return filepath.SkipDir
			}
			return nil
		}
		ext := filepath.Ext(path)
		if ext == ".go" || info.Name() == "go.mod" {
			content, err := os.ReadFile(path)
			if err != nil {
				return err
			}
			newContent := strings.ReplaceAll(string(content), oldPath, newPath)
			if newContent != string(content) {
				return os.WriteFile(path, []byte(newContent), info.Mode())
			}
		}
		return nil
	})
}

func ScaffoldModule(root, moduleName string) error {
	modPath := filepath.Join(root, moduleName)
	if _, err := os.Stat(modPath); err == nil {
		return fmt.Errorf("module directory already exists: %s", modPath)
	}

	fmt.Printf("🏗️ Scaffolding Module: %s\n", moduleName)

	// 1. Create Directory Taxonomy
	dirs := []string{
		"00000-Identity-Foundations/P0000-pkg",
		"10000-Autonomous-Actors",
		"20000-Context-Bridges",
		"30000-Federated-Services",
		"40000-Communication-Contracts/430-Protocol-Definitions/000-gen",
		"50000-Intelligence-Framework",
		"60000-Information-Storage",
		"70000-Environmental-Harness/dagger",
		"80000-System-Governance",
		"90000-Enablement-Labs",
		"C0100-Configuration-Registry",
		"C0400-Artifact-Repository",
		"C0500-Agent-Intelligence-Outputs",
		"C0700-System-Operations",
		"C0990-Ephemeral-Scratch",
		"K0000-KnowledgeAnchor",
	}

	for _, d := range dirs {
		if err := os.MkdirAll(filepath.Join(modPath, d), 0755); err != nil {
			return err
		}
	}

	// 2. Initialize go.mod
	goModContent := fmt.Sprintf("module %s\n\ngo 1.25.7\n", moduleName)
	if err := os.WriteFile(filepath.Join(modPath, "go.mod"), []byte(goModContent), 0644); err != nil {
		return err
	}

	// 3. Add placeholders
	readme := fmt.Sprintf("# %s\n\nOlympus Sovereign Module\n", moduleName)
	os.WriteFile(filepath.Join(modPath, "README.md"), []byte(readme), 0644)

	daggerJson := `{
    "sdk": "go",
    "source": "70000-Environmental-Harness/dagger"
}`
	os.WriteFile(filepath.Join(modPath, "dagger.json"), []byte(daggerJson), 0644)
	os.WriteFile(filepath.Join(modPath, "70000-Environmental-Harness/dagger/main.go"), []byte("package main\n"), 0644)

	return nil
}

func VerifyFleet(root string) error {
	fmt.Println("🔍 Verifying Fleet Integrity...")

	modules, err := FindGoModules(root)
	if err != nil {
		return err
	}

	var issues []string

	// 1. Check for go.work consistency
	workPath := filepath.Join(root, "go.work")
	workContent, err := os.ReadFile(workPath)
	if err != nil {
		issues = append(issues, "Missing root go.work file")
	} else {
		for _, m := range modules {
			rel, _ := filepath.Rel(root, m.Path)
			relPath := filepath.ToSlash(rel)
			if !strings.Contains(string(workContent), "./"+relPath) {
				issues = append(issues, fmt.Sprintf("Module %s not registered in go.work", m.Name))
			}
		}
	}

	// 2. Check Module Consistency
	for _, mod := range modules {
		modPath := filepath.Join(mod.Path, "go.mod")
		content, err := os.ReadFile(modPath)
		if err != nil {
			issues = append(issues, fmt.Sprintf("Missing go.mod in %s", mod.Name))
			continue
		}

		// Check Go Version
		if !strings.Contains(string(content), "go 1.25.7") {
			issues = append(issues, fmt.Sprintf("Module %s uses incorrect Go version", mod.Name))
		}

		// Check local replacements exist
		lines := strings.Split(string(content), "\n")
		for _, line := range lines {
			if strings.HasPrefix(strings.TrimSpace(line), "replace ") {
				parts := strings.Fields(line)
				if len(parts) >= 2 {
					target := parts[1] // The module name being replaced
					// Project Mandate: Do NOT use replace directives in go.mod files for local resolution.
					issues = append(issues, fmt.Sprintf("Module %s contains forbidden 'replace' directive for %s", mod.Name, target))
				}
			}
		}
	}

	if len(issues) > 0 {
		fmt.Printf("❌ Found %d issues during fleet verification:\n", len(issues))
		for _, issue := range issues {
			fmt.Printf("  - %s\n", issue)
		}
		return fmt.Errorf("fleet verification failed")
	}

	fmt.Println("✅ Fleet integrity verified. All structures are valid.")
	return nil
}
