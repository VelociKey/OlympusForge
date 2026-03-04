package main

import (
	"archive/zip"
	"bufio"
	"bytes"
	"context"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"dagger.io/dagger"
)

type Tool struct {
	Name       string
	Owner      string
	Repo       string
	Version    string
	BinaryName string
	LocalPath  string
}

func main() {
	buildCmd := flag.NewFlagSet("build", flag.ExitOnError)
	ingestCmd := flag.NewFlagSet("ingest", flag.ExitOnError)

	if len(os.Args) < 2 {
		fmt.Println("Usage: fleet-forge-pipeline <command> [options]")
		fmt.Println("Commands: build, ingest")
		os.Exit(1)
	}

	switch os.Args[1] {
	case "build":
		handleBuild(buildCmd, os.Args[2:])
	case "ingest":
		handleIngest(ingestCmd, os.Args[2:])
	default:
		fmt.Printf("Unknown command: %s\n", os.Args[1])
		os.Exit(1)
	}
}

func handleBuild(fs *flag.FlagSet, args []string) {
	fmt.Println("Sovereign Build Pipeline initialized...")

	ctx := context.Background()

	// Initialize Dagger client
	client, err := dagger.Connect(ctx, dagger.WithLogOutput(os.Stderr))
	if err != nil {
		log.Fatalf("❌ Failed to connect to Dagger engine: %v", err)
	}
	defer client.Close()

	rootDir := os.Getenv("ANTIGRAVITY_ROOT")
	if rootDir == "" {
		wd, _ := os.Getwd()
		rootDir = filepath.ToSlash(wd)
	}

	src := client.Host().Directory(rootDir)

	// Set up the builder Go container
	builder := client.Container().From("golang:1.24-alpine").
		WithDirectory("/src", src).
		WithWorkdir("/src").
		WithEnvVariable("CGO_ENABLED", "0").
		WithEnvVariable("GOCACHE", "/src/00SDLC/Olympus2/C0990-Ephemeral-Scratch/go-cache")

	toolsDir := "00SDLC/OlympusForge/82000-Toolchain-Fleet"
	outDir := filepath.ToSlash(filepath.Join(toolsDir, "bin"))

	entries, err := os.ReadDir(filepath.Join(rootDir, filepath.FromSlash(toolsDir)))
	if err != nil {
		log.Fatalf("Failed to read tools directory: %v", err)
	}

	for _, e := range entries {
		if e.IsDir() && (strings.HasPrefix(e.Name(), "fleet-") || strings.HasPrefix(e.Name(), "gemaid")) && e.Name() != "bin" && e.Name() != "templates" {
			toolName := e.Name()
			mainPath := filepath.ToSlash(filepath.Join("./", toolsDir, toolName))
			exePath := filepath.ToSlash(filepath.Join(outDir, toolName+".exe"))

			fmt.Printf("🔨 Building %s via Dagger...\n", toolName)

			builder = builder.WithExec([]string{
				"go", "build", "-ldflags", "-s -w", "-trimpath", "-o", exePath, mainPath,
			})
		}
	}

	_, err = builder.Directory(outDir).Export(ctx, filepath.Join(rootDir, filepath.FromSlash(outDir)))
	if err != nil {
		log.Fatalf("❌ Failed to export built binaries from Dagger container: %v", err)
	}

	fmt.Println("✅ All tools built successfully and securely placed in the bin directory via Dagger.")
}

func handleIngest(fs *flag.FlagSet, args []string) {
	target := fs.String("target", "all", "Specific tool to ingest (or 'all')")
	fs.Parse(args)

	fmt.Println("Sovereign Ingest Pipeline initialized...")
	registryPath := "olympus.fleet/00SDLC/OlympusForge/C0100-Configuration-Registry/DEPENDENCIES.jebnf"

	tools, err := loadTools(registryPath)
	if err != nil {
		log.Fatalf("❌ Failed to load registry: %v", err)
	}

	for _, tool := range tools {
		if *target != "all" && tool.Name != *target {
			continue
		}
		if err := restoreTool(tool); err != nil {
			log.Printf("❌ Failed to restore %s: %v\n", tool.Name, err)
		} else {
			log.Printf("✅ Restored %s to %s\n", tool.Name, tool.LocalPath)
		}
	}
}

