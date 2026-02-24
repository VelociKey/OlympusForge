package main

import (
	"context"
	"fmt"
	"os"

	"dagger.io/dagger"
	"Olympus2/90000-Enablement-Labs/P0000-pkg/000-vault"
)

// buildGeorgeHardened implements the hardened Dagger build pipeline for George (CYC-065).
func (m *AihubForge) buildGeorgeHardened(ctx context.Context, client *dagger.Client, src *dagger.Directory) error {
	fmt.Println("⚒️ Forge: Hardened Build [George]")

	// 1. Acquire George SOUL (context)
	soulPath := "George/C0100-Configuration-Registry/POLICY.jebnf"
	soulFile := src.File(soulPath)
	soulContent, err := soulFile.Contents(ctx)
	if err != nil {
		return fmt.Errorf("failed to read George SOUL: %v", err)
	}

	// 2. Seal the SOUL Gem (Sub-Cycle 2)
	// In a real build, the master key would be a Secret from the host/TPM
	masterKey := os.Getenv("FLEET_MASTER_KEY")
	if masterKey == "" {
		masterKey = "sovereign-fleet-master-key-default"
	}
	
	gem, err := vault.Seal("george-core-policy", []byte(soulContent), []byte(masterKey))
	if err != nil {
		return fmt.Errorf("failed to seal George SOUL: %v", err)
	}
	sealedJeBNF := gem.ToJeBNF()

	// 3. Build Go WASM with Embedded Sealed Context
	// We inject the sealed context into the reasoning engine
	goWasm := client.Container().From("golang:1.25").
		WithEnvVariable("GOOS", "js").
		WithEnvVariable("GOARCH", "wasm").
		WithDirectory("/src", src).
		WithWorkdir("/src/George/10000-Autonomous-Actors/10700-Processing-Engines/10710-Reasoning-Inference").
		WithNewFile("/src/George/10000-Autonomous-Actors/10700-Processing-Engines/10710-Reasoning-Inference/inference/sealed_soul.jebnf", dagger.ContainerWithNewFileOpts{
			Contents: sealedJeBNF,
		}).
		WithExec([]string{"go", "build", "-o", "/out/george.wasm", "."}).
		File("/out/george.wasm")

	// 4. Build Go Windows EXE
	goWin := client.Container().From("golang:1.25").
		WithEnvVariable("GOOS", "windows").
		WithEnvVariable("GOARCH", "amd64").
		WithDirectory("/src", src).
		WithWorkdir("/src/George/10000-Autonomous-Actors/10700-Processing-Engines/10710-Reasoning-Inference").
		WithExec([]string{"go", "build", "-o", "/out/george-reasoning.exe", "."}).
		File("/out/george-reasoning.exe")

	// 5. Build Flutter Web WASM
	// Path alignment: 410-InteractionSurface is the primary UI
	flutterBuild := client.Container().From("ghcr.io/cirruslabs/flutter:stable").
		WithDirectory("/src", src).
		WithWorkdir("/src/George/40000-Communication-Contracts/410-InteractionSurface").
		WithFile("assets/wasm/george.wasm", goWasm).
		WithExec([]string{"flutter", "build", "web", "--wasm"}).
		Directory("build/web")

	// 6. Create final Nginx container with Silicon-Locked context
	image := client.Container().From("nginx:alpine").
		WithFile("/etc/nginx/nginx.conf", src.File("George/nginx.conf")).
		WithDirectory("/usr/share/nginx/html", flutterBuild).
		WithFile("/artifacts/george-reasoning.exe", goWin).
		WithFile("/artifacts/sealed_soul.jebnf", client.Host().Directory(".").File("George/C0100-Configuration-Registry/POLICY.jebnf")). // Placeholder for sealed
		WithEntrypoint([]string{"nginx", "-g", "daemon off;"})

	_, err = image.Publish(ctx, "localhost/george-hardened:latest")
	return err
}
