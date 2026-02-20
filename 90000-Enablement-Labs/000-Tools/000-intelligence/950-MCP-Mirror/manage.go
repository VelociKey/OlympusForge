package main

import (
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

const (
	mcpRepo = "https://github.com/mark3labs/mcp-go"
	zcDir   = "ZC0400-Sovereign-Source"
)

func main() {
	installCmd := flag.NewFlagSet("install", flag.ExitOnError)
	updateCmd := flag.NewFlagSet("update", flag.ExitOnError)

	if len(os.Args) < 2 {
		fmt.Println("Expected 'install' or 'update' subcommands")
		os.Exit(1)
	}

	switch os.Args[1] {
	case "install":
		installCmd.Parse(os.Args[2:])
		runInstall()
	case "update":
		updateCmd.Parse(os.Args[2:])
		runUpdate()
	default:
		fmt.Println("Expected 'install' or 'update' subcommands")
		os.Exit(1)
	}
}

func getPaths() (root, target string) {
	cwd, _ := os.Getwd()
	curr := cwd
	for {
		if _, err := os.Stat(filepath.Join(curr, "OlympusForge")); err == nil {
			root = curr
			break
		}
		parent := filepath.Dir(curr)
		if parent == curr {
			root = cwd
			break
		}
		curr = parent
	}
	target = filepath.Join(root, "OlympusForge", zcDir, "mcp-go")
	return
}

func runInstall() {
	root, target := getPaths()
	parent := filepath.Dir(target)

	fmt.Printf("🚀 Provisioning Sovereign Mirror for MCP: %s\n", target)

	if err := os.MkdirAll(parent, 0755); err != nil {
		fmt.Printf("Failed to create ZC directory: %v\n", err)
		return
	}

	if _, err := os.Stat(target); err == nil {
		fmt.Println("✅ Mirror already exists. Use 'update' to sync.")
		return
	}

	cmd := exec.Command("git", "clone", mcpRepo, target)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		fmt.Printf("❌ Failed to clone MCP repo: %v\n", err)
		return
	}

	fmt.Println("✅ MCP Mirror installed successfully.")

	// Create shim
	binDir := filepath.Join(root, "OlympusForge", "90000-Enablement-Labs", "000-Tools", "000-bin")
	shimPath := filepath.Join(binDir, "mcp-inspector.cmd")

	// Note: We'd need to build the inspector binary first if we wanted a direct shim,
	// but for now we'll shim the go run command.
	content := fmt.Sprintf("@echo off\ngo run \"%%~dp0../../../../OlympusForge/%s/mcp-go/test/inspector/main.go\" %%*\n", zcDir)
	os.WriteFile(shimPath, []byte(content), 0755)

	fmt.Printf("✅ Shim created: %s\n", shimPath)
}

func runUpdate() {
	_, target := getPaths()
	fmt.Printf("🔄 Updating Sovereign Mirror: %s\n", target)

	cmd := exec.Command("git", "pull")
	cmd.Dir = target
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		fmt.Printf("❌ Update failed: %v\n", err)
		return
	}

	fmt.Println("✅ MCP Mirror updated.")
}
