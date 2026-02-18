package main

import (
	"crypto/sha256"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"
)

type Node struct {
	Name     string
	Path     string
	Hash     string
	Children map[string]*Node
}

func NewNode(name string) *Node {
	return &Node{Name: name, Children: make(map[string]*Node)}
}

func main() {
	workspaces, _ := filepath.Glob("Olympus*")
	
	fleetDNA := fmt.Sprintf("# AntigravitySpace-Olympus-Fleet (jeBNF-MPH)\ngenerated_at = %q\n\n", time.Now().Format(time.RFC3339))

	for _, ws := range workspaces {
		info, err := os.Stat(ws)
		if err != nil || !info.IsDir() { continue }

		newHash := calculateWorkspaceHash(ws)
		fleetDNA += fmt.Sprintf("Workspace %q %s %q\n", ws, newHash, "General Domain Workspace")
	}

	os.WriteFile("FLEET_DNA.jebnf", []byte(fleetDNA), 0644)
	fmt.Println("FLEET_DNA.jebnf updated.")
}

func calculateWorkspaceHash(ws string) string {
	h := sha256.New()
	filepath.Walk(ws, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() { return nil }
		f, _ := os.Open(path)
		defer f.Close()
		io.Copy(h, f)
		return nil
	})
	return fmt.Sprintf("%x", h.Sum(nil))
}
