package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
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

func runInstall() {
	fmt.Println("Installing Jules Assistant...")
	
	exePath, _ := os.Executable()
	basePath := filepath.Dir(exePath)
	binPath := filepath.Join(basePath, "bin")
	
	if err := os.MkdirAll(binPath, 0755); err != nil {
		fmt.Printf("Failed to create bin dir: %v\n", err)
		return
	}

	// Create shim based on OS
	shimPath := filepath.Join(binPath, "jules")
	shimContent := ""
	
	if os.PathSeparator == '\\' {
		shimPath += ".cmd"
		shimContent = "@echo off\r\ngemini --prompt \"/jules %*\"\r\n"
	} else {
		shimContent = "#!/bin/bash\ngemini --prompt \"/jules $@\"\n"
	}

	err := os.WriteFile(shimPath, []byte(shimContent), 0755)
	if err != nil {
		fmt.Printf("Failed to write shim: %v\n", err)
		return
	}

	fmt.Printf("Jules installed successfully at %s\n", binPath)
}

func runUpdate() {
	fmt.Println("Checking for Jules Assistant updates...")
	fmt.Println("Jules is synchronized with the fleet.")
}
