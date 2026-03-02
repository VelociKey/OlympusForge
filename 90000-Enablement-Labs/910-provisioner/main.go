package main

import (
	"archive/zip"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"io/fs"
	"log/slog"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"dagger.io/dagger"
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
	targetTool := flag.String("tool", "", "Provision a single tool by name")
	provisionAll := flag.Bool("all", false, "Force provision all tools regardless of state")
	registryPath := flag.String("registry", "", "Path to tool_registry.jebnf")
	flag.Parse()

	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	ctx := context.Background()

	basePath, _ := os.Getwd()
	// Root detection
	if _, err := os.Stat("OlympusForge"); err == nil {
		basePath = filepath.Join(basePath, "OlympusForge", "90000-Enablement-Labs", "000-Tools")
	}

	if *registryPath == "" {
		*registryPath = filepath.Join(basePath, "..", "..", "C0100-Configuration-Registry", "tool_registry.jebnf")
	}

	statePath := filepath.Join(basePath, "provisioner_state.json")
	state := loadState(statePath)

	logger.Info("Starting Forge Provisioner 3.0 (Targeted & Registry-Driven)", "base", basePath)

	// 1. Parse Registry
	allTools, err := ParseRegistry(*registryPath)
	if err != nil {
		logger.Error("Failed to parse registry", "path", *registryPath, "error", err)
		os.Exit(1)
	}

	// 2. Filter Tools
	var toolsToProcess []ToolDefinition
	if *targetTool != "" {
		for _, t := range allTools {
			if t.Name == *targetTool {
				toolsToProcess = append(toolsToProcess, t)
				break
			}
		}
		if len(toolsToProcess) == 0 {
			logger.Error("Target tool not found in registry", "tool", *targetTool)
			os.Exit(1)
		}
	} else {
		toolsToProcess = allTools
	}

	registry := &Registry{Symbols: make(map[string]string)}

	for _, tool := range toolsToProcess {
		lastState, exists := state.Tools[tool.Name]
		shouldProvision := *provisionAll || !exists || lastState.Version != tool.Version || tool.Version == "latest" || tool.Origin == "local-source"

		if shouldProvision {
			err := provisionTool(ctx, tool, basePath, logger)
			if err != nil {
				logger.Error("Failed to provision tool", "tool", tool.Name, "error", err)
				continue
			}
			state.Tools[tool.Name] = ToolState{
				Version:   tool.Version,
				Timestamp: time.Now(),
			}
		} else {
			logger.Info("Tool is up to date, skipping", "tool", tool.Name, "version", tool.Version)
		}
	}

	// Re-add symbols for all tools in registry even if we didn't provision them this run
	for _, t := range allTools {
		binPath := filepath.Join(basePath, "000-bin", t.Name+".cmd")
		if runtime.GOOS != "windows" {
			binPath = filepath.Join(basePath, "000-bin", t.Name)
		}
		registry.Symbols[t.Name] = binPath
	}

	// Save state and Registry
	saveState(statePath, state)
	saveRegistry(basePath, registry, logger)

	// Post-Provisioning: Rebuild mPSH Tool Index for O(1) lookup
	logger.Info("Triggering mPSH Tool Index rebuild (fleet-findTool)")
	rebuildCmd := exec.Command("fleet-findTool.exe", "-root", filepath.Join(basePath, "..", "..", ".."), "-rebuild")
	if out, err := rebuildCmd.CombinedOutput(); err != nil {
		logger.Warn("Failed to rebuild mPSH index", "error", err, "output", string(out))
	} else {
		logger.Info("mPSH Tool Index successfully synchronized")
	}
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
		return provisionLocalSource(ctx, tool, toolDir, binDir, basePath, logger)
	case "npm":
		return provisionNPM(tool, toolDir, binDir, logger)
	default:
		return fmt.Errorf("unsupported origin: %s", tool.Origin)
	}
}

func provisionExternal(tool ToolDefinition, toolDir string, binDir string, logger *slog.Logger) error {
	logger.Info("Downloading external tool", "tool", tool.Name, "url", tool.Package)
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
	logger.Info("Attempting GH-Release download via gh CLI", "tool", tool.Name, "repo", tool.Package)
	os.MkdirAll(toolDir, 0755)
	
	cmd := exec.Command("gh", "release", "download", tool.Version, "-R", tool.Package, "-p", "*.zip", "--dir", toolDir, "--clobber")
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("gh release download failed: %s: %w", string(out), err)
	}

	return createShim(tool, toolDir, binDir)
}

