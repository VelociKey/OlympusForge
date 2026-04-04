package main

import (
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
)

func main() {
	slog.Info("Starting OlympusForge Provisioner (Bazel-Driven)")

	// Detect Fleet Root
	wd, _ := os.Getwd()
	root := wd
	if filepath.Base(wd) == "OlympusForge" {
		root = filepath.Dir(wd)
	}

	forgePkg := filepath.Join(root, "00SDLC", "OlympusForge", "90000-Enablement-Labs", "900-Forge")

	slog.Info("Building OlympusForge via Forge Pipeline", "mode", "Bazel")

	// Standardized build command: -target and -workspace
	cmd := exec.Command("go", "run", forgePkg, "-target", "native", "-workspace", "00SDLC/OlympusForge")
	cmd.Dir = root
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		slog.Error("OlympusForge provisioning failed", "error", err)
		os.Exit(1)
	}

	slog.Info("OlympusForge provisioning complete")
}
