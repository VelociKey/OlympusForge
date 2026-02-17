package main

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

type Node struct {
	Name     string
	Files    []FileInfo
	Children map[string]*Node
}

type FileInfo struct {
	Name string
	Ext  string
	Hash string
	Size int64
}

func NewNode(name string) *Node {
	return &Node{Name: name, Children: make(map[string]*Node)}
}

func main() {
	workspaces, _ := filepath.Glob("Olympus*")
	currentFleetState := loadFleetState("FLEET_DNA.jebnf")
	
	fleetDNA := fmt.Sprintf("# AntigravitySpace-Olympus-Fleet (jeBNF-MPH)\ngenerated_at = %q\n\n", time.Now().Format(time.RFC3339))

	for _, ws := range workspaces {
		info, err := os.Stat(ws)
		if err != nil || !info.IsDir() { continue }

		newHash := calculateWorkspaceHash(ws)
		if oldHash, ok := currentFleetState[ws]; ok && oldHash == newHash {
			fmt.Printf("Workspace %s unchanged. Skipping re-index.\n", ws)
			fleetDNA += fmt.Sprintf("Workspace %q %s %q\n", ws, newHash, inferDescription(ws))
			continue
		}

		fmt.Printf("Deep Indexing: %s (Min-Storage MPH Phase)...\n", ws)
		root := NewNode(ws)
		
		// Phase 1: Build Semantic Trie
		filepath.Walk(ws, func(path string, info os.FileInfo, err error) error {
			if err != nil || info.IsDir() || strings.Contains(path, ".git") || strings.Contains(path, ".workspace-dna.") {
				return nil
			}
			rel, _ := filepath.Rel(ws, path)
			parts := strings.Split(rel, string(os.PathSeparator))
			
			curr := root
			for i := 0; i < len(parts)-1; i++ {
				if _, ok := curr.Children[parts[i]]; !ok {
					curr.Children[parts[i]] = NewNode(parts[i])
				}
				curr = curr.Children[parts[i]]
			}
			
			fileName := parts[len(parts)-1]
			curr.Files = append(curr.Files, FileInfo{
				Name: fileName,
				Ext:  filepath.Ext(fileName),
				Hash: hashFile(path),
				Size: info.Size(),
			})
			return nil
		})

		// Phase 2: Generate Token-Minimum jeBNF
		wsDNA := fmt.Sprintf("# %s Workspace DNA (jeBNF-MPH)\nworkspace_name = %q\n\n", ws, ws)
		wsDNA += generateJebnf(root, 0, new(int))
		
		os.WriteFile(filepath.Join(ws, ".workspace-dna.jebnf"), []byte(wsDNA), 0644)
		fleetDNA += fmt.Sprintf("Workspace %q %s %q\n", ws, newHash, inferDescription(ws))
	}

	os.WriteFile("FLEET_DNA.jebnf", []byte(fleetDNA), 0644)
}

func generateJebnf(node *Node, indent int, index *int) string {
	var sb strings.Builder
	pad := strings.Repeat("  ", indent)
	
	// Sort children for determinism
	childNames := make([]string, 0, len(node.Children))
	for name := range node.Children { childNames = append(childNames, name) }
	sort.Strings(childNames)

	// Sort files for determinism
	sort.Slice(node.Files, func(i, j int) bool { return node.Files[i].Name < node.Files[j].Name })

	if indent > 0 {
		sb.WriteString(fmt.Sprintf("%sDomain %q {\n", pad, node.Name))
	}

	for _, f := range node.Files {
		// Minimum storage: ID [Ext] [Name] [Hash] [Size]
		sb.WriteString(fmt.Sprintf("%s  Item %d %s %q %s %d\n", pad, *index, f.Ext, f.Name, f.Hash, f.Size))
		*index++
	}

	for _, name := range childNames {
		sb.WriteString(generateJebnf(node.Children[name], indent+1, index))
	}

	if indent > 0 {
		sb.WriteString(fmt.Sprintf("%s}\n", pad))
	}
	return sb.String()
}

func calculateWorkspaceHash(ws string) string {
	h := sha256.New()
	filepath.Walk(ws, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() || strings.Contains(path, ".git") || strings.Contains(path, ".workspace-dna.") {
			return nil
		}
		h.Write([]byte(path))
		h.Write([]byte(fmt.Sprintf("%d", info.ModTime().Unix())))
		return nil
	})
	return hex.EncodeToString(h.Sum(nil))
}

func loadFleetState(path string) map[string]string {
	state := make(map[string]string)
	content, _ := os.ReadFile(path)
	lines := strings.Split(string(content), "\n")
	for _, line := range lines {
		if strings.HasPrefix(line, "Workspace") {
			parts := strings.Fields(line)
			if len(parts) >= 3 {
				name := strings.Trim(parts[1], "\"")
				hash := parts[2]
				state[name] = hash
			}
		}
	}
	return state
}

func hashFile(path string) string {
	f, _ := os.Open(path)
	defer f.Close()
	h := sha256.New()
	io.Copy(h, f)
	return hex.EncodeToString(h.Sum(nil))
}

func inferDescription(name string) string {
	switch {
	case strings.Contains(name, "Cognition"): return "Primary Cognitive Synthesis & Logic Construction"
	case strings.Contains(name, "Delegation"): return "Autonomous Delegation & Market Coordination"
	case strings.Contains(name, "Fabric"): return "Universal Semantic Source of Truth"
	case strings.Contains(name, "Forge"): return "Central Build & Binary Artifact Repository"
	case name == "Olympus2": return "System Ghost Architecture & Communication Contracts"
	default: return "General Domain Workspace"
	}
}
