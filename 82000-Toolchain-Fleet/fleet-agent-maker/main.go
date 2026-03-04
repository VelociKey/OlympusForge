package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: fleet-agent-maker <request.jebnf>")
		os.Exit(1)
	}

	requestPath := os.Args[1]
	requestData, err := ioutil.ReadFile(requestPath)
	if err != nil {
		fmt.Printf("Error reading request: %v\n", err)
		os.Exit(1)
	}

	// Simple extraction for prototype (regex or string split)
	// In production, use 00SDLC/OlympusGrammar/parser
	role := extractValue(string(requestData), "Role")
	role = strings.ReplaceAll(role, "/", "-")
	objective := extractValue(string(requestData), "PrimaryObjective")

	fmt.Printf("Creating Agent for Role: %s\n", role)
	fmt.Printf("Objective: %s\n", objective)

	// Create output dir in @SCRATCH
	// For this simulation, we use C:\aAntigravitySpace\00SDLC\OlympusForge\C0990-Ephemeral-Scratch
	scratchRoot := filepath.Join("00SDLC", "OlympusForge", "C0990-Ephemeral-Scratch")
	agentID := fmt.Sprintf("agent-%s", role)
	agentDir := filepath.Join(scratchRoot, agentID)

	os.MkdirAll(agentDir, 0755)

	// Clone base-go template
	templateDir := filepath.Join("00SDLC", "OlympusForge", "83000-Agent-Templates", "base-go")
	cloneTemplate(templateDir, agentDir, role, objective)

	fmt.Printf("✅ Agent scaffolded at: %s\n", agentDir)
}

func extractValue(data, key string) string {
	lines := strings.Split(data, "\n")
	for _, line := range lines {
		if strings.Contains(line, key) {
			parts := strings.Split(line, "\"")
			if len(parts) > 1 {
				return parts[1]
			}
		}
	}
	return "unknown"
}

func cloneTemplate(src, dst, role, objective string) {
	filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
		if info.IsDir() {
			return nil
		}
		rel, _ := filepath.Rel(src, path)
		targetPath := filepath.Join(dst, rel)
		
		data, _ := ioutil.ReadFile(path)
		content := string(data)
		
		// Inject role-specific objective
		if rel == "main.go" {
			content = strings.Replace(content, "// Agent loop logic", fmt.Sprintf("// Role: %s\n\t// Objective: %s", role, objective), 1)
		}
		if rel == "CYCLE_METRICS.jebnf" {
			content = strings.Replace(content, "[OBJECTIVE]", objective, 1)
		}
		
		os.MkdirAll(filepath.Dir(targetPath), 0755)
		ioutil.WriteFile(targetPath, []byte(content), 0644)
		return nil
	})
}
