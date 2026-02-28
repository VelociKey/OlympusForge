package main

import (
	"context"
	"path/filepath"
)

type Olympusforge struct{}

// BuildGo compiles Go source from the provided directory into a Windows binary.
func (m *Olympusforge) BuildGo(ctx context.Context, src *dagger.Directory, path string) *dagger.File {
	return dag.Container().
		From("golang:1.25-alpine").
		WithDirectory("/src", src).
		WithWorkdir("/src").
		WithExec([]string{"go", "build", "-o", "/out/app.exe", path}).
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
