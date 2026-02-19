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
		if err != nil || !info.IsDir() { return nil }
		name := info.Name()
		if strings.HasPrefix(name, ".") || name == "node_modules" || name == "data" || name == "dist" {
			return filepath.SkipDir
		}
		// Check for go.mod instead of .go files to identify modules accurately
		if _, err := os.Stat(filepath.Join(path, "go.mod")); err == nil {
			rel, _ := filepath.Rel(root, path)
			modName := filepath.ToSlash(rel)
			if modName == "." { return nil } // Skip root as we handled its removal
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
	content := fmt.Sprintf("module %s\n\ngo 1.25.7\n", mod.Name)
	var sb strings.Builder
	sb.WriteString(content)
	sb.WriteString("\n// Local Resolution\n")
	sort.Slice(allModules, func(i, j int) bool {
		return allModules[i].Name < allModules[j].Name
	})
	for _, m := range allModules {
		if m.Name == mod.Name { continue }
		rel, _ := filepath.Rel(mod.Path, m.Path)
		relPath := filepath.ToSlash(rel)
		if !strings.HasPrefix(relPath, ".") { relPath = "./" + relPath }
		sb.WriteString(fmt.Sprintf("replace %s => %s\n", m.Name, relPath))
	}
	sovereigns := []string{"text", "pretty", "go-internal", "check.v1", "gopkg.in/check.v1"}
	for _, s := range sovereigns {
		target := s
		if s == "gopkg.in/check.v1" { target = "check.v1" }
		// Libraries are now in Olympus2/00000-Identity-Foundations/P0000-pkg/
		libPath := filepath.Join(root, "Olympus2", "00000-Identity-Foundations", "P0000-pkg", target)
		rel, _ := filepath.Rel(mod.Path, libPath)
		relPath := filepath.ToSlash(rel)
		if !strings.HasPrefix(relPath, ".") { relPath = "./" + relPath }
		sb.WriteString(fmt.Sprintf("replace %s => %s\n", s, relPath))
	}
	return os.WriteFile(modPath, []byte(sb.String()), 0644)
}

func SyncGoWork(root string, modules []Module) error {
	var sb strings.Builder
	sb.WriteString("go 1.25.7\n\nuse (\n")
	
	// Sort modules for deterministic output
	sort.Slice(modules, func(i, j int) bool {
		return modules[i].Path < modules[j].Path
	})

	// Olympus2 is the product root. go.work should live in Olympus2/
	targetRoot := filepath.Join(root, "Olympus2")

	for _, m := range modules {
		rel, _ := filepath.Rel(targetRoot, m.Path)
		sb.WriteString(fmt.Sprintf("\t./%s\n", filepath.ToSlash(rel)))
	}
	sb.WriteString(")\n")
	return os.WriteFile(filepath.Join(targetRoot, "go.work"), []byte(sb.String()), 0644)
}

func MigrateImports(root, oldPath, newPath string) error {
	return filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil { return err }
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
			if err != nil { return err }
			newContent := strings.ReplaceAll(string(content), oldPath, newPath)
			if newContent != string(content) {
				fmt.Printf("Updating %s\n", path)
				return os.WriteFile(path, []byte(newContent), info.Mode())
			}
		}
		return nil
	})
}
