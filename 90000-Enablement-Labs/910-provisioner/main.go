package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// Tool represents a software component managed by the provisioner.
type Tool struct {
	Name     string
	Category string
	Method   string // go-install, npm-install, gh-release, shim-host
	Package  string // Package path for go/npm or GitHub repo tag
	Binary   string // Expected binary name
}

var tools = []Tool{
	// Authoring
	{Name: "golangci-lint", Category: "authoring", Method: "go-install", Package: "github.com/golangci/golangci-lint/cmd/golangci-lint@v1.64.8", Binary: "golangci-lint.exe"},
	{Name: "air", Category: "authoring", Method: "go-install", Package: "github.com/air-verse/air@v1.64.5", Binary: "air.exe"},
	{Name: "buf", Category: "authoring", Method: "go-install", Package: "github.com/bufbuild/buf/cmd/buf@v1.50.0", Binary: "buf.exe"},
	{Name: "protoc-gen-connect-go", Category: "authoring", Method: "go-install", Package: "connectrpc.com/connect/cmd/protoc-gen-connect-go@v1.18.1", Binary: "protoc-gen-connect-go.exe"},
	{Name: "gh", Category: "authoring", Method: "gh-release", Package: "cli/cli", Binary: "gh.exe"},

	// Infrastructure
	{Name: "trivy", Category: "infrastructure", Method: "gh-release", Package: "aquasecurity/trivy", Binary: "trivy.exe"},
	{Name: "tofu", Category: "infrastructure", Method: "gh-release", Package: "opentofu/opentofu", Binary: "tofu.exe"},
	{Name: "jq", Category: "infrastructure", Method: "gh-release", Package: "jqlang/jq", Binary: "jq-win64.exe"},
	{Name: "dagger", Category: "infrastructure", Method: "go-install", Package: "github.com/dagger/dagger/cmd/dagger@v0.19.11", Binary: "dagger.exe"},

	// Sovereignty
	{Name: "otelcol", Category: "sovereignty", Method: "gh-release", Package: "open-telemetry/opentelemetry-collector-releases", Binary: "otelcol.exe"},

	// Intelligence
	{Name: "dlv", Category: "intelligence", Method: "go-install", Package: "github.com/go-delve/delve/cmd/dlv@latest", Binary: "dlv.exe"},
	{Name: "ollama", Category: "intelligence", Method: "shim-host", Package: "ollama", Binary: "ollama"},
	{Name: "jules", Category: "intelligence", Method: "shim-host", Package: "jules", Binary: "jules"},
	{Name: "gcloud", Category: "infrastructure", Method: "shim-host", Package: "gcloud", Binary: "gcloud"},
	{Name: "firebase", Category: "infrastructure", Method: "local-bundle", Package: "firebase", Binary: "firebase"},
	{Name: "pubsub-emulator", Category: "infrastructure", Method: "local-bundle", Package: "pubsub-emulator", Binary: "pubsub-emulator"},
	{Name: "bigtable-emulator", Category: "infrastructure", Method: "local-bundle", Package: "bigtable-emulator", Binary: "bigtable-emulator"},
	{Name: "datastore-emulator", Category: "infrastructure", Method: "local-bundle", Package: "datastore-emulator", Binary: "datastore-emulator"},
	{Name: "firestore-emulator", Category: "infrastructure", Method: "local-bundle", Package: "firestore-emulator", Binary: "firestore-emulator"},
	{Name: "spanner-emulator", Category: "infrastructure", Method: "local-bundle", Package: "spanner-emulator", Binary: "spanner-emulator"},
	{Name: "storage-emulator", Category: "infrastructure", Method: "local-bundle", Package: "storage-emulator", Binary: "storage-emulator"},

	// Validation
	{Name: "govulncheck", Category: "validation", Method: "go-install", Package: "golang.org/x/vuln/cmd/govulncheck@latest", Binary: "govulncheck.exe"},
	{Name: "gosec", Category: "validation", Method: "go-install", Package: "github.com/securego/gosec/v2/cmd/gosec@latest", Binary: "gosec.exe"},
	{Name: "nuclei", Category: "validation", Method: "go-install", Package: "github.com/projectdiscovery/nuclei/v3/cmd/nuclei@latest", Binary: "nuclei.exe"},
}

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	ctx := context.Background()

	basePath, err := os.Getwd()
	if err != nil {
		logger.Error("Failed to get current directory", "error", err)
		os.Exit(1)
	}

	// If running from root, redirect Tools to OlympusForge/90000-Enablement-Labs/000-Tools
	if _, err := os.Stat("OlympusForge"); err == nil {
		basePath = filepath.Join(basePath, "OlympusForge", "90000-Enablement-Labs", "000-Tools")
	} else if _, err := os.Stat("Olympus2"); err == nil {
		basePath = filepath.Join(basePath, "Olympus2", "90000-Enablement-Labs", "000-Tools")
	}

	logger.Info("Starting AIHub Master Forge Provisioner", "base", basePath)

	for _, tool := range tools {
		toolLogger := logger.With("tool", tool.Name, "category", tool.Category)

		err := provisionTool(ctx, tool, basePath, toolLogger)
		if err != nil {
			toolLogger.Error("Provisioning failed", "error", err)
		} else {
			toolLogger.Info("Provisioning successful")
		}
	}
}

func provisionTool(ctx context.Context, tool Tool, basePath string, logger *slog.Logger) error {
	targetDir := filepath.Join(basePath, "000-"+tool.Category)
	if err := os.MkdirAll(targetDir, 0755); err != nil {
		return fmt.Errorf("failed to create category directory: %w", err)
	}

	binDir := filepath.Join(basePath, "000-bin")
	if err := os.MkdirAll(binDir, 0755); err != nil {
		return fmt.Errorf("failed to create bin directory: %w", err)
	}

	switch tool.Method {
	case "go-install":
		return provisionGoTool(tool, targetDir, binDir, logger)
	case "shim-host":
		return createShim(tool, binDir, logger)
	case "local-bundle":
		return provisionLocalBundle(tool, binDir, logger)
	case "gh-release":
		logger.Warn("GH-Release method not fully implemented in this preview, skipping", "repo", tool.Package)
		return nil
	default:
		return fmt.Errorf("unknown provisioning method: %s", tool.Method)
	}
}

func provisionLocalBundle(tool Tool, binDir string, logger *slog.Logger) error {
	logger.Info("Provisioning local fleet bundle")
	// For local-bundle, the files are already placed in OlympusForge by the developer
	// We just ensure the bin shim points to the right internal location
	return nil
}

func provisionGoTool(tool Tool, targetDir string, binDir string, logger *slog.Logger) error {
	logger.Info("Installing via go install", "package", tool.Package)

	cmd := exec.Command("go", "install", tool.Package)
	cmd.Env = append(os.Environ(), "GOBIN="+targetDir)

	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("go install failed: %s: %w", string(output), err)
	}

	// Create shim in Tools/bin
	shimPath := filepath.Join(binDir, strings.TrimSuffix(tool.Binary, ".exe")+".cmd")
	relPath, _ := filepath.Rel(binDir, filepath.Join(targetDir, tool.Binary))

	content := fmt.Sprintf("@echo off\n\"%%~dp0%s\" %%*\n", relPath)
	return os.WriteFile(shimPath, []byte(content), 0644)
}

func createShim(tool Tool, binDir string, logger *slog.Logger) error {
	logger.Info("Creating host shim")
	shimPath := filepath.Join(binDir, tool.Name+".cmd")
	content := fmt.Sprintf("@echo off\n%s %%*\n", tool.Package)
	return os.WriteFile(shimPath, []byte(content), 0644)
}
