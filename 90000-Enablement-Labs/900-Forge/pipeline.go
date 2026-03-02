package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"dagger.io/dagger"
)

const FleetGoVersion = "1.26.0"
const FleetFlutterVersion = "3.41.1" // "The Year of Fire Horse" (Feb 2026)

// AihubForge is the central Sovereign Factory for the Olympus2 fleet.
// It provides deterministic build pipelines for workstation-native, containerized, and cloud-native targets.
type AihubForge struct{}

// isNoBuild checks if the workspace has a .nobuild marker
func (m *AihubForge) isNoBuild(workspace string) bool {
	path := filepath.Join(workspace, ".nobuild")
	if _, err := os.Stat(path); err == nil {
		return true
	}
	// Also check if workspace is just "." and root has .nobuild
	if workspace == "." || workspace == "" {
		if _, err := os.Stat(".nobuild"); err == nil {
			return true
		}
	}
	return false
}

// goBase returns a cached Go builder container
func (m *AihubForge) goBase(client *dagger.Client) *dagger.Container {
	return client.Container().From("golang:"+FleetGoVersion).
		WithEnvVariable("CGO_ENABLED", "0")
}

// flutterBase returns a cached Flutter builder container
func (m *AihubForge) flutterBase(client *dagger.Client) *dagger.Container {
	return client.Container().From("ghcr.io/cirruslabs/flutter:stable").
		WithExec([]string{"flutter", "config", "--enable-web"})
}

