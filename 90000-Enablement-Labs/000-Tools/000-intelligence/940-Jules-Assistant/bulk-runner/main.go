package main

import (
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"sync"
)

func main() {
	workspaces := []string{
		"OlympusGCP-Compute",
		"OlympusGCP-Data",
		"OlympusGCP-Events",
		"OlympusGCP-FinOps",
		"OlympusGCP-Firebase",
		"OlympusGCP-Intelligence",
		"OlympusGCP-Messaging",
		"OlympusGCP-Observability",
		"OlympusGCP-Storage",
		"OlympusGCP-Vault",
	}

	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	root, _ := os.Getwd()
	// Navigate up to root if needed
	for {
		if _, err := os.Stat(filepath.Join(root, "OlympusForge")); err == nil {
			break
		}
		parent := filepath.Dir(root)
		if parent == root {
			break
		}
		root = parent
	}

	julesPath := filepath.Join(root, "olympus.fleet/00SDLC/OlympusForge/90000-Enablement-Labs/000-Tools/000-bin/jules.cmd")
	if _, err := os.Stat(julesPath); err != nil {
		logger.Error("Jules shim not found", "path", julesPath)
		os.Exit(1)
	}

	throttleLimit := 3
	semaphore := make(chan struct{}, throttleLimit)
	var wg sync.WaitGroup

	logger.Info("🚀 Starting Bulk Jules Execution", "threshold", throttleLimit)

	for _, ws := range workspaces {
		wsPath := filepath.Join(root, ws)
		if _, err := os.Stat(wsPath); err != nil {
			logger.Warn("Workspace not found, skipping", "workspace", ws)
			continue
		}

		wg.Add(1)
		go func(path string) {
			defer wg.Done()
			semaphore <- struct{}{}        // Acquire
			defer func() { <-semaphore }() // Release

			logger.Info("Starting Jules", "workspace", path)
			cmd := exec.Command(julesPath, "/generate:tests", "--target="+path, "--type=unit,integration")

			// Capture output to a log file
			logDir := filepath.Join(root, "olympus.fleet/00SDLC/Olympus2/C0500-Agent-Intelligence-Outputs/LPSV")
			os.MkdirAll(logDir, 0755)
			logFile := filepath.Join(logDir, "jules_"+filepath.Base(path)+".log")
			f, _ := os.Create(logFile)
			defer f.Close()

			cmd.Stdout = f
			cmd.Stderr = f

			if err := cmd.Run(); err != nil {
				logger.Error("Jules failed", "workspace", path, "error", err)
			} else {
				logger.Info("Finished Jules", "workspace", path)
			}
		}(wsPath)
	}

	wg.Wait()
	logger.Info("✅ Bulk Jules Execution Complete")
}
