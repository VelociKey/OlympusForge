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
	"time"
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

// ToolState tracks the last provisioned version/state
type ToolState struct {
	Version   string    `json:"version"`
	Timestamp time.Time `json:"timestamp"`
}

type ProvisionerState struct {
	Tools map[string]ToolState `json:"tools"`
}

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	ctx := context.Background()

	basePath, _ := os.Getwd()
	// Root detection
	if _, err := os.Stat("OlympusForge"); err == nil {
		basePath = filepath.Join(basePath, "OlympusForge", "90000-Enablement-Labs", "000-Tools")
	}

	statePath := filepath.Join(basePath, "provisioner_state.json")
	state := loadState(statePath)

	logger.Info("Starting Forge Provisioner 2.0 (Incremental)", "base", basePath)

	tools := []ToolDefinition{
		// Foundation
		{Name: "gh", Category: "foundation", Version: "v2.67.0", Origin: "external", Package: "https://github.com/cli/cli/releases/download/v2.67.0/gh_2.67.0_windows_amd64.zip", Binary: "bin/gh.exe"},
		{Name: "git", Category: "foundation", Version: "v2.48.1", Origin: "external", Package: "https://github.com/git-for-windows/git/releases/download/v2.48.1.windows.1/MinGit-2.48.1-64-bit.zip", Binary: "cmd/git.exe"},

		// Authoring
		{Name: "buf", Category: "authoring", Version: "v1.50.0", Origin: "go-install", Package: "github.com/bufbuild/buf/cmd/buf@v1.50.0", Binary: "buf.exe"},
		{Name: "air", Category: "authoring", Version: "v1.64.5", Origin: "go-install", Package: "github.com/air-verse/air@v1.64.5", Binary: "air.exe"},
		{Name: "golangci-lint", Category: "authoring", Version: "v1.64.4", Origin: "go-install", Package: "github.com/golangci/golangci-lint/cmd/golangci-lint@v1.64.4", Binary: "golangci-lint.exe"},

		// Security
		{Name: "trivy", Category: "security", Version: "v0.59.1", Origin: "external", Package: "https://github.com/aquasecurity/trivy/releases/download/v0.59.1/trivy_0.59.1_windows-64bit.zip", Binary: "trivy.exe"},

		// Infrastructure
		{Name: "jdk", Category: "infrastructure", Version: "25", Origin: "external", Package: "https://github.com/adoptium/temurin25-binaries/releases/download/jdk-25.0.0%2B9/OpenJDK25U-jdk_x64_windows_hotspot_25.0.0_9.zip", Binary: "jdk-25.0.0+9/bin/java.exe"},
		{Name: "gcloud", Category: "infrastructure", Version: "latest", Origin: "external", Package: "https://dl.google.com/dl/cloudsdk/channels/rapid/google-cloud-sdk-windows-x86_64-bundled-python.zip", Binary: "google-cloud-sdk/bin/gcloud.cmd"},
		{Name: "gradle", Category: "infrastructure", Version: "8.13", Origin: "external", Package: "https://services.gradle.org/distributions/gradle-8.13-bin.zip", Binary: "gradle-8.13/bin/gradle.bat"},
		{Name: "firebase", Category: "infrastructure", Version: "latest", Origin: "npm", Package: "firebase-tools", Binary: "firebase.cmd"},

		// Intelligence (Gemini Tools)
		{Name: "gemini-cli", Category: "intelligence", Version: "latest", Origin: "npm", Package: "@google/gemini-cli", Binary: "gemini.cmd"},
		{Name: "george-bootstrap", Category: "intelligence", Version: "v1.0.0", Origin: "local-source", Package: "90000-Enablement-Labs/000-Tools/000-intelligence/george-bootstrap", Binary: "george-bootstrap.exe"},
		{Name: "fleet-doctor", Category: "intelligence", Version: "v1.0.0", Origin: "local-source", Package: "90000-Enablement-Labs/000-Tools/000-maintenance/fleet-doctor", Binary: "fleet-doctor.exe"},
		{Name: "fleet-ls", Category: "intelligence", Version: "v1.0.0", Origin: "local-source", Package: "../Olympus2/90000-Enablement-Labs/.000-Tools/fleet-ls", Binary: "fleet-ls.exe"},
		{Name: "fleet-dagger-up", Category: "intelligence", Version: "v1.0.0", Origin: "local-source", Package: "../Olympus2/90000-Enablement-Labs/.000-Tools/fleet-dagger-up", Binary: "fleet-dagger-up.exe"},
		{Name: "fleet-large-file-finder", Category: "intelligence", Version: "v1.0.0", Origin: "local-source", Package: "90000-Enablement-Labs/000-Tools/000-intelligence/fleet-large-file-finder", Binary: "fleet-large-file-finder.exe"},
		{Name: "fleet-daily-chronicle", Category: "intelligence", Version: "v1.0.0", Origin: "local-source", Package: "90000-Enablement-Labs/000-Tools/000-intelligence/fleet-daily-chronicle", Binary: "fleet-daily-chronicle.exe"},
	}

	registry := &Registry{Symbols: make(map[string]string)}

	for _, tool := range tools {
		lastState, exists := state.Tools[tool.Name]
		if exists && lastState.Version == tool.Version && tool.Version != "latest" {
			// Skip if version matches and is not "latest"
			logger.Debug("Tool version matches last run, skipping", "tool", tool.Name, "version", tool.Version)
		} else {
			err := provisionTool(ctx, tool, basePath, logger)
			if err != nil {
				logger.Error("Failed to provision tool", "tool", tool.Name, "error", err)
				continue
			}
			state.Tools[tool.Name] = ToolState{
				Version:   tool.Version,
				Timestamp: time.Now(),
			}
		}

		// Map for Symbol Table (Quick Lookup)
		binPath := filepath.Join(basePath, "000-bin", tool.Name+".cmd")
		if runtime.GOOS != "windows" {
			binPath = filepath.Join(basePath, "000-bin", tool.Name)
		}
		registry.Symbols[tool.Name] = binPath
	}

	// Save state and Registry
	saveState(statePath, state)
	saveRegistry(basePath, registry, logger)
}

