package main

import (
	"context"
	"fmt"
	"path/filepath"

	"dagger.io/dagger"
)

// buildRustProof demonstrates Dagger-orchestrated Rust build for CYC-108.
func (m *AihubForge) buildRustProof(ctx context.Context, client *dagger.Client, src *dagger.Directory) error {
	fmt.Println("⚒️ Forge: ADLC Phase 0 - Rust Proof [CYC-108]")

	rustWorkspace := "20POSI/George/90000-Enablement-Labs/ADLC-Proof/rust-hello"

	// 1. Define Rust Builder Container
	// We use the official Rust image as our base.
	builder := client.Container().From("rust:1.85-slim-bookworm").
		WithDirectory("/src", src).
		WithWorkdir("/src/" + rustWorkspace)

	// 2. Execute Cargo Build
	// This proves that Dagger can manage the Rust toolchain and dependencies.
	fmt.Println("⏳ Forge: Executing 'cargo build' in Dagger container...")
	build := builder.WithExec([]string{"cargo", "build", "--release"})

	// 3. Execute Tests
	// Verification is mandatory.
	fmt.Println("⏳ Forge: Executing 'cargo test' in Dagger container...")
	test := build.WithExec([]string{"cargo", "test"})

	// 4. Extract Binary
	binary := test.File("target/release/rust-hello")

	// 5. Verify Binary Execution (Final Proof)
	fmt.Println("⏳ Forge: Verifying Rust binary execution...")
	output, err := client.Container().From("debian:bookworm-slim").
		WithFile("/bin/rust-hello", binary).
		WithExec([]string{"/bin/rust-hello"}).
		Stdout(ctx)

	if err != nil {
		return fmt.Errorf("rust-hello execution failed: %v", err)
	}

	fmt.Printf("✅ Rust Proof Output: %s\n", output)

	// 6. Export Binary to Host for inspection
	exportPath := filepath.Join(rustWorkspace, "bin/rust-hello")
	_, err = test.Directory("target/release").Export(ctx, filepath.Join(rustWorkspace, "target/release"))
	if err != nil {
		return fmt.Errorf("failed to export rust artifacts: %v", err)
	}

	fmt.Printf("✅ Rust Proof Successful. Binary exported to %s\n", exportPath)
	return nil
}
