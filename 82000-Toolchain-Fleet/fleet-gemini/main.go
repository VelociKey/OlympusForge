package main

import (
	"fmt"
	"os"
	"os/exec"
)

func main() {
	// The target node script for gemini-cli
	targetScript := "C:\\aAntigravitySpace\\00SDLC\\OlympusForge\\930-System-Enablers\\920-Intelligence-Labs\\gemini-cli\\node_modules\\@google\\gemini-cli\\dist\\index.js"

	nodeExe, err := exec.LookPath("node")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: 'node' executable not found in PATH.\n")
		os.Exit(1)
	}

	cmdArgs := append([]string{targetScript}, os.Args[1:]...)
	cmd := exec.Command(nodeExe, cmdArgs...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin

	if err := cmd.Run(); err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			os.Exit(exitErr.ExitCode())
		}
		os.Exit(1)
	}
}
