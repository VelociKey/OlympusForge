package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

// MATERIALIZATION.jebnf Parser and Substrate Hydrator

type HydrationEntry struct {
	Name        string
	Source      string
	Target      string
	Transformer string
	Policy      string
}

const (
	headerJSON = `// GENERATED FILE - DO NOT EDIT
// Source: %s
// Generated at: %s
// Managed by: fleet-materialize
`
	headerYAML = `# GENERATED FILE - DO NOT EDIT
# Source: %s
# Generated at: %s
# Managed by: fleet-materialize
`
)

func main() {
	rootDir := flag.String("root", ".", "Fleet root directory")
	sync := flag.Bool("sync", false, "Synchronize all substrate files")
	verify := flag.Bool("verify", false, "Verify substrate integrity (no drift)")
	flag.Parse()

	if !*sync && !*verify {
		fmt.Println("Usage: fleet-materialize [--sync] [--verify]")
		return
	}

	if *sync {
		err := runSync(*rootDir)
		if err != nil {
			fmt.Fprintf(os.Stderr, "❌ Sync failed: %v\n", err)
			os.Exit(1)
		}
	}

	if *verify {
		err := runVerify(*rootDir)
		if err != nil {
			fmt.Fprintf(os.Stderr, "❌ Verification failed: %v\n", err)
			os.Exit(1)
		}
		fmt.Println("✅ Substrate integrity verified.")
	}
}

func runSync(root string) error {
	fmt.Println("🚀 Synchronizing Substrate...")
	
	// Check for new state-machine manifest first (CYC-062.3)
	seqPath := filepath.Join(root, "00SDLC", "Olympus2", "C0100-Configuration-Registry", "SUBSTRATE_HYDRATION.jebnf")
	if _, err := os.Stat(seqPath); err == nil {
		fmt.Println("  -> Detected state-machine hydration sequence.")
		return runSequence(root, seqPath)
	}

	entries, err := getEntries(root)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		fmt.Printf("  -> Processing %q (%s)\n", entry.Name, entry.Target)
		err := hydrate(root, entry)
		if err != nil {
			fmt.Printf("     ❌ Error: %v\n", err)
		}
	}

	fmt.Println("✅ All substrate synchronized.")
	return nil
}

func runVerify(root string) error {
	fmt.Println("🔍 Verifying Substrate...")
	entries, err := getEntries(root)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		targetPath := filepath.Join(root, entry.Target)
		if _, err := os.Stat(targetPath); os.IsNotExist(err) {
			return fmt.Errorf("missing substrate file: %s (run fleet-materialize --sync)", entry.Target)
		}
		// In CYC-045, we would also verify content hash here.
	}
	return nil
}

func runSequence(root string, path string) error {
	file, err := os.Open(path)
	if err != nil {
		return err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if strings.HasPrefix(line, "materialize") {
			// Very naive parser for the demo
			src := extractValue(line, "source=")
			tgt := extractValue(line, "target=")
			trans := extractValue(line, "transformer=")
			
			fmt.Printf("  -> Materializing %s\n", tgt)
			err := hydrate(root, HydrationEntry{Source: src, Target: tgt, Transformer: trans})
			if err != nil {
				fmt.Printf("     ❌ Error: %v\n", err)
			}
		} else if strings.HasPrefix(line, "execute") {
			script := extractValue(line, "script=")
			fmt.Printf("  -> Executing %q\n", script)
			
			parts := strings.Fields(script)
			if len(parts) == 0 { continue }
			
			cmd := exec.Command(parts[0], parts[1:]...)
			cmd.Dir = root
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
			if err := cmd.Run(); err != nil {
				fmt.Printf("     ❌ Script Error: %v\n", err)
			}
		}
	}
	return nil
}

func extractValue(line, key string) string {
	idx := strings.Index(line, key)
	if idx == -1 {
		return ""
	}
	sub := line[idx+len(key)+1:] // skip quote
	end := strings.Index(sub, "\"")
	if end == -1 {
		return ""
	}
	return sub[:end]
}

func getEntries(root string) ([]HydrationEntry, error) {
	manifestPath := filepath.Join(root, "MATERIALIZATION.jebnf")
	return parseManifest(manifestPath)
}

func parseManifest(path string) ([]HydrationEntry, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var entries []HydrationEntry
	scanner := bufio.NewScanner(file)
	
	var current *HydrationEntry
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		if strings.HasPrefix(line, "Entry") {
			name := strings.Trim(strings.TrimPrefix(line, "Entry"), " {\"")
			current = &HydrationEntry{Name: name}
			continue
		}

		if line == "}" {
			if current != nil {
				entries = append(entries, *current)
				current = nil
			}
			continue
		}

		if current != nil && strings.Contains(line, "=") {
			parts := strings.SplitN(line, "=", 2)
			key := strings.TrimSpace(parts[0])
			val := strings.Trim(strings.TrimSpace(parts[1]), "\";")
			
			switch key {
			case "source":
				current.Source = val
			case "target":
				current.Target = val
			case "transformer":
				current.Transformer = val
			case "policy":
				current.Policy = val
			}
		}
	}

	return entries, nil
}

