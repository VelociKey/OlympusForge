package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
)

type Registry struct {
	Versions map[string]string
}

func main() {
	dest := flag.String("dest", "", "Destination path for the UI scaffold")
	name := flag.String("name", "olympus_ui", "Internal package name")
	flag.Parse()

	if *dest == "" {
		fmt.Println("Usage: fleet-scaffold-ui -dest <path> [-name <name>]")
		os.Exit(1)
	}

	registry, err := loadRegistry("olympus.fleet/00SDLC/OlympusForge/C0100-Configuration-Registry/DEPENDENCIES.jebnf")
	if err != nil {
		log.Fatalf("Fatal: Could not load registry: %v", err)
	}

	templatePath := "olympus.fleet/00SDLC/OlympusForge/82000-Toolchain-Fleet/templates/flutter-sovereign"
	
	fmt.Printf("🎨 Scaffolding Sovereign UI into: %s\n", *dest)

	err = copyAndHydrate(templatePath, *dest, registry, *name)
	if err != nil {
		log.Fatalf("Fatal: Scaffolding failed: %v", err)
	}

	fmt.Printf("✅ UI Scaffold complete for '%s' using jeBNF versions.\n", *name)
}

func loadRegistry(path string) (*Registry, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	reg := &Registry{Versions: make(map[string]string)}
	scanner := bufio.NewScanner(file)

	var currentName string
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") { continue }

		if strings.HasPrefix(line, "Library") {
			currentName = strings.Trim(strings.TrimPrefix(line, "Library"), " {\"")
			continue
		}

		if line == "}" { currentName = ""; continue }

		if currentName != "" && strings.Contains(line, "Version") {
			parts := strings.SplitN(line, "=", 2)
			val := strings.Trim(strings.TrimSpace(parts[1]), "\";")
			reg.Versions[currentName] = val
		}
	}
	return reg, nil
}

func copyAndHydrate(src, dst string, reg *Registry, appName string) error {
	return filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
		if err != nil { return err }
		relPath, _ := filepath.Rel(src, path)
		targetPath := filepath.Join(dst, relPath)

		if info.IsDir() {
			return os.MkdirAll(targetPath, info.Mode())
		}

		if filepath.Base(path) == "pubspec.yaml" {
			return hydratePubspec(path, targetPath, reg, appName)
		}

		return copyFile(path, targetPath)
	})
}

func hydratePubspec(src, dst string, reg *Registry, appName string) error {
	content, err := os.ReadFile(src)
	if err != nil { return err }

	s := string(content)
	s = strings.Replace(s, "name: flutter_sovereign_template", "name: "+appName, 1)
	s = strings.Replace(s, "{{CONNECT_VERSION}}", "^"+reg.Versions["connectrpc-dart"], 1)
	s = strings.Replace(s, "{{CRONET_VERSION}}", "^"+reg.Versions["cronet-http"], 1)

	return os.WriteFile(dst, []byte(s), 0644)
}

func copyFile(src, dst string) error {
	in, err := os.Open(src); if err != nil { return err }
	defer in.Close()
	out, err := os.Create(dst); if err != nil { return err }
	defer out.Close()
	_, err = io.Copy(out, in)
	return err
}
