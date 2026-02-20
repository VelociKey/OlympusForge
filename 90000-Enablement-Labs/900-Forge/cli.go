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
	flag.Parse()

	if *workspace == "" && *target != "all" {
		log.Fatal("workspace is required unless target is 'all'")
	}

	ctx := context.Background()
	forge := &AihubForge{}

	fmt.Printf("🚀 Starting AihubForge Build [target=%s, workspace=%s]\n", *target, *workspace)
	if err := forge.Build(ctx, *target, *workspace); err != nil {
		log.Fatalf("❌ Build failed: %v", err)
	}
	fmt.Println("✅ Build completed successfully")
}