func loadTools(path string) ([]Tool, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var tools []Tool
	var current Tool
	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if strings.HasPrefix(line, "Tool") {
			name := strings.Trim(strings.TrimPrefix(line, "Tool"), " {\"")
			current = Tool{Name: name}
		}
		if line == "}" && current.Name != "" {
			tools = append(tools, current)
			current = Tool{}
		}

		if current.Name != "" {
			parts := strings.SplitN(line, "=", 2)
			if len(parts) == 2 {
				key := strings.TrimSpace(parts[0])
				val := strings.Trim(strings.TrimSpace(parts[1]), "\";")
				switch key {
				case "Owner":
					current.Owner = val
				case "Repo":
					current.Repo = val
				case "Version":
					current.Version = val
				case "BinaryName":
					current.BinaryName = val
				case "LocalPath":
					current.LocalPath = val
				}
			}
		}
	}
	return tools, nil
}

func restoreTool(t Tool) error {
	if t.Name == "flutter" || t.Name == "firebase" {
		fmt.Printf("⚠️ Skipping %s (Manual/System prerequisite)\n", t.Name)
		return nil
	}

	var url string
	isZip := false

	switch t.Name {
	case "buf":
		url = fmt.Sprintf("https://github.com/bufbuild/buf/releases/download/%s/buf-Windows-x86_64.exe", t.Version)
	case "trivy":
		// https://github.com/aquasecurity/trivy/releases/download/v0.69.3/trivy_0.69.3_windows-64bit.zip
		verNoV := strings.TrimPrefix(t.Version, "v")
		url = fmt.Sprintf("https://github.com/aquasecurity/trivy/releases/download/%s/trivy_%s_windows-64bit.zip", t.Version, verNoV)
		isZip = true
	case "cosign":
		// https://github.com/sigstore/cosign/releases/download/v2.4.1/cosign-windows-amd64.exe
		url = fmt.Sprintf("https://github.com/sigstore/cosign/releases/download/%s/cosign-windows-amd64.exe", t.Version)
	default:
		url = fmt.Sprintf("https://github.com/%s/%s/releases/download/%s/%s-Windows-x86_64.exe", t.Owner, t.Repo, t.Version, t.Repo)
	}

	fmt.Printf("⬇️ Downloading %s from %s...\n", t.Name, url)

	targetPath := filepath.Clean(t.LocalPath)
	if strings.HasPrefix(targetPath, ".") {
		targetPath = filepath.Join("00SDLC", "OlympusForge", targetPath)
	}

	if err := os.MkdirAll(filepath.Dir(targetPath), 0755); err != nil {
		return fmt.Errorf("mkdir failed: %w", err)
	}

	resp, err := http.Get(url)
	if err != nil {
		return fmt.Errorf("network error: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return fmt.Errorf("HTTP %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("read body failed: %w", err)
	}

	if isZip {
		return extractFromZip(body, t.BinaryName, targetPath)
	}

	return os.WriteFile(targetPath, body, 0755)
}

func extractFromZip(data []byte, binaryName string, targetPath string) error {
	reader, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		return err
	}

	for _, file := range reader.File {
		if filepath.Base(file.Name) == binaryName {
			rc, err := file.Open()
			if err != nil {
				return err
			}
			defer rc.Close()

			out, err := os.Create(targetPath)
			if err != nil {
				return err
			}
			defer out.Close()

			_, err = io.Copy(out, rc)
			return err
		}
	}

	return fmt.Errorf("binary %s not found in zip", binaryName)
}
