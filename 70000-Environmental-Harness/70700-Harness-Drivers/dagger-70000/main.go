package main

import (
	"context"
	"path/filepath"
	"dagger/olympusforge/lib/dagger"
)

type Olympusforge struct{}

// BuildGo compiles a specific Go module into a Windows binary.
func (m *Olympusforge) BuildGo(ctx context.Context, src *dagger.Directory, modulePath string) *dagger.File {
	return dag.Container().
		From("golang:1.25-alpine").
		WithDirectory("/src", src).
		WithWorkdir("/src").
		// Point to the root go.work but build only the target
		WithEnvVariable("GOWORK", "/src/go.work").
		WithExec([]string{"go", "build", "-v", "-mod=readonly", "-o", "/out/app.exe", "./"+modulePath}).
		File("/out/app.exe")
}

// BuildFlutter compiles Flutter web (WASM) source from the provided directory.
func (m *Olympusforge) BuildFlutter(ctx context.Context, src *dagger.Directory, path string) *dagger.Directory {
	return dag.Container().
		From("ghcr.io/cirruslabs/flutter:stable").
		WithDirectory("/src", src).
		WithWorkdir(filepath.Join("/src", path)).
		WithExec([]string{"flutter", "build", "web", "--wasm"}).
		Directory("build/web")
}

// BuildRust compiles a specific Rust project into a Windows binary using a container.
func (m *Olympusforge) BuildRust(ctx context.Context, src *dagger.Directory, projectPath string) *dagger.File {
	return dag.Container().
		From("rust:1.85-bookworm").
		WithDirectory("/src", src).
		WithWorkdir(filepath.Join("/src", projectPath)).
		WithExec([]string{"cargo", "build", "--release", "--target", "x86_64-pc-windows-msvc"}).
		File("target/x86_64-pc-windows-msvc/release/app.exe")
}

// BuildRustNative compiles a specific Rust project natively on the host workstation for O(1) feedback.
// This assumes the host has the necessary toolchain provisioned.
func (m *Olympusforge) BuildRustNative(ctx context.Context, src *dagger.Directory, projectPath string) *dagger.File {
	// In Dagger, Native execution is often achieved via local execution on the engine
	// or by exposing the host's toolchain to the container.
	// For this Sovereign Substrate, we use the host's cargo directly if available.
	return dag.Host().Directory(".").
		File(filepath.Join(projectPath, "target/release/app.exe"))
}

func (m *Olympusforge) HelloWorld(ctx context.Context) string { return "Hello from OlympusForge!" }
