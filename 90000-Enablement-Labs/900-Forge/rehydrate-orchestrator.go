package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"dagger.io/dagger"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func run() error {
	ctx := context.Background()
	client, err := dagger.Connect(ctx, dagger.WithLogOutput(os.Stdout))
	if err != nil {
		return err
	}
	defer client.Close()

	src := client.Host().Directory("C:/aAntigravitySpace")

	// Hardened Builder Container
	builder := client.Container().
		From("golang:1.23-alpine").
		WithDirectory("/src", src).
		WithWorkdir("/src").
		WithEnvVariable("CGO_ENABLED", "0").
		WithEnvVariable("GO111MODULE", "on").
		WithEnvVariable("GOWORK", "/src/go.work")

	targets := []string{
		"00SDLC/OlympusConductor/10000-Autonomous-Actors/110-devpm-relay",
		"00SDLC/OlympusConductor/30100-Execution-Points/110-conductor-project",
		"00SDLC/OlympusConductor/30100-Execution-Points/120-conductor-cycle",
		"00SDLC/OlympusConductor/30100-Execution-Points/150-gemaid-harden",
	}

	for _, t := range targets {
		fmt.Printf("🔨 Dagger Rehydrating: %s\n", t)
		name := filepath.Base(t)
		builder = builder.WithExec([]string{"go", "build", "-ldflags=-s -w", "-trimpath", "-o", "/bin/" + name + ".exe", t})
	}

	_, err = builder.Directory("/bin").Export(ctx, "C:/aAntigravitySpace/00SDLC/OlympusForge/82000-Toolchain-Fleet/bin")
	return err
}
