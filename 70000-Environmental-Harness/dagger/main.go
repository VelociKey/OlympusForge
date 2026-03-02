package main

import (
	"context"
	"path/filepath"
	"dagger/olympusforge/internal/dagger"
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

func (m *Olympusforge) HelloWorld(ctx context.Context) string { return "Hello from OlympusForge!" }
