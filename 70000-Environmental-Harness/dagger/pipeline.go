package main

import (
	"context"
	"fmt"
	"path/filepath"

	"dagger/olympusforge/internal/dagger"
)

// Build orchestrates the build process for a given workspace.
func (m *Olympusforge) Build(ctx context.Context, src *dagger.Directory, target string, workspace string) error {
	// 1. Resolve Sovereign Paths
	// From 70000/dagger, root is 3 levels up (OlympusForge)

	switch target {
	case "native":
		return m.buildNative(ctx, src, workspace)
	case "podman":
		return fmt.Errorf("podman target not yet implemented")
	default:
		return fmt.Errorf("unknown target: %s", target)
	}
}

// buildNative cross-compiles for the Windows host using the Trixie-Native pipeline.
func (m *Olympusforge) buildNative(ctx context.Context, src *dagger.Directory, workspace string) error {
	fmt.Printf("🛡️ Trixie-Native Build: %s (Target=Windows/AMD64)\n", workspace)

	// Use Google-Clean Trixie Go Image (2026 Standard)
	goBuilder := dag.Container().From("golang:1.26-alpine").
		WithDirectory("/src", src).
		WithWorkdir(filepath.Join("/src", workspace)).
		WithEnvVariable("GOOS", "windows").
		WithEnvVariable("GOARCH", "amd64").
		WithEnvVariable("CGO_ENABLED", "0")

	// The Build Pulse
	binaryName := filepath.Base(workspace) + ".exe"
	build := goBuilder.WithExec([]string{"go", "build", "-o", "bin/" + binaryName, "."})

	// Export to the Workspace bin/ (Topographical Repatriation)
	_, err := build.Directory("bin/").Export(ctx, filepath.Join(workspace, "bin"))
	return err
}

// nvwA2UI builds the Flutter/Dart WASM frontend.
func (m *Olympusforge) nvwA2UI(ctx context.Context, target string) *dagger.Container {
	
	// Fetching the Google-Clean Flutter SDK directly (v3.41.2 Stable)
	flutterTarball := dag.HTTP("https://storage.googleapis.com/flutter_infra_release/releases/stable/linux/flutter_linux_3.41.2-stable.tar.xz")

	// Provisioning the Trixie-Native Flutter Environment
	engine := dag.Container().
		From("debian:trixie-slim").
		WithFile("/tmp/flutter.tar.xz", flutterTarball).
		WithExec([]string{"sh", "-c", "tar -xf /tmp/flutter.tar.xz -C /usr/local"}).
		WithEnvVariable("PATH", "/usr/local/flutter/bin:/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin", dagger.ContainerWithEnvVariableOpts{Expand: true}).
		WithExec([]string{"flutter", "config", "--no-analytics"}).
		WithExec([]string{"flutter", "doctor"})

	fmt.Printf("✅ nvwA2UI Sovereign Builder [Trixie] Ready.\n")
	return engine
}
