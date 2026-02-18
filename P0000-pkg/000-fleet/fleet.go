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
		matches, _ := filepath.Glob(filepath.Join(path, "*.go"))
		if len(matches) > 0 {
			rel, _ := filepath.Rel(root, path)
			modName := filepath.ToSlash(rel)
			if modName == "." { modName = "fleet-root" }
			if strings.HasSuffix(modName, "/v1") { modName += "x" }
			if strings.HasSuffix(modName, "/v2") { modName += "x" }
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
		rel, _ := filepath.Rel(mod.Path, filepath.Join(root, target))
		relPath := filepath.ToSlash(rel)
		if !strings.HasPrefix(relPath, ".") { relPath = "./" + relPath }
		sb.WriteString(fmt.Sprintf("replace %s => %s\n", s, relPath))
	}
	return os.WriteFile(modPath, []byte(sb.String()), 0644)
}

func SyncGoWork(root string, modules []Module) error {
	var sb strings.Builder
	sb.WriteString("go 1.25.7\n\nuse (\n")
	for _, m := range modules {
		rel, _ := filepath.Rel(root, m.Path)
		sb.WriteString(fmt.Sprintf("\t./%s\n", filepath.ToSlash(rel)))
	}
	sb.WriteString(")\n")
	return os.WriteFile(filepath.Join(root, "go.work"), []byte(sb.String()), 0644)
}
