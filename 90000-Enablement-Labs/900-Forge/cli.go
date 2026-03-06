package main

import (
	"context"
	"flag"
	"fmt"
	"log"
)

func maincli() {
	target := flag.String("target", "native", "build target (native, podman, gcp, all)")
	workspace := flag.String("workspace", "", "workspace to build")
	assess := flag.Bool("assess", false, "run maturity assessment instead of build")
	rustProof := flag.Bool("rust-proof", false, "run ADLC Phase 0 Rust Proof (CYC-108)")
	flag.Parse()

	if !*rustProof && *workspace == "" && *target != "all" && *workspace != "all" {
		log.Fatal("workspace is required unless target is 'all' or --rust-proof is set")
	}

	if *target == "gcp" {
		fmt.Printf("⚠️  CAUTION: You are about to trigger a GCP Cloud Build for workspace: %s\n", *workspace)
		if !confirm("Are you sure you want to proceed?") {
			log.Fatal("Build aborted by user (Verification 1)")
		}
		if !confirm("Are you REALLY sure? This will push images to the production Artifact Registry.") {
			log.Fatal("Build aborted by user (Verification 2)")
		}
	}

	ctx := context.Background()
	forge := &AihubForge{}

	if *rustProof {
		fmt.Println("🦀 Starting ADLC Phase 0 Rust Proof (CYC-108)...")
		if err := forge.RustProof(ctx); err != nil {
			log.Fatalf("❌ Rust Proof failed: %v", err)
		}
		fmt.Println("✅ Rust Proof completed successfully")
		return
	}

	if *assess {
		fmt.Printf("🛡️ Starting Athena Maturity Assessment for workspace: %s\n", *workspace)
		if err := forge.Assess(ctx, *workspace); err != nil {
			log.Fatalf("❌ Assessment failed: %v", err)
		}
		fmt.Println("✅ Assessment completed successfully")
		return
	}

	fmt.Printf("🚀 Starting AihubForge Build [target=%s, workspace=%s]\n", *target, *workspace)
	if err := forge.Build(ctx, *target, *workspace); err != nil {
		log.Fatalf("❌ Build failed: %v", err)
	}
	fmt.Println("✅ Build completed successfully")
}

func confirm(prompt string) bool {
	fmt.Printf("🛡️  %s [y/N]: ", prompt)
	var response string
	fmt.Scanln(&response)
	return response == "y" || response == "Y"
}