func provisionLocalSource(ctx context.Context, tool ToolDefinition, toolDir string, binDir string, basePath string, logger *slog.Logger) error {
	logger.Info("Daggerizing local source build (High Fidelity)", "tool", tool.Name, "src", tool.Package)
	
	client, err := dagger.Connect(ctx, dagger.WithLogOutput(os.Stdout))
	if err != nil { return err }
	defer client.Close()

	rootPath, _ := filepath.Abs(filepath.Join(basePath, "..", "..", ".."))
	isFlutter := strings.Contains(tool.Package, "interaction-surface") || strings.Contains(tool.Package, "BrandingFactory")
	
	// Quality Approach: Include ALL mod/sum files plus the tool's source
	// This makes go.work happy without uploading 10GB of node_modules/git history
	includes := []string{
		"**/go.mod",
		"**/go.sum",
		"**/pubspec.yaml",
		"go.work",
		"go.work.sum",
		tool.Package + "/**",
	}

	// Special case for tools that have cross-workspace dependencies beyond just mod files
	if tool.Name == "fleet-index" {
		includes = append(includes, "olympus.fleet/00SDLC/Olympus2/90000-Enablement-Labs/P0000-pkg/000-search/**")
		includes = append(includes, "olympus.fleet/00SDLC/Olympus2/00000-Identity-Foundations/**")
	}

	src := client.Host().Directory(rootPath, dagger.HostDirectoryOpts{
		Include: includes,
	})

	var builder *dagger.Container
	if isFlutter {
		builder = client.Container().
			From("ghcr.io/cirruslabs/flutter:stable").
			WithDirectory("/src", src).
			WithWorkdir("/src/"+tool.Package).
			WithExec([]string{"flutter", "build", "web", "--wasm"})
	} else {
		builder = client.Container().
			From("golang:1.26.0").
			WithEnvVariable("CGO_ENABLED", "0").
			WithEnvVariable("GOOS", runtime.GOOS).
			WithEnvVariable("GOARCH", runtime.GOARCH).
			WithDirectory("/src", src).
			WithWorkdir("/src/"+tool.Package).
			WithExec([]string{"go", "build", "-o", tool.Binary, "."})
	}

	_, err = builder.File(tool.Binary).Export(ctx, filepath.Join(toolDir, tool.Binary))
	if err != nil { return err }

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
	if _, err := os.Stat(binPath); err != nil {
		// Try recursive search for the binary
		found := ""
		filepath.WalkDir(toolDir, func(p string, d fs.DirEntry, err error) error {
			if !d.IsDir() && d.Name() == tool.Binary {
				found = p
				return filepath.SkipAll
			}
			return nil
		})
		if found != "" {
			binPath = found
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
	client := &http.Client{Timeout: 10 * time.Minute}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil { return err }
	req.Header.Set("User-Agent", "OlympusForge-Provisioner/3.0")

	resp, err := client.Do(req)
	if err != nil { return err }
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to download: status %d", resp.StatusCode)
	}

	out, err := os.Create(dest)
	if err != nil { return err }
	defer out.Close()
	_, err = io.Copy(out, resp.Body)
	return err
}

func unzip(src, dest string) error {
	r, err := zip.OpenReader(src)
	if err != nil { return err }
	defer r.Close()

	for _, f := range r.File {
		fpath := filepath.Join(dest, f.Name)
		if f.FileInfo().IsDir() {
			os.MkdirAll(fpath, os.ModePerm)
			continue
		}
		if err := os.MkdirAll(filepath.Dir(fpath), os.ModePerm); err != nil { return err }
		outFile, err := os.OpenFile(fpath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
		if err != nil { return err }
		rc, err := f.Open()
		if err != nil {
			outFile.Close()
			return err
		}
		_, err = io.Copy(outFile, rc)
		outFile.Close()
		rc.Close()
		if err != nil { return err }
	}
	return nil
}

func saveRegistry(basePath string, reg *Registry, logger *slog.Logger) {
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
