package main

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
)

func main() {

	fmt.Println("Starting OlympusForge Tool Provisioner...")

	// 0. Ensure go.work exists (Self-Healing)
	if _, err := os.Stat("go.work"); os.IsNotExist(err) {
		fmt.Println("⚠️ go.work missing! Attempting auto-repair...")
		// Use template from OlympusForge
		templatePath := filepath.Join("OlympusForge", "90000-Enablement-Labs", "go.work.template")
		if err := copyFile(templatePath, "go.work"); err != nil {
			fmt.Printf("Failed to restore template: %v\n", err)
		}

		sovereignPath := filepath.Join("90000-Enablement-Labs", "000-Tools", "sovereign", "main.go")
		repairCmd := exec.Command("go", "run", sovereignPath, "init")
		repairCmd.Dir = "OlympusForge"
		repairCmd.Stdout = os.Stdout
		repairCmd.Stderr = os.Stderr
		if err := repairCmd.Run(); err != nil {
			fmt.Printf("Repair failed: %v. Please run 'sovereign init' manually from within OlympusForge.\n", err)
		} else {
			fmt.Println("✅ go.work successfully regenerated.")
		}
	}

	// 1. Run the Go-native provisioner logic
	// Path is relative to fleet root

	provisionerPath := filepath.Join("OlympusForge", "90000-Enablement-Labs", "910-provisioner", "main.go")

	fmt.Println("Running tool provisioner...")

	cmd := exec.Command("go", "run", provisionerPath)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {

		fmt.Printf("Tool provisioning failed: %v\n", err)
		os.Exit(1)
	}

	// 2. Run NomenclatureStabilizer to ensure Sovereign Imports are aligned

	stabilizerPath := filepath.Join("OlympusAtelier", "90000-Enablement-Labs", "930-System-Enablers", "000-NomenclatureStabilizer", "main.go")

	fmt.Println("Running NomenclatureStabilizer...")

	stabCmd := exec.Command("go", "run", stabilizerPath)
	stabCmd.Stdout = os.Stdout
	stabCmd.Stderr = os.Stderr
	if err := stabCmd.Run(); err != nil {

		fmt.Printf("Stabilization failed: %v\n", err)
		os.Exit(1)
	}

	// 3. Final Workspace Sync

	fmt.Println("Synchronizing workspace...")

	syncCmd := exec.Command("go", "work", "sync")
	syncCmd.Run()

	// 4. Project Workspace Provisioning (George Sandbox)
	fmt.Println("Provisioning Project Workspaces...")
	georgeProvPath := filepath.Join("George", "provision.go")
	if _, err := os.Stat(georgeProvPath); err == nil {
		fmt.Println("Running George Provisioner...")
		georgeCmd := exec.Command("go", "run", georgeProvPath)
		georgeCmd.Stdout = os.Stdout
		georgeCmd.Stderr = os.Stderr
		if err := georgeCmd.Run(); err != nil {
			fmt.Printf("⚠️ George provisioning failed: %v\n", err)
		}
	}

	fmt.Println("OlympusForge provisioning complete.")
}

func copyFile(src, dst string) error {
	sourceFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer sourceFile.Close()

	destFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer destFile.Close()

	_, err = io.Copy(destFile, sourceFile)
	return err
}
