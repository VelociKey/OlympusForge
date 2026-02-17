package main

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

type ModNode struct {
	Name string
	Path string
	Deps []string
}

func main() {
	fmt.Println("Generating Fleet-Wide Topological Map...")
	nodes := make(map[string]ModNode)
	
	// 1. Discover all modules and their paths
	workspaces, _ := filepath.Glob("Olympus*")
	for _, ws := range workspaces {
		filepath.Walk(ws, func(path string, info os.FileInfo, err error) error {
			if err != nil || info.IsDir() || info.Name() != "go.mod" { return nil }
			
			content, _ := os.ReadFile(path)
			modName := extractModule(string(content))
			if modName != "" {
				nodes[modName] = ModNode{
					Name: modName,
					Path: filepath.Dir(path),
					Deps: extractDeps(string(content)),
				}
			}
			return nil
		})
	}

	// 2. Generate FLEET_TOPOLOGY.jebnf
	var sb strings.Builder
	sb.WriteString("# Olympus Fleet Topology (jeBNF)\n")
	sb.WriteString(fmt.Sprintf("mapped_at = %q\n\n", time.Now().Format(time.RFC3339)))

	sb.WriteString("[Nodes]\n")
	for _, n := range nodes {
		sb.WriteString(fmt.Sprintf("Node %q %q\n", n.Name, n.Path))
	}

	sb.WriteString("\n[Dependency_Edges]\n")
	for _, n := range nodes {
		for _, d := range n.Deps {
			// Only map internal fleet dependencies
			if strings.Contains(d, "github.com/VelociKey") {
				sb.WriteString(fmt.Sprintf("Edge %q -> %q\n", n.Name, d))
			}
		}
	}

	os.WriteFile("FLEET_TOPOLOGY.jebnf", []byte(sb.String()), 0644)
	fmt.Println("Fleet Topology (FLEET_TOPOLOGY.jebnf) successfully generated.")
}

func extractModule(content string) string {
	re := regexp.MustCompile(`module\s+([^\s\n\r]+)`)
	match := re.FindStringSubmatch(content)
	if len(match) > 1 { return match[1] }
	return ""
}

func extractDeps(content string) []string {
	deps := []string{}
	// Matches either 'require module/path v1.2.3' or inside a require block
	re := regexp.MustCompile(`(?:require|replace)\s+([^\s\n\r]+)`)
	matches := re.FindAllStringSubmatch(content, -1)
	for _, m := range matches {
		if len(m) > 1 && !strings.HasPrefix(m[1], "(") && m[1] != "module" {
			deps = append(deps, m[1])
		}
	}
	return deps
}