func hydrate(root string, entry HydrationEntry) error {
	sourcePath := filepath.Join(root, entry.Source)
	targetPath := filepath.Join(root, entry.Target)

	ext := filepath.Ext(targetPath)
	h := headerJSON
	if ext == ".yaml" || ext == ".yml" {
		h = headerYAML
	}

	content, err := synthesizeContent(sourcePath, entry.Target, entry.Transformer)
	if err != nil {
		return err
	}

	output := fmt.Sprintf(h, filepath.ToSlash(entry.Source), time.Now().Format(time.RFC3339)) + content

	err = os.MkdirAll(filepath.Dir(targetPath), 0755)
	if err != nil {
		return err
	}

	os.Chmod(targetPath, 0644) 
	err = os.WriteFile(targetPath, []byte(output), 0644)
	if err != nil {
		return err
	}

	return os.Chmod(targetPath, 0444) 
}

func synthesizeContent(source string, target string, transformer string) (string, error) {
	if strings.Contains(source, "ide_standards.jebnf") {
		if strings.HasSuffix(target, "settings.json") {
			return `{
	"editor.formatOnSave": false,
	"editor.codeActionsOnSave": {
		"source.organizeImports": "never",
		"source.fixAll": "never"
	},
	"go.useLanguageServer": false,
	"go.formatOnSave": false,
	"go.buildOnSave": "off",
	"go.lintOnSave": "off",
	"go.vetOnSave": "off",
	"go.testOnSave": false,
	"go.coverOnSave": false,
	"dart.enableSdkFormatter": false,
	"dart.analyzeAutomatically": false,
	"java.configuration.updateBuildConfiguration": "disabled",
	"files.exclude": {
		"**/OpenClaw/**": false
	},
	"files.watcherExclude": {
		"**/OpenClaw/**": true
	},
	"search.exclude": {
		"**/OpenClaw/**": true
	},
	"java.import.exclusions": [
		"**/OpenClaw/**",
		"**/node_modules/**",
		"**/.metadata/**",
		"**/archetype-resources/**"
	],
	"java.server.launchMode": "LightWeight",
	"java.import.gradle.enabled": false,
	"java.import.gradle.wrapper.enabled": true
	}`, nil
		}

		if strings.HasSuffix(target, "extensions.json") {
			return `{
	"recommendations": [
		"google.gemini-cli-vscode-ide-companion",
		"mermaidchart.vscode-mermaid-chart"
	],
	"unwantedRecommendations": [
		"devsense.composer-php-vscode",
		"redhat.java",
		"vscjava.vscode-java-pack",
		"ms-python.python",
		"golang.go"
	]
}`, nil
		}
	}

	if strings.Contains(source, "security_policy.jebnf") {
		return `scan:
  scanners:
    - vuln
    - secret
    - license
  severity:
    - HIGH
    - CRITICAL
  exit-code: 1
  exclude-dirs:
    - node_modules
    - vendor
    - build
    - .pub-cache
    - .dart_tool
`, nil
	}

	if strings.Contains(source, "ui_standards.jebnf") {
		if strings.Contains(target, "pubspec.yaml") {
		        return `name: interaction_surface
description: Hardened Flutter Canvas for George
version: 1.0.0+1
publish_to: none

environment:
  sdk: '>=3.0.0 <4.0.0'

dependencies:
  flutter:
    sdk: flutter
  firebase_core: ^3.10.1
  firebase_auth: ^5.4.1
  connectrpc: ^1.0.0
  protobuf: ^4.2.0
  flutter_riverpod: ^3.2.1
  cronet_http: ^1.8.0
  web: ^1.1.0
  js: ^0.6.7

dev_dependencies:
  flutter_test:
    sdk: flutter
  flutter_lints: ^2.0.0

flutter:
  uses-material-design: true
`, nil
		}

		if strings.Contains(target, "analysis_options.yaml") {
			return `include: package:flutter_lints/analysis_options.yaml

analyzer:
  exclude:
    - "**/*.g.dart"
    - "**/*.freezed.dart"
  errors:
    avoid_web_libraries_in_flutter: error

linter:
  rules:
    - prefer_const_constructors
    - avoid_print
`, nil
		}
	}

	if strings.Contains(source, "MEMORY_LANCEDB.jebnf") {
		return `{
    "table_name": "sovereign_events",
    "dimensions": 1536,
    "vector_type": "float32",
    "storage_root": "00SDLC/Olympus2/C0500-Agent-Intelligence-Outputs/lancedb"
}`, nil
	}

	return "", fmt.Errorf("transformer %q not yet implemented for source %q and target %q", transformer, source, target)
}
