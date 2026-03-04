package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"sync"
	"syscall"
)

type Service struct {
	Name    string
	SrcDir  string
	BinPath string
	Args    []string
}

func main() {
	fmt.Println("Olympus Fleet Mesh-Up: Orchestrating Sovereign Mesh...")

	rootDir, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}

	// Define all mesh services
	services := []Service{
		{
			Name:    "MCP-Gateway",
			SrcDir:  "00SDLC/OlympusMCP/10000-Autonomous-Actors/MCPGateway",
			BinPath: "olympus.fleet/00SDLC/OlympusForge/82000-Toolchain-Fleet/mcp-gateway.exe",
			Args:    []string{"-backend", "http://localhost:8091"}, // Default to SDLCBackend
		},
		{
			Name:    "SDLC-Backend",
			SrcDir:  "00SDLC/OlympusMCP/10000-Autonomous-Actors/SDLCBackend",
			BinPath: "00SDLC/OlympusMCP/10000-Autonomous-Actors/SDLCBackend/SDLCBackend.exe",
		},
		{ Name: "Vault-Manager", SrcDir: "00SDLC/OlympusGCP-Vault/10000-Autonomous-Actors/VaultManager", BinPath: "00SDLC/OlympusGCP-Vault/10000-Autonomous-Actors/VaultManager/VaultManager.exe" },
		{ Name: "Storage-Manager", SrcDir: "00SDLC/OlympusGCP-Storage/10000-Autonomous-Actors/StorageManager", BinPath: "00SDLC/OlympusGCP-Storage/10000-Autonomous-Actors/StorageManager/StorageManager.exe" },
		{ Name: "Data-Manager", SrcDir: "00SDLC/OlympusGCP-Data/10000-Autonomous-Actors/DataManager", BinPath: "00SDLC/OlympusGCP-Data/10000-Autonomous-Actors/DataManager/DataManager.exe" },
		{ Name: "Events-Manager", SrcDir: "00SDLC/OlympusGCP-Events/10000-Autonomous-Actors/EventsManager", BinPath: "00SDLC/OlympusGCP-Events/10000-Autonomous-Actors/EventsManager/EventsManager.exe" },
		{ Name: "Compute-Manager", SrcDir: "00SDLC/OlympusGCP-Compute/10000-Autonomous-Actors/ComputeManager", BinPath: "00SDLC/OlympusGCP-Compute/10000-Autonomous-Actors/ComputeManager/ComputeManager.exe" },
		{ Name: "Intelligence-Manager", SrcDir: "00SDLC/OlympusGCP-Intelligence/10000-Autonomous-Actors/IntelligenceManager", BinPath: "00SDLC/OlympusGCP-Intelligence/10000-Autonomous-Actors/IntelligenceManager/IntelligenceManager.exe" },
		{ Name: "Messaging-Manager", SrcDir: "00SDLC/OlympusGCP-Messaging/10000-Autonomous-Actors/MessagingManager", BinPath: "00SDLC/OlympusGCP-Messaging/10000-Autonomous-Actors/MessagingManager/MessagingManager.exe" },
		{ Name: "Observability-Manager", SrcDir: "00SDLC/OlympusGCP-Observability/10000-Autonomous-Actors/ObservabilityManager", BinPath: "00SDLC/OlympusGCP-Observability/10000-Autonomous-Actors/ObservabilityManager/ObservabilityManager.exe" },
		{ Name: "FinOps-Manager", SrcDir: "00SDLC/OlympusGCP-FinOps/10000-Autonomous-Actors/FinOpsManager", BinPath: "00SDLC/OlympusGCP-FinOps/10000-Autonomous-Actors/FinOpsManager/FinOpsManager.exe" },
		{ Name: "Firebase-Manager", SrcDir: "00SDLC/OlympusGCP-Firebase/10000-Autonomous-Actors/FirebaseManager", BinPath: "00SDLC/OlympusGCP-Firebase/10000-Autonomous-Actors/FirebaseManager/FirebaseManager.exe" },
	}

	// Step 1: Build all services
	fmt.Println("Step 1: Rebuilding Mesh Binaries (Go 1.26.0)...")
	for i := range services {
		fmt.Printf("   Compiling %s...\n", services[i].Name)
		cmd := exec.Command("go", "build", "-o", filepath.Base(services[i].BinPath), "main.go")
		cmd.Dir = filepath.Join(rootDir, services[i].SrcDir)
		if output, err := cmd.CombinedOutput(); err != nil {
			log.Fatalf("Failed to build %s: %v\nOutput: %s", services[i].Name, err, string(output))
		}
		// Ensure BinPath is absolute for execution
		services[i].BinPath = filepath.Join(rootDir, services[i].SrcDir, filepath.Base(services[i].BinPath))
	}

	// Step 2: Start all services
	fmt.Println("Step 2: Launching Mesh Node Cluster...")
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var wg sync.WaitGroup
	for _, svc := range services {
		wg.Add(1)
		go func(s Service) {
			defer wg.Done()
			runService(ctx, s)
		}(svc)
	}

	// Wait for interrupt
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)
	<-sigCh

	fmt.Println("\nShutdown signal received. Terminating mesh...")
	cancel()
	wg.Wait()
	fmt.Println("Mesh cluster offline.")
}

func runService(ctx context.Context, svc Service) {
	cmd := exec.CommandContext(ctx, svc.BinPath, svc.Args...)
	cmd.Dir = filepath.Dir(svc.BinPath)
	cmd.Env = os.Environ()
	cmd.Env = append(cmd.Env, "ANTIGRAVITY_ROOT="+os.Getenv("ANTIGRAVITY_ROOT"))

	stdout, _ := cmd.StdoutPipe()
	stderr, _ := cmd.StderrPipe()

	if err := cmd.Start(); err != nil {
		fmt.Printf(" [%s] Start failed: %v\n", svc.Name, err)
		return
	}

	fmt.Printf(" [%s] Online (PID: %d)\n", svc.Name, cmd.Process.Pid)

	// Stream logs with prefixes
	go streamLogs(svc.Name, stdout)
	go streamLogs(svc.Name, stderr)

	<-ctx.Done()
	// Process will be killed by context cancellation
	fmt.Printf(" [%s] Stopping...\n", svc.Name)
}

func streamLogs(name string, r io.Reader) {
	buf := make([]byte, 1024)
	for {
		n, err := r.Read(buf)
		if n > 0 {
			fmt.Printf("[%s] %s", name, string(buf[:n]))
		}
		if err != nil {
			return
		}
	}
}
