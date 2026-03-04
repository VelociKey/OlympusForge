package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

func main() {
	fmt.Println("Starting OlympusForge Provisioner (Dagger-Driven)...")

	// Detect Fleet Root
	wd, _ := os.Getwd()
	root := wd
	if filepath.Base(wd) == "OlympusForge" {
		root = filepath.Dir(wd)
	}

	forgePkg := filepath.Join(root, "00SDLC", "OlympusForge", "90000-Enablement-Labs", "900-Forge")

	fmt.Println("🔨 Building OlympusForge via Forge Pipeline...")

	// For OlympusForge, we typically want to 'Assess' or build specific tools.
	// For now, we default to 'native' build or assessment.
	// We use '-workspace 00SDLC/OlympusForge' to target this workspace context.
	cmd := exec.Command("go", "run", forgePkg, "-target", "native", "-workspace", "00SDLC/OlympusForge")
	cmd.Dir = root
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		fmt.Printf("❌ OlympusForge provisioning failed: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("✅ OlympusForge provisioning complete.")
}
