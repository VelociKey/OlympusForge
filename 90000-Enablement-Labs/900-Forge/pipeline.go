package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"dagger.io/dagger"
)

// AihubForge is the central Sovereign Factory for the Olympus2 fleet.
// It provides deterministic build pipelines for workstation-native, containerized, and cloud-native targets.
type AihubForge struct{}

// Build triggers the build pipeline with the specified target and workspace.
// Targets: native (workstation), podman (local OCI), gcp (Google Artifact Registry)
func (m *AihubForge) Build(ctx context.Context, target string, workspace string) error {
	client, err := dagger.Connect(ctx, dagger.WithLogOutput(os.Stdout))
	if err != nil {
		return fmt.Errorf("failed to connect to dagger: %v", err)
	}
	defer client.Close()

	// 1. Acquire Source from Fleet Root (relative to Forge location)
	fmt.Println("🔍 Forge: Acquiring source from relative root ../../..")
	src := client.Host().Directory("../../..", dagger.HostDirectoryOpts{
		Exclude: []string{"C0400-Artifacts", "C0990-Scratch", ".git", "node_modules", ".gemini/tmp", "tmp_builds", "tmp_bin"},
	})

	// 2. Multi-Workspace Dispatch
	if workspace == "all" {
		return m.BuildAllClusters(ctx, target)
	}

	// 3. Dispatch to Target Strategy
	switch target {
	case "native":
		return m.buildNative(ctx, client, src, workspace)
	case "podman":
		return m.buildPodman(ctx, client, src, workspace)
	case "gcp":
		return m.buildGCP(ctx, client, src, workspace)
	default:
		return fmt.Errorf("invalid target: %s. Options: native, podman, gcp, all", target)
	}
}

// BuildAllClusters iterates over all 10 GCP clusters and builds them for the specified target.
func (m *AihubForge) BuildAllClusters(ctx context.Context, target string) error {
	clusters := []string{
		"OlympusGCP-Compute",
		"OlympusGCP-Data",
		"OlympusGCP-Events",
		"OlympusGCP-FinOps",
		"OlympusGCP-Firebase",
		"OlympusGCP-Intelligence",
		"OlympusGCP-Messaging",
		"OlympusGCP-Observability",
		"OlympusGCP-Storage",
		"OlympusGCP-Vault",
	}
	for _, cluster := range clusters {
		if err := m.Build(ctx, target, cluster); err != nil {
			return err
		}
	}
	return nil
}

// buildNative builds binaries directly on the host or in a host-mirrored environment
func (m *AihubForge) buildNative(ctx context.Context, client *dagger.Client, src *dagger.Directory, workspace string) error {
	fmt.Printf("⚒️ Forge: Native Build [%s]\n", workspace)

	// Build for host OS/Arch (Windows cross-compile since we are in a Linux container)
	// We use sh -c to find and build the first main.go in the workspace
	builder := client.Container().From("golang:1.25").
		WithEnvVariable("GOOS", "windows").
		WithEnvVariable("GOARCH", "amd64").
		WithEnvVariable("GOWORK", "/src/go.work").
		WithDirectory("/src", src).
		WithWorkdir("/src/" + workspace).
		WithExec([]string{"sh", "-c", "MAIN_PATH=$(find . -name main.go | head -n 1); if [ -z \"$MAIN_PATH\" ]; then echo 'ERROR: No main.go found in '\"$(pwd)\"; exit 1; fi; go build -o bin/service.exe $MAIN_PATH"})

	// Export binary back to host
	_, err := builder.Directory("bin").Export(ctx, filepath.Join(workspace, "bin"))
	return err
}

// buildPodman builds OCI images for local Podman Desktop execution
func (m *AihubForge) buildPodman(ctx context.Context, client *dagger.Client, src *dagger.Directory, workspace string) error {
	fmt.Printf("⚒️ Forge: Podman Image Build [%s]\n", workspace)

	image := m.sealedImage(client, src, workspace)

	// Tag for local Podman
	tag := fmt.Sprintf("localhost/%s:latest", workspace)
	_, err := image.Publish(ctx, tag)
	return err
}

// buildGCP builds and publishes images to Google Artifact Registry
func (m *AihubForge) buildGCP(ctx context.Context, client *dagger.Client, src *dagger.Directory, workspace string) error {
	return m.PublishService(ctx, client, src, workspace)
}

// PublishService tags and pushes the image to Google Artifact Registry.
func (m *AihubForge) PublishService(ctx context.Context, client *dagger.Client, src *dagger.Directory, workspace string) error {
	// Double-Verification Safeguard
	if os.Getenv("GAR_PUSH_CONFIRMED") != "true" {
		return fmt.Errorf("ABORTING PUSH: GAR Image Push requires double-verification. Please set GAR_PUSH_CONFIRMED=true to proceed")
	}

	projectID := os.Getenv("GCP_PROJECT_ID")
	if projectID == "" {
		projectID = "olympus-project"
	}
	garRegion := os.Getenv("GCP_GAR_REGION")
	if garRegion == "" {
		garRegion = "us-east1"
	}

	fmt.Printf("⚒️ Forge: GCP Cloud Run Build [%s] -> %s\n", workspace, projectID)

	image := m.sealedImage(client, src, workspace)

	// Tag for GAR
	tag := fmt.Sprintf("%s-docker.pkg.dev/%s/olympus-fleet/%s:latest", garRegion, projectID, strings.ToLower(workspace))
	_, err := image.Publish(ctx, tag)
	return err
}

// sealedImage creates a deterministic, attestation-ready OCI image
func (m *AihubForge) sealedImage(client *dagger.Client, src *dagger.Directory, workspace string) *dagger.Container {
	return client.Container().From("debian:trixie-slim").
		WithDirectory("/app", src.Directory(workspace)).
		WithWorkdir("/app").
		WithExec([]string{"apt-get", "update"}).
		WithExec([]string{"apt-get", "install", "-y", "ca-certificates"}).
		WithEntrypoint([]string{"./bin/service"})
}
