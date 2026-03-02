package main

import (
	"flag"
	"fmt"
	"io/fs"
	"os/exec"
	"path/filepath"
)

type Prerequisite struct {
	Name     string
	Command  string
	Args     []string
	WingetID string
}

var prereqs = []Prerequisite{
	{Name: "Podman Desktop", Command: "podman", Args: []string{"--version"}, WingetID: "RedHat.Podman-Desktop"},
	{Name: "Go Runtime", Command: "go", Args: []string{"version"}, WingetID: "GoLang.Go"},
	{Name: "Git CLI", Command: "git", Args: []string{"--version"}, WingetID: "Git.Git"},
}

func main() {
	install := flag.Bool("install", false, "Install missing prerequisites")
	update := flag.Bool("update", false, "Update existing prerequisites to latest version")
	scorch := flag.Bool("scorch", false, "Report legacy binaries found outside the Forge")
	flag.Parse()

	fmt.Println("🚀 Olympus Fleet Bootstrap: Sovereign Alignment")

	for _, p := range prereqs {
		fmt.Printf("🔍 Checking: %s... ", p.Name)
		if isInstalled(p.Command, p.Args...) {
			fmt.Println("INSTALLED")
			if *update && p.WingetID != "" {
				updateViaWinget(p.WingetID)
			}
		} else {
			fmt.Println("MISSING")
			if *install && p.WingetID != "" {
				installViaWinget(p.WingetID)
			}
		}
	}

	if *install {
		setupPodmanMachine()
	}

	if *scorch {
		runBinaryNormalizationReport()
	}

	fmt.Println("\n✅ Bootstrap sequence complete.")
}

func isInstalled(name string, args ...string) bool {
	cmd := exec.Command(name, args...)
	err := cmd.Run()
	return err == nil
}

func installViaWinget(id string) {
	fmt.Printf("🛠️  Installing %s via Winget... ", id)
	cmd := exec.Command("winget", "install", "-e", "--id", id, "--silent", "--accept-source-agreements", "--accept-package-agreements")
	err := cmd.Run()
	if err != nil {
		fmt.Printf("FAILED: %v\n", err)
	} else {
		fmt.Println("SUCCESS")
	}
}

func updateViaWinget(id string) {
	fmt.Printf("🔄 Updating %s via Winget... ", id)
	cmd := exec.Command("winget", "upgrade", "-e", "--id", id, "--silent", "--accept-source-agreements", "--accept-package-agreements")
	err := cmd.Run()
	if err != nil {
		fmt.Printf("FAILED or ALREADY AT LATEST: %v\n", err)
	} else {
		fmt.Println("SUCCESS")
	}
}

func setupPodmanMachine() {
	fmt.Printf("🐳 Initializing fresh Podman Machine (Capped at 50GB)... ")
	cmd := exec.Command("podman", "machine", "init", "--cpus", "4", "--memory", "4096", "--disk-size", "50")
	err := cmd.Run()
	if err != nil {
		fmt.Printf("FAILED or ALREADY EXISTS: %v\n", err)
	} else {
		fmt.Println("SUCCESS")
		fmt.Println("▶️  Starting machine...")
		exec.Command("podman", "machine", "start").Run()
	}
}

func runBinaryNormalizationReport() {
	fmt.Println("\n--- Sovereign Scorch Report ---")
	filepath.WalkDir(".", func(path string, d fs.DirEntry, err error) error {
		if err != nil { return nil }
		if d.IsDir() && (d.Name() == "81000-Toolchain-External" || d.Name() == "82000-Toolchain-Fleet" || d.Name() == "node_modules" || d.Name() == ".git") {
			return filepath.SkipDir
		}
		ext := filepath.Ext(path)
		if ext == ".exe" || ext == ".cmd" || ext == ".bat" {
			fmt.Printf("⚠️  ORPHAN: %s\n", path)
		}
		return nil
	})
}
