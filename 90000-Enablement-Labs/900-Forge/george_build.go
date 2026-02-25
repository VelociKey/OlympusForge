package main

import (
	"context"
	"fmt"
	"os"

	"dagger.io/dagger"
	"Olympus2/90000-Enablement-Labs/P0000-pkg/000-vault"
)

// buildGeorgeHardened implements the Pure-Wasm architecture for George.
// It embeds the Go reasoning engine into the Flutter WasmGC payload.
func (m *AihubForge) buildGeorgeHardened(ctx context.Context, client *dagger.Client, src *dagger.Directory) error {
	fmt.Println("⚒️ Forge: Pure-Wasm Hardened Build [George]")

	// 1. Acquire George SOUL (context)
	soulPath := "George/C0100-Configuration-Registry/POLICY.jebnf"
	soulFile := src.File(soulPath)
	soulContent, err := soulFile.Contents(ctx)
	if err != nil {
		return fmt.Errorf("failed to read George SOUL: %v", err)
	}

	// 2. Seal the SOUL Gem
	masterKey := os.Getenv("FLEET_MASTER_KEY")
	if masterKey == "" {
		masterKey = "sovereign-fleet-master-key-default"
	}
	
	gem, err := vault.Seal("george-core-policy", []byte(soulContent), []byte(masterKey))
	if err != nil {
		return fmt.Errorf("failed to seal George SOUL: %v", err)
	}
	sealedJeBNF := gem.ToJeBNF()

	// 3a. Build Go WASM Engine (with embedded sealed context)
	// This is the core reasoning logic that will run in the browser.
	goWasm := client.Container().From("golang:"+FleetGoVersion).
		WithEnvVariable("GOOS", "js").
		WithEnvVariable("GOARCH", "wasm").
		WithDirectory("/src", src).
		WithNewFile("/src/go.work", m.minimalGoWork("George")).
		WithEnvVariable("GOWORK", "/src/go.work").
		WithWorkdir("/src/George/10000-Autonomous-Actors/10700-Processing-Engines/10710-Reasoning-Inference").
		WithNewFile("/src/George/10000-Autonomous-Actors/10700-Processing-Engines/10710-Reasoning-Inference/inference/sealed_soul.jebnf", sealedJeBNF).
		WithExec([]string{"go", "build", "-o", "/out/george.wasm", "."}).
		File("/out/george.wasm")

	// 3b. Build Go Linux Backend (The CDE Binary)
	// This is the server-side component for the CDE.
	goLinux := client.Container().From("golang:"+FleetGoVersion+"-alpine").
		WithDirectory("/src", src).
		WithNewFile("/src/go.work", m.minimalGoWork("George")).
		WithEnvVariable("GOWORK", "/src/go.work").
		WithWorkdir("/src/George/10000-Autonomous-Actors/10700-Processing-Engines/10710-Reasoning-Inference").
		WithExec([]string{"go", "build", "-o", "/out/george-reasoning-linux", "."}).
		File("/out/george-reasoning-linux")

	// 4. Build Flutter Web WASM (Embedding the Go Engine)
	// We inject george.wasm into assets/wasm so Flutter can load it.
	fmt.Println("⏳ Forge: Go Backend Compiled. Starting Flutter Web Build (Standard Dagger Image)...")
	
	// Use standard Flutter image (Dagger-managed)
	flutterBuild := client.Container().From("ghcr.io/cirruslabs/flutter:stable").
		WithDirectory("/src", src).
		WithWorkdir("/src/George/40000-Communication-Contracts/410-InteractionSurface").
		WithExec([]string{"mkdir", "-p", "assets/wasm"}).
		WithFile("assets/wasm/george.wasm", goWasm).
		WithExec([]string{"flutter", "config", "--enable-web"}).
		WithExec([]string{"flutter", "pub", "get"}).
		WithExec([]string{"flutter", "pub", "upgrade"}). // Handle dependency updates
		WithExec([]string{"flutter", "build", "web", "--wasm"}).
		Directory("build/web")

	// 5. Create final "Shell" Image
	// A lightweight Nginx container that serves the static Wasm artifacts.
	// It relies on host-mounted Ollama for intelligence.
	image := client.Container().From("nginx:alpine").
		WithExec([]string{"apk", "--no-cache", "add", "ca-certificates", "curl"}).
		WithFile("/etc/nginx/nginx.conf", src.File("George/nginx.conf")).
		WithDirectory("/usr/share/nginx/html", flutterBuild).
		WithFile("/entrypoint.sh", src.File("George/entrypoint.sh")).
		WithExec([]string{"chmod", "+x", "/entrypoint.sh"}).
		// CDE Injection: Add Backend Binary
		WithDirectory("/artifacts", client.Directory()).
		WithFile("/artifacts/george-reasoning-linux", goLinux).
		WithExec([]string{"chmod", "+x", "/artifacts/george-reasoning-linux"}).
		// CDE Injection: Add Source Code
		WithDirectory("/src/go", src, dagger.ContainerWithDirectoryOpts{
			Include: []string{
				"go.work",
				"George/",
			},
			Exclude: []string{
				"George/40000-Communication-Contracts/410-InteractionSurface/build",
				"George/40000-Communication-Contracts/410-InteractionSurface/.dart_tool",
				"George/.git",
			},
		}).
		WithEntrypoint([]string{"/entrypoint.sh"})

	// Export to tarball for local loading (bypasses registry requirement)
	_, err = image.Export(ctx, "george.tar")
	if err == nil {
		fmt.Println("✅ Image exported to george.tar. Run 'podman load -i george.tar' to import.")
	}
	return err
}
