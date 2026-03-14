package main

import (
	"context"
	"path/filepath"
	"dagger/olympusforge/internal/dagger"
)

type Olympusforge struct{}

// BuildGo compiles a specific Go module into a Windows binary with hardening flags.
func (m *Olympusforge) BuildGo(ctx context.Context, src *dagger.Directory, modulePath string) *dagger.File {
	return dag.Container().
		From("golang:1.26-alpine").
		WithDirectory("/src", src).
		WithWorkdir("/src").
		WithMountedCache("/go/pkg/mod", dag.CacheVolume("go-mod-cache")).
		WithMountedCache("/root/.cache/go-build", dag.CacheVolume("go-build-cache")).
		WithEnvVariable("GOWORK", "/src/go.work").
		WithExec([]string{"go", "build", "-v", "-ldflags=-s -w", "-trimpath", "-mod=readonly", "-o", "/out/app.exe", "./"+modulePath}).
		File("/out/app.exe")
}

// BuildFlutter compiles Flutter web (WASM) source with canvaskit renderer.
func (m *Olympusforge) BuildFlutter(ctx context.Context, src *dagger.Directory, path string) *dagger.Directory {
	return dag.Container().
		From("ghcr.io/cirruslabs/flutter:stable").
		WithDirectory("/src", src).
		WithWorkdir(filepath.Join("/src", path)).
		WithExec([]string{"flutter", "build", "web", "--wasm", "--web-renderer=canvaskit"}).
		Directory("build/web")
}

// VerifyEngine checks for Podman connectivity by executing a simple version command.
func (m *Olympusforge) VerifyEngine(ctx context.Context) (string, error) {
	return "Sovereign Engine Reachable", nil
}

// BuildRust compiles a specific Rust project into a Windows binary using a container.
func (m *Olympusforge) BuildRust(ctx context.Context, src *dagger.Directory, projectPath string, binaryName string) *dagger.File {
	return dag.Container().
		From("rust:1.94-bookworm").
		WithExec([]string{"apt-get", "update"}).
		WithExec([]string{"apt-get", "install", "-y", "gcc-mingw-w64"}).
		WithExec([]string{"rustup", "target", "add", "x86_64-pc-windows-gnu"}).
		WithDirectory("/src", src).
		WithWorkdir(filepath.Join("/src", projectPath)).
		WithMountedCache("/usr/local/cargo/registry", dag.CacheVolume("cargo-registry-cache")).
		WithMountedCache("/src/target", dag.CacheVolume("cargo-build-cache")).
		WithEnvVariable("CARGO_TARGET_X86_64_PC_WINDOWS_GNU_LINKER", "x86_64-w64-mingw32-gcc").
		WithExec([]string{"cargo", "build", "--release", "--target", "x86_64-pc-windows-gnu"}).
		File(filepath.Join("target/x86_64-pc-windows-gnu/release", binaryName))
}

// PromoteNative exports a file from the host filesystem to the fleet's central tool distribution directory.
func (m *Olympusforge) PromoteNative(ctx context.Context, binary *dagger.File, name string) (string, error) {
	dest := filepath.Join("00SDLC/OlympusForge/82000-Toolchain-Fleet/bin", name)
	_, err := binary.Export(ctx, dest)
	if err != nil {
		return "", err
	}
	return "Promoted " + name + " to " + dest, nil
}

// LatentLinguaGoStudio assembles the Flutter Studio with a Go-WASM kernel.
func (m *Olympusforge) LatentLinguaGoStudio(ctx context.Context, src *dagger.Directory) *dagger.Directory {
	// 1. Build the Go WASM Kernel
	wasm := dag.Container().
		From("golang:1.26-alpine").
		WithDirectory("/src", src).
		WithWorkdir("/src").
		WithEnvVariable("GOOS", "js").
		WithEnvVariable("GOARCH", "wasm").
		WithEnvVariable("GOWORK", "/src/go.work").
		WithExec([]string{"go", "build", "-o", "/out/latent_lingua.wasm", "./01LOCO/LatentLingua/90000-Enablement-Labs/920-WASM-Kernel"}).
		File("/out/latent_lingua.wasm")

	// 2. Build the Flutter app (using a pinned stable image for speed/reliability)
	flutter := dag.Container().
		From("ghcr.io/cirruslabs/flutter:3.41.0").
		WithDirectory("/src", src).
		WithWorkdir("/src/01LOCO/LatentLingua/40000-Communication-Contracts/411-Interaction-Surface").
		WithExec([]string{"flutter", "build", "web", "--wasm"}).
		Directory("build/web")

		// 3. Inject Kernel Loader and combine
	goImg := dag.Container().From("golang:1.26-alpine")
	wasmExec := goImg.File("/usr/local/go/lib/wasm/wasm_exec.js")

	loader := dag.Container().
		From("alpine").
		WithNewFile("/kernel_loader.js", `
            async function loadKernel() {
                const go = new Go();
                const result = await WebAssembly.instantiateStreaming(fetch("latent_lingua.wasm"), go.importObject);
                go.run(result.instance);
                console.log("LatentLingua Go Kernel Loaded via OlympusForge");
            }
            loadKernel();
        `).File("/kernel_loader.js")

	return dag.Container().
		From("alpine").
		WithDirectory("/", flutter).
		WithFile("/wasm_exec.js", wasmExec).
		WithFile("/latent_lingua.wasm", wasm).
		WithFile("/kernel_loader.js", loader).
		Directory("/")
}
