package main

import (
	"archive/zip"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

// ToolDefinition matches the jeBNF structure
type ToolDefinition struct {
	Name     string
	Category string
	Version  string
	Origin   string // go-install, gh-release, external, local-source, npm
	Package  string
	Binary   string
}

// Registry stores the symbol table for quick lookup
type Registry struct {
	Symbols map[string]string `json:"symbols"`
}

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	ctx := context.Background()

	basePath, _ := os.Getwd()
	// Root detection
	if _, err := os.Stat("OlympusForge"); err == nil {
		basePath = filepath.Join(basePath, "OlympusForge", "90000-Enablement-Labs", "000-Tools")
	}

	logger.Info("Starting Forge Provisioner 2.0", "base", basePath)

	// In a real implementation, we would parse the .jebnf here.
	// For this task, I will use a hardcoded slice derived from the .jebnf
	// but architected to be easily swappable for a parser.
	// Expanded toolset based on Olympus requirements
	tools := []ToolDefinition{
		// Foundation
		{Name: "go", Category: "foundation", Version: "1.24.0", Origin: "external", Package: "https://go.dev/dl/go1.24.0.windows-amd64.zip", Binary: "bin/go.exe"},
		{Name: "git", Category: "foundation", Version: "2.47.1", Origin: "external", Package: "https://github.com/git-for-windows/git/releases/download/v2.47.1.windows.1/MinGit-2.47.1-64-bit.zip", Binary: "cmd/git.exe"},
		{Name: "gh", Category: "foundation", Version: "2.66.1", Origin: "external", Package: "https://github.com/cli/cli/releases/download/v2.66.1/gh_2.66.1_windows_amd64.zip", Binary: "bin/gh.exe"},
		//{Name: "flutter", Category: "foundation", Version: "3.24.1", Origin: "external", Package: "https://storage.googleapis.com/flutter_infra_release/releases/stable/windows/flutter_windows_3.24.1-stable.zip", Binary: "bin/flutter.bat"},

		// Authoring
		{Name: "buf", Category: "authoring", Version: "v1.50.0", Origin: "go-install", Package: "github.com/bufbuild/buf/cmd/buf@v1.50.0", Binary: "buf.exe"},
		{Name: "air", Category: "authoring", Version: "v1.64.5", Origin: "go-install", Package: "github.com/air-verse/air@v1.64.5", Binary: "air.exe"},
		{Name: "golangci-lint", Category: "authoring", Version: "v1.64.4", Origin: "go-install", Package: "github.com/golangci/golangci-lint/cmd/golangci-lint@v1.64.4", Binary: "golangci-lint.exe"},

		// Security
		{Name: "trivy", Category: "security", Version: "v0.59.1", Origin: "gh-release", Package: "aquasecurity/trivy", Binary: "trivy.exe"},

		// Infrastructure
		{Name: "gcloud", Category: "infrastructure", Version: "latest", Origin: "external", Package: "https://dl.google.com/dl/cloudsdk/channels/rapid/google-cloud-sdk-windows-x86_64-bundled-python.zip", Binary: "bin/gcloud.cmd"},
		{Name: "gradle", Category: "infrastructure", Version: "8.13", Origin: "external", Package: "https://services.gradle.org/distributions/gradle-8.13-bin.zip", Binary: "bin/gradle.bat"},
		{Name: "firebase", Category: "infrastructure", Version: "latest", Origin: "npm", Package: "firebase-tools", Binary: "firebase.cmd"},

		// Intelligence (Gemini Tools)
		{Name: "george-bootstrap", Category: "intelligence", Version: "v1.0.0", Origin: "local-source", Package: "90000-Enablement-Labs/000-Tools/000-intelligence/george-bootstrap", Binary: "george-bootstrap.exe"},
		{Name: "fleet-doctor", Category: "intelligence", Version: "v1.0.0", Origin: "local-source", Package: "90000-Enablement-Labs/000-Tools/000-maintenance/fleet-doctor", Binary: "fleet-doctor.exe"},
	}

	registry := &Registry{Symbols: make(map[string]string)}

	for _, tool := range tools {
		err := provisionTool(ctx, tool, basePath, logger)
		if err != nil {
			logger.Error("Failed to provision tool", "tool", tool.Name, "error", err)
			continue
		}

		// Map for Symbol Table (Quick Lookup)
		binPath := filepath.Join(basePath, "000-bin", tool.Name+".cmd")
		if runtime.GOOS != "windows" {
			binPath = filepath.Join(basePath, "000-bin", tool.Name)
		}
		registry.Symbols[tool.Name] = binPath
	}

	// Save Symbol Table (Registry)
	saveRegistry(basePath, registry, logger)
}

func provisionTool(ctx context.Context, tool ToolDefinition, basePath string, logger *slog.Logger) error {
	categoryDir := filepath.Join(basePath, "000-"+tool.Category)
	os.MkdirAll(categoryDir, 0755)

	binDir := filepath.Join(basePath, "000-bin")
	os.MkdirAll(binDir, 0755)

	toolDir := filepath.Join(categoryDir, tool.Name)

	switch tool.Origin {
	case "external":
		return provisionExternal(tool, toolDir, binDir, logger)
	case "go-install":
		return provisionGoInstall(tool, toolDir, binDir, logger)
	case "gh-release":
		return provisionGHRelease(tool, toolDir, binDir, logger)
	case "local-source":
		return provisionLocalSource(tool, toolDir, binDir, basePath, logger)
	case "npm":
		return provisionNPM(tool, toolDir, binDir, logger)
	default:
		return fmt.Errorf("unsupported origin: %s", tool.Origin)
	}
}