func loadState(path string) ProvisionerState {
	state := ProvisionerState{Tools: make(map[string]ToolState)}
	data, err := os.ReadFile(path)
	if err == nil {
		json.Unmarshal(data, &state)
	}
	return state
}

func saveState(path string, state ProvisionerState) {
	data, _ := json.MarshalIndent(state, "", "  ")
	os.WriteFile(path, data, 0644)
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
	logger.Info("Downloading external tool", "tool", tool.Name, "url", tool.Package)
	
	// Clean up old dir to ensure fresh unzip
	os.RemoveAll(toolDir)
	os.MkdirAll(toolDir, 0755)

	tmpZip := toolDir + ".zip"
	if err := downloadFile(tool.Package, tmpZip); err != nil {
		return err
	}
	defer os.Remove(tmpZip)

	if err := unzip(tmpZip, toolDir); err != nil {
		return fmt.Errorf("failed to unzip %s: %w", tmpZip, err)
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
	// Attempt to use 'gh' if available
	logger.Info("Attempting GH-Release download via gh CLI", "tool", tool.Name, "repo", tool.Package)
	os.MkdirAll(toolDir, 0755)
	
	// gh release download <tag> -R <repo> -p <pattern>
	cmd := exec.Command("gh", "release", "download", tool.Version, "-R", tool.Package, "-p", "*.zip", "--dir", toolDir)
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("gh release download failed: %s: %w", string(out), err)
	}

	// This is a simplification; we'd need to find the zip inside toolDir and unzip it.
	return createShim(tool, toolDir, binDir)
}

func provisionLocalSource(tool ToolDefinition, toolDir string, binDir string, basePath string, logger *slog.Logger) error {
	logger.Info("Building local source", "tool", tool.Name, "src", tool.Package)
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
	// Check if binPath exists before creating shim
	if _, err := os.Stat(binPath); err != nil {
		// Look for it one level deeper if unzip was messy
		files, _ := filepath.Glob(filepath.Join(toolDir, "*", tool.Binary))
		if len(files) > 0 {
			binPath = files[0]
		} else {
			return fmt.Errorf("could not find binary %q in %q", tool.Binary, toolDir)
		}
	}

	relPath, _ := filepath.Rel(binDir, binPath)
	shimPath := filepath.Join(binDir, tool.Name+".cmd")
	return writeShim(shimPath, relPath)
}

func writeShim(path string, relTarget string) error {
	content := fmt.Sprintf("@echo off\n\"%%~dp0%s\" %%*\n", relTarget)
	return os.WriteFile(path, []byte(content), 0644)
}

func downloadFile(url string, dest string) error {
	// Set a User-Agent to avoid being blocked by some servers (like GitHub or Adoptium)
	client := &http.Client{Timeout: 5 * time.Minute}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil { return err }
	req.Header.Set("User-Agent", "OlympusForge-Provisioner/2.0")

	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to download: status %d", resp.StatusCode)
	}

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
			outFile.Close()
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
	// jeBNF version for agent navigation (Mantra aligned)
	regPath := filepath.Join(basePath, "tool_symbols.jebnf")
	var jb strings.Builder
	jb.WriteString("# Forge Tool Symbols (jeBNF-mPSH)\n")
	jb.WriteString(fmt.Sprintf("generated_at = %q\n\n", time.Now().Format(time.RFC3339)))
	jb.WriteString("Symbols {\n")
	for name, path := range reg.Symbols {
		jb.WriteString(fmt.Sprintf("  %s = %q\n", name, filepath.ToSlash(path)))
	}
	jb.WriteString("}\n")
	os.WriteFile(regPath, []byte(jb.String()), 0644)

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

	logger.Info("Symbol Table Registry generated", "jebnf", regPath, "go", goPath, "count", len(reg.Symbols))
}
