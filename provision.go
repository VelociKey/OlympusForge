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
	fmt.Println("⏳ This process uses the Fleet-Standard Dagger Pipeline.")

	// Standardized build command: -target and -workspace
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
