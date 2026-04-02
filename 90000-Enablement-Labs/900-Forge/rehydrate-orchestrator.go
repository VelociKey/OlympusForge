package main

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func parseLocalTargets(path string) ([]string, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var targets []string
	scanner := bufio.NewScanner(file)
	inLocalBuilds := false

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		if strings.HasPrefix(line, "LocalBuilds") {
			inLocalBuilds = true
			continue
		}

		if inLocalBuilds {
			if line == "];" || line == "]" {
				inLocalBuilds = false
				break
			}
			// Extract "path"
			target := strings.Trim(line, "\", ")
			if target != "" {
				targets = append(targets, target)
			}
		}
	}

	return targets, scanner.Err()
}

func run() error {
	ctx := context.Background()

	manifestPath := filepath.Join("00SDLC", "OlympusBuilder", "C0100-Configuration-Registry", "REHYDRATE_MANIFEST.jebnf")
	targets, err := parseLocalTargets(manifestPath)
	if err != nil {
		return fmt.Errorf("failed to load local targets: %w", err)
	}

	bazelExe := filepath.Join("00SDLC", "OlympusForge", "81000-Toolchain-External", "bazel", "bazel.exe")

	outDir := filepath.Join("00SDLC", "OlympusForge", "82000-Toolchain-Fleet", "bin")
	if err := os.MkdirAll(outDir, 0755); err != nil {
	        return err
	}

	wd, _ := os.Getwd()

	for _, t := range targets {

		// Convert "olympus.fleet/..." to Bazel target "//..."
		bazelTarget := "//" + strings.TrimPrefix(t, "olympus.fleet/")
		name := filepath.Base(t)
		
		fmt.Printf("🔨 Bazel Rehydrating: %s\n", bazelTarget)
		
		cmd := exec.CommandContext(ctx, bazelExe, "--nowindows_enable_symlinks", "build", "--incompatible_stop_exporting_language_modules=false", bazelTarget)    

		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("bazel build failed for %s: %w", bazelTarget, err)
		}

		// Bazel root is in 00SDLC/OlympusBuilder/@CACHE/bazel-root
		// The hash (75zooped) is based on the output_user_root configuration
		bazelRoot := filepath.Join(wd, "00SDLC", "OlympusBuilder", "@CACHE", "bazel-root")
		pkgPath := strings.ReplaceAll(strings.TrimPrefix(bazelTarget, "//"), ":", "/")

		binRoot := filepath.Join(bazelRoot, "75zooped", "execroot", "_main", "bazel-out", "x64_windows-fastbuild", "bin", pkgPath)

		candidates := []string{
		        filepath.Join(binRoot, name+"_", name+".exe"),
		        filepath.Join(binRoot, name+".exe"),
		}

		var srcPath string
		for _, c := range candidates {
		        if _, err := os.Stat(c); err == nil {
		                srcPath = c
		                break
		        }
		}

		if srcPath == "" {
		        return fmt.Errorf("could not find compiled binary for %s in any of %v", name, candidates)
		}
		dstPath := filepath.Join(outDir, name+".exe")
		if err := copyFile(srcPath, dstPath); err != nil {
			return fmt.Errorf("failed to extract compiled %s: %w", name, err)
		}
		fmt.Printf("✅ %s extracted to Fleet Binaries\n", name)
	}

	return nil
}

func copyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()
	
	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()
	
	_, err = io.Copy(out, in)
	return err
}