func provisionExternal(tool ToolDefinition, toolDir string, binDir string, logger *slog.Logger) error {
	if _, err := os.Stat(toolDir); err == nil {
		logger.Info("Tool already exists, skipping download", "tool", tool.Name)
		return createShim(tool, toolDir, binDir)
	}

	logger.Info("Downloading external tool", "tool", tool.Name, "url", tool.Package)
	tmpZip := toolDir + ".zip"
	if err := downloadFile(tool.Package, tmpZip); err != nil {
		return err
	}
	defer os.Remove(tmpZip)

	if err := unzip(tmpZip, toolDir); err != nil {
		return err
	}

	return createShim(tool, toolDir, binDir)
}

func provisionGoInstall(tool ToolDefinition, toolDir string, binDir string, logger *slog.Logger) error {
	logger.Info("Go installing tool", "tool", tool.Name, "package", tool.Package)
	os.MkdirAll(toolDir, 0755)

	cmd := exec.Command("go", "install", tool.Package)
	cmd.Env = append(os.Environ(), "GOBIN="+toolDir)
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("go install failed: %s: %w", string(out), err)
	}

	return createShim(tool, toolDir, binDir)
}

func provisionGHRelease(tool ToolDefinition, toolDir string, binDir string, logger *slog.Logger) error {
	if strings.HasPrefix(tool.Package, "http") {
		return provisionExternal(tool, toolDir, binDir, logger)
	}
	// Simple mock for GH Release asset discovery - for now we just log
	logger.Warn("Complex GH-Release resolution not yet implemented, use direct URL in Package", "tool", tool.Name)
	return nil
}

func provisionLocalSource(tool ToolDefinition, toolDir string, binDir string, basePath string, logger *slog.Logger) error {
	logger.Info("Building local source", "tool", tool.Name, "src", tool.Package)
	// basePath is .../000-Tools. Package is relative to .../90000-Enablement-Labs.
	// Forge root is basePath/../.. i.e. OlympusForge root.
	forgeRoot, _ := filepath.Abs(filepath.Join(basePath, "..", ".."))
	absSrc := filepath.Join(forgeRoot, tool.Package)
	os.MkdirAll(toolDir, 0755)

	targetBin := filepath.Join(toolDir, tool.Binary)
	cmd := exec.Command("go", "build", "-o", targetBin, ".")
	cmd.Dir = absSrc
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("local build failed: %s: %w", string(out), err)
	}

	return createShim(tool, toolDir, binDir)
}

func provisionNPM(tool ToolDefinition, toolDir string, binDir string, logger *slog.Logger) error {
	logger.Info("NPM installing tool", "tool", tool.Name, "package", tool.Package)
	os.MkdirAll(toolDir, 0755)

	cmd := exec.Command("npm", "install", "-g", tool.Package, "--prefix", toolDir)
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("npm install failed: %s: %w", string(out), err)
	}

	return createShim(tool, toolDir, binDir)
}

func createShim(tool ToolDefinition, toolDir string, binDir string) error {
	binPath := filepath.Join(toolDir, tool.Binary)
	relPath, _ := filepath.Rel(binDir, binPath)
	shimPath := filepath.Join(binDir, tool.Name+".cmd")
	return writeShim(shimPath, relPath)
}

func writeShim(path string, relTarget string) error {
	content := fmt.Sprintf("@echo off\n\"%%~dp0%s\" %%*\n", relTarget)
	return os.WriteFile(path, []byte(content), 0644)
}

func downloadFile(url string, dest string) error {
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	out, err := os.Create(dest)
	if err != nil {
		return err
	}
	defer out.Close()
	_, err = io.Copy(out, resp.Body)
	return err
}

func unzip(src, dest string) error {
	r, err := zip.OpenReader(src)
	if err != nil {
		return err
	}
	defer r.Close()

	os.MkdirAll(dest, 0755)

	for _, f := range r.File {
		fpath := filepath.Join(dest, f.Name)
		if f.FileInfo().IsDir() {
			os.MkdirAll(fpath, os.ModePerm)
			continue
		}

		if err := os.MkdirAll(filepath.Dir(fpath), os.ModePerm); err != nil {
			return err
		}

		outFile, err := os.OpenFile(fpath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
		if err != nil {
			return err
		}

		rc, err := f.Open()
		if err != nil {
			return err
		}

		_, err = io.Copy(outFile, rc)
		outFile.Close()
		rc.Close()
		if err != nil {
			return err
		}
	}
	return nil
}

func saveRegistry(basePath string, reg *Registry, logger *slog.Logger) {
	// JSON version for cross-language agents
	regPath := filepath.Join(basePath, "tool_registry.json")
	data, _ := json.MarshalIndent(reg, "", "  ")
	os.WriteFile(regPath, data, 0644)

	// Go version for compiled-in MPH performance (Mantra aligned)
	registryDir := filepath.Join(basePath, "registry")
	os.MkdirAll(registryDir, 0755)
	goPath := filepath.Join(registryDir, "registry_gen.go")

	var sb strings.Builder
	sb.WriteString("package registry\n\n// Generated Tool Registry - DO NOT EDIT\n\n")
	sb.WriteString("var ToolRegistry = map[string]string{\n")
	for name, path := range reg.Symbols {
		sb.WriteString(fmt.Sprintf("\t%q: %q,\n", name, path))
	}
	sb.WriteString("}\n")
	os.WriteFile(goPath, []byte(sb.String()), 0644)

	logger.Info("Symbol Table Registry generated", "json", regPath, "go", goPath, "count", len(reg.Symbols))
}
