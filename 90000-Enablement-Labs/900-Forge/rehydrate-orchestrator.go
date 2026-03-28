package main

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"dagger.io/dagger"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func parseLocalTargets(path string) ([]string, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var targets []string
	scanner := bufio.NewScanner(file)
	inLocalBuilds := false

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		if strings.HasPrefix(line, "LocalBuilds") {
			inLocalBuilds = true
			continue
		}

		if inLocalBuilds {
			if line == "];" || line == "]" {
				inLocalBuilds = false
				break
			}
			// Extract "path"
			target := strings.Trim(line, "\", ")
			if target != "" {
				targets = append(targets, target)
			}
		}
	}

	return targets, scanner.Err()
}

func run() error {
	ctx := context.Background()

	// Load targets from manifest
	manifestPath := "C0100-Configuration-Registry/REHYDRATE_MANIFEST.jebnf"
	targets, err := parseLocalTargets(manifestPath)
	if err != nil {
		return fmt.Errorf("failed to load local targets: %w", err)
	}

	client, err := dagger.Connect(ctx, dagger.WithLogOutput(os.Stdout))
	if err != nil {
		return err
	}
	defer client.Close()

	// Use relative context root
	src := client.Host().Directory(".")

	// Hardened Builder Container
	builder := client.Container().
		From("golang:1.23-alpine").
		WithDirectory("/src", src).
		WithWorkdir("/src").
		WithEnvVariable("CGO_ENABLED", "0").
		WithEnvVariable("GO111MODULE", "on").
		WithEnvVariable("GOWORK", "/src/go.work")

	for _, t := range targets {
		fmt.Printf("🔨 Dagger Rehydrating: %s\n", t)
		name := filepath.Base(t)
		builder = builder.WithExec([]string{"go", "build", "-ldflags=-s -w", "-trimpath", "-o", "/bin/" + name + ".exe", t})
	}

	_, err = builder.Directory("/bin").Export(ctx, "00SDLC/OlympusForge/82000-Toolchain-Fleet/bin")
	return err
}
