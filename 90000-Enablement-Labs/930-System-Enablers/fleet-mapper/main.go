package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type Node struct {
	Name string
	Path string
}

func main() {
	root, _ := filepath.Abs("../../../..")
	var nodes []Node

	filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil || !info.IsDir() { return nil }
		if strings.HasPrefix(info.Name(), ".") { return filepath.SkipDir }
		
		rel, _ := filepath.Rel(root, path)
		nodes = append(nodes, Node{
			Name: filepath.ToSlash(rel),
			Path: filepath.ToSlash(rel),
		})
		return nil
	})

	// Generate FLEET_TOPOLOGY.jebnf
	var sb strings.Builder
	sb.WriteString("# Olympus Fleet Topology (jeBNF)\n")
	sb.WriteString(fmt.Sprintf("mapped_at = %q\n\n", time.Now().Format(time.RFC3339)))

	sb.WriteString("[Nodes]\n")
	for _, n := range nodes {
		sb.WriteString(fmt.Sprintf("Node %q %q\n", n.Name, n.Path))
	}

	sb.WriteString("\n[Dependency_Edges]\n")
	// Placeholder for edge discovery
	
	os.WriteFile("FLEET_TOPOLOGY.jebnf", []byte(sb.String()), 0644)
	fmt.Println("FLEET_TOPOLOGY.jebnf generated.")
}
