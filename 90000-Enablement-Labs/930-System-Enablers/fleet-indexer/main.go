package main

import (
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
	root, _ := filepath.Abs("../../../..")
	var modules []Module

	filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil || !info.IsDir() { return nil }
		if info.Name() == ".git" || info.Name() == "node_modules" { return filepath.SkipDir }
		
		if _, err := os.Stat(filepath.Join(path, "go.mod")); err == nil {
			rel, _ := filepath.Rel(root, path)
			modules = append(modules, Module{
				Name: filepath.ToSlash(rel),
				Path: "./" + filepath.ToSlash(rel),
			})
		}
		return nil
	})

	// Generate FLEET_GO_DNA.jebnf
	var sb strings.Builder
	sb.WriteString("# Olympus Fleet Go Module Registry (jeBNF)\n")
	sb.WriteString(fmt.Sprintf("indexed_at = %q\n\n", time.Now().Format(time.RFC3339)))
	for _, m := range modules {
		sb.WriteString(fmt.Sprintf("Module %q %q\n", m.Name, m.Path))
	}
	os.WriteFile("FLEET_GO_DNA.jebnf", []byte(sb.String()), 0644)
	fmt.Println("FLEET_GO_DNA.jebnf generated.")
}
