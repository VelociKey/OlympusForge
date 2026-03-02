package main

import (
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

func main() {
	threads := flag.Int("threads", 4, "Maximum concurrent test threads")
	outputLog := flag.String("output", "olympus.fleet/00SDLC/Olympus2/C0500-Agent-Intelligence-Outputs/LPSV/error_log.lpsv", "Path to error log")
	targetWS := flag.String("ws", "", "Specific workspace to target (comma-separated)")
	flag.Parse()

	allWorkspaces := []string{
		"George", "OlympusActors-Cognition", "OlympusActors-Delegation",
		"OlympusAscent", "OlympusAssurance", "OlympusAtelier",
		"OlympusFabric", "OlympusForge", "OlympusGCP-Compute",
		"OlympusGCP-Data", "OlympusGCP-Events", "OlympusGCP-FinOps",
		"OlympusGCP-Firebase", "OlympusGCP-Intelligence", "OlympusGCP-Messaging",
		"OlympusGCP-Observability", "OlympusGCP-Storage", "OlympusGCP-Vault",
		"OlympusGrammar", "OlympusInfrastructure", "OlympusMCP",
		"OlympusVision", "OpenClaw",
	}

	var workspaces []string
	if *targetWS != "" {
		targets := strings.Split(*targetWS, ",")
		for _, t := range targets {
			workspaces = append(workspaces, strings.TrimSpace(t))
		}
	} else {
		workspaces = allWorkspaces
	}

	os.MkdirAll(filepath.Dir(*outputLog), 0755)
	logFile, err := os.Create(*outputLog)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating log: %v\n", err)
		os.Exit(1)
	}
	defer logFile.Close()

	fmt.Printf("🎯 Starting Mission Runner (Threads: %d)\n", *threads)
	fmt.Printf("📝 Logging failures to: %s\n", *outputLog)

	sem := make(chan struct{}, *threads)
	var wg sync.WaitGroup
	var mu sync.Mutex

	startTime := time.Now()

	for _, ws := range workspaces {
		wg.Add(1)
		go func(workspace string) {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()

			fmt.Printf("🏃 Testing %s...\n", workspace)
			
			err := filepath.Walk(workspace, func(path string, info os.FileInfo, err error) error {
				if err != nil || !info.IsDir() {
					return nil
				}
				if strings.HasPrefix(info.Name(), ".") || info.Name() == "node_modules" || info.Name() == "vendor" {
					return filepath.SkipDir
				}

				// Check if this directory has go files and test files
				entries, _ := os.ReadDir(path)
				hasGo := false
				hasTests := false
				for _, e := range entries {
					if !e.IsDir() {
						if strings.HasSuffix(e.Name(), ".go") {
							hasGo = true
						}
						if strings.HasSuffix(e.Name(), "_test.go") {
							hasTests = true
						}
					}
				}

				if hasGo && hasTests {
					rel, _ := filepath.Rel(workspace, path)
					fmt.Printf("  [pkg] %s/%s\n", workspace, rel)
					cmd := exec.Command("go", "test", "-v", ".")
					cmd.Dir = path
					cmd.Env = os.Environ()
					
					output, err := cmd.CombinedOutput()
					if err != nil {
						mu.Lock()
						fmt.Fprintf(logFile, "timestamp: %s | workspace: %s | path: %s | error: %v | output: %s\n",
							time.Now().Format(time.RFC3339), workspace, rel, err, strings.ReplaceAll(string(output), "\n", " "))
						mu.Unlock()
						fmt.Printf("  ❌ %s/%s FAILED\n", workspace, rel)
					} else {
						// Too much noise if we print every package pass
						// fmt.Printf("  ✅ %s/%s PASSED\n", workspace, rel)
					}
				}
				return nil
			})
			if err != nil {
				fmt.Printf("⚠️ Error walking %s: %v\n", workspace, err)
			}
			fmt.Printf("✅ %s Complete\n", workspace)
		}(ws)
	}

	wg.Wait()
	fmt.Printf("✨ Mission Complete. Duration: %v\n", time.Since(startTime))
}