// Build triggers the build pipeline with the specified target and workspace.
// Targets: native (workstation), podman (local OCI), gcp (Google Artifact Registry)
func (m *AihubForge) Build(ctx context.Context, target string, workspace string) error {
	if m.isNoBuild(workspace) {
		fmt.Printf("⏭️  Forge: Skipping %s (.nobuild marker detected)\n", workspace)
		return nil
	}

	client, err := dagger.Connect(ctx, dagger.WithLogOutput(os.Stdout))
	if err != nil {
		return fmt.Errorf("failed to connect to dagger: %v", err)
	}
	defer client.Close()

	// 1. Acquire Source from Fleet Root (Selective Sync for speed)
	fmt.Println("🔍 Forge: Performing Selective Sync [Inclusion Mode]")
	src := client.Host().Directory("../../..", dagger.HostDirectoryOpts{
		Include: []string{
			workspace + "/**",
			"olympus.fleet/00SDLC/Olympus2/**",
			"olympus.fleet/00SDLC/OlympusForge/81000-Tools/**",                     // New Authority Tools
			"olympus.fleet/00SDLC/Olympus2/90000-Enablement-Labs/P0000-pkg/**",                       // New Authority Packages
			"olympus.fleet/00SDLC/OlympusForge/90000-Enablement-Labs/900-Forge/**", // The build tool itself
			"olympus.fleet/00SDLC/OlympusForge/bin/linux/flutter/**",               // Include Fleet Flutter SDK
			"olympus.fleet/00SDLC/OlympusGrammar/**",
			"olympus.fleet/00SDLC/OlympusAtelier/**",
		},
		Exclude: []string{
			"**/node_modules",
			"**/.git",
			"**/.gemini/tmp",
			"olympus.fleet/00SDLC/Olympus2/gen",
			"olympus.fleet/00SDLC/Olympus2/gen/**",
			"olympus.fleet/00SDLC/Olympus2/C0990-Ephemeral-Scratch", // Massive scratch space
			"olympus.fleet/00SDLC/Olympus2/C0400-Artifact-Repository",
			"**/*.exe",
			"olympus.fleet/00SDLC/OlympusForge/models", // Double-check exclusion
			"go.work",
			"go.sum",
		},
	})

	// 2. Multi-Workspace Dispatch
	if workspace == "all" {
		return m.BuildAllClusters(ctx, target)
	}

	// 3. George Restriction: Only Podman allowed for Sandbox Safety
	if workspace == "George" && target != "podman" {
		fmt.Printf("⚠️ George workspace is restricted to 'podman' target only (Sandbox Safety Mandate). Redirecting to podman...\n")
		target = "podman"
	}

	// 4. Dispatch to Target Strategy
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

// minimalGoWork returns a go.work string containing only the workspaces synced into the container.
func (m *AihubForge) minimalGoWork(workspace string) string {
	// These are the core workspaces required for almost all builds
	// OlympusForge is NOT core for runtime services (only for tooling), so we exclude it by default
	// unless the workspace itself is OlympusForge.
	core := []string{
		"Olympus2",
		"OlympusGrammar",
		"OlympusAtelier",
	}
	
	lines := []string{"go 1.25.7", "use ("}
	seen := make(map[string]bool)
	
	wss := append(core, workspace)
	for _, ws := range wss {
		if !seen[ws] {
			lines = append(lines, fmt.Sprintf("\t./%s", ws))
			seen[ws] = true
		}
	}
	// Explicitly add OlympusForge only if we are building it (or a subdir of it)
	if strings.HasPrefix(workspace, "OlympusForge") {
		lines = append(lines, "\t./OlympusForge")
	}

	lines = append(lines, ")")
	return strings.Join(lines, "\n")
}

func (m *AihubForge) Assess(ctx context.Context, workspace string) error {
	fmt.Printf("🔍 Forge: Connecting to Dagger for Assessment of %s...\n", workspace)
	client, err := dagger.Connect(ctx, dagger.WithLogOutput(os.Stdout))
	if err != nil {
		return fmt.Errorf("failed to connect to dagger: %v", err)
	}
	defer client.Close()

	fmt.Println("🔍 Forge: Acquiring source for Assessment (Selective Sync)")
	src := client.Host().Directory("../../..", dagger.HostDirectoryOpts{
		Include: []string{
			workspace + "/**",
			"olympus.fleet/00SDLC/Olympus2/**",
			"olympus.fleet/00SDLC/OlympusForge/**",
			"olympus.fleet/00SDLC/OlympusGrammar/**",
			"olympus.fleet/00SDLC/OlympusAtelier/**",
		},
		Exclude: []string{
			"**/node_modules",
			"**/.git",
			"**/.gemini/tmp",
			"olympus.fleet/00SDLC/Olympus2/gen",
			"olympus.fleet/00SDLC/Olympus2/gen/**",
			"**/*.exe",
			"go.work",
			"go.sum",
		},
	})

	if workspace == "all" {
		return m.AssessAllClusters(ctx, client, src)
	}

	targets := strings.Split(workspace, ",")
	for _, target := range targets {
		t := strings.TrimSpace(target)
		if t == "" {
			continue
		}
		if err := validateAgentMaturity(ctx, client, src, t); err != nil {
			fmt.Printf("⚠️ Assessment failed for %s: %v\n", t, err)
		}
	}
	return nil
}

func (m *AihubForge) AssessAllClusters(ctx context.Context, client *dagger.Client, src *dagger.Directory) error {
	clusters := []string{
		"George",
		"Olympus2",
		"OlympusMCP",
		"OlympusForge",
		"OlympusGCP-Events",
		"OlympusGCP-Vault",
	}
	for _, cluster := range clusters {
		if err := validateAgentMaturity(ctx, client, src, cluster); err != nil {
			fmt.Printf("⚠️ Assessment failed for %s: %v\n", cluster, err)
		}
	}
	return nil
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

// detectStack identifies if a workspace is Go or Flutter
func (m *AihubForge) detectStack(workspace string) string {
	if _, err := os.Stat(filepath.Join(workspace, "pubspec.yaml")); err == nil {
		return "flutter"
	}
	if _, err := os.Stat(filepath.Join(workspace, "go.mod")); err == nil {
		return "go"
	}
	return "unknown"
}

// buildNative builds binaries directly on the host or in a host-mirrored environment
func (m *AihubForge) buildNative(ctx context.Context, client *dagger.Client, src *dagger.Directory, workspace string) error {
	stack := m.detectStack(workspace)
	fmt.Printf("⚒️ Forge: Native Build [%s] (Stack: %s)\n", workspace, stack)

	// Load Podman Mount System configuration
	config, err := m.loadProvisionConfig(workspace)
	if err != nil {
		return fmt.Errorf("failed to load provision config: %v", err)
	}

	var builder *dagger.Container
	switch stack {
	case "go":
		builder = m.goBase(client).
			WithEnvVariable("GOOS", "windows").
			WithEnvVariable("GOARCH", "amd64").
			WithDirectory("/src", src).
			WithNewFile("/src/go.work", m.minimalGoWork(workspace)).
			WithEnvVariable("GOWORK", "/src/go.work").
			WithWorkdir("/src/"+workspace)

		// Inject Mount Envs for the build or runtime
		for _, mount := range config.Mounts {
			envName := fmt.Sprintf("MOUNT_%s_PATH", strings.ToUpper(mount.Name))
			builder = builder.WithEnvVariable(envName, mount.Source)
		}

		builder = builder.WithExec([]string{"sh", "-c", "MAIN_PATH=$(find . -name main.go | head -n 1); if [ -z \"$MAIN_PATH\" ]; then echo 'ERROR: No main.go found in '\"$(pwd)\"; exit 1; fi; go build -o bin/service.exe $MAIN_PATH"})
	case "flutter":
		builder = m.flutterBase(client).
			WithDirectory("/src", src).
			WithWorkdir("/src/"+workspace).
			WithExec([]string{"flutter", "build", "web", "--wasm"})
	default:
		return fmt.Errorf("unknown stack for workspace: %s", workspace)
	}

	// Export binary or build artifacts back to host
	exportDir := "bin"
	if stack == "flutter" {
		exportDir = "build/web"
	}
	_, err = builder.Directory(exportDir).Export(ctx, filepath.Join(workspace, exportDir))
	return err
}

// buildLinux builds binaries for linux/amd64 inside the container
func (m *AihubForge) buildLinux(ctx context.Context, client *dagger.Client, src *dagger.Directory, workspace string) (*dagger.Directory, error) {
	stack := m.detectStack(workspace)
	fmt.Printf("⚒️ Forge: Linux Build [%s] (Stack: %s)\n", workspace, stack)

	var builder *dagger.Container
	switch stack {
	case "go":
		builder = m.goBase(client).
			WithEnvVariable("GOOS", "linux").
			WithEnvVariable("GOARCH", "amd64").
			WithDirectory("/src", src).
			WithNewFile("/src/go.work", m.minimalGoWork(workspace)).
			WithEnvVariable("GOWORK", "/src/go.work").
			WithWorkdir("/src/" + workspace).
			WithExec([]string{"sh", "-c", "MAIN_PATH=$(find . -name main.go | head -n 1); if [ -z \"$MAIN_PATH\" ]; then echo 'ERROR: No main.go found in '\"$(pwd)\"; exit 1; fi; go build -o bin/service $MAIN_PATH"})
		return builder.Directory("bin"), nil
	case "flutter":
		builder = m.flutterBase(client).
			WithDirectory("/src", src).
			WithWorkdir("/src/" + workspace).
			WithExec([]string{"flutter", "build", "web", "--wasm"})
		return builder.Directory("build/web"), nil
	default:
		return nil, fmt.Errorf("unknown stack for workspace: %s", workspace)
	}
}

// buildPodman builds OCI images for local Podman Desktop execution
func (m *AihubForge) buildPodman(ctx context.Context, client *dagger.Client, src *dagger.Directory, workspace string) error {
	fmt.Printf("⚒️ Forge: Podman Image Build [%s]\n", workspace)

	// Load Podman Mount System configuration
	config, err := m.loadProvisionConfig(workspace)
	if err != nil {
		return fmt.Errorf("failed to load provision config: %v", err)
	}

	binDir, err := m.buildLinux(ctx, client, src, workspace)
	if err != nil {
		return err
	}
	image := m.sealedImage(client, src, binDir, workspace)

	// Apply Mounts and Emulators to the runtime container
	image = m.applyMountsToContainer(client, image, config)

	// Tag for local Podman
	tag := fmt.Sprintf("localhost/%s:latest", strings.ToLower(workspace))
	_, err = image.Publish(ctx, tag)
	if err != nil {
		return err
	}

	fmt.Printf("\n✅ Podman Image Built: %s\n", tag)
	if len(config.Mounts) > 0 || len(config.Emulators) > 0 {
		fmt.Println("🚀 RUN COMMAND (with Podman Mount System):")
		fmt.Println(m.generatePodmanRun(workspace, config))
	}
	return nil
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

	binDir, err := m.buildLinux(ctx, client, src, workspace)
	if err != nil {
		return err
	}
	image := m.sealedImage(client, src, binDir, workspace)

	// Tag for GAR
	tag := fmt.Sprintf("%s-docker.pkg.dev/%s/olympus-fleet/%s:latest", garRegion, projectID, strings.ToLower(workspace))
	_, err = image.Publish(ctx, tag)
	return err
}

// sealedImage creates a deterministic, attestation-ready OCI image
func (m *AihubForge) sealedImage(client *dagger.Client, src *dagger.Directory, binDir *dagger.Directory, workspace string) *dagger.Container {
	return client.Container().From("debian:trixie-slim").
		WithDirectory("/app", src.Directory(workspace)).
		WithDirectory("/app/bin", binDir).
		WithWorkdir("/app").
		WithExec([]string{"apt-get", "update"}).
		WithExec([]string{"apt-get", "install", "-y", "ca-certificates"}).
		WithEntrypoint([]string{"./bin/service"})
}
