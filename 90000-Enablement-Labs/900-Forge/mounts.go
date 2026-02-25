package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"dagger.io/dagger"
)

// MountConfig represents a single declarative mount from PROVISION.jebnf
type MountConfig struct {
	Name   string
	Source string
	Target string
	Type   string // ReadOnly or ReadWrite
}

// EmulatorConfig represents a required local emulator
type EmulatorConfig struct {
	Name string
	Host string
}

// ProvisionConfig holds the parsed workspace provisioning requirements
type ProvisionConfig struct {
	Mounts    []MountConfig
	Emulators []EmulatorConfig
}

// loadProvisionConfig parses a workspace's PROVISION.jebnf (minimal implementation)
func (m *AihubForge) loadProvisionConfig(workspace string) (*ProvisionConfig, error) {
	path := filepath.Join(workspace, "PROVISION.jebnf")
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return &ProvisionConfig{}, nil // Optional file
	}

	content, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	config := &ProvisionConfig{}
	lines := strings.Split(string(content), "\n")

	// Minimal parser logic for prototype
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") || strings.HasPrefix(line, "mounts") || strings.HasPrefix(line, "emulators") || line == "{" || line == "}" {
			continue
		}

		if strings.Contains(line, "source=") && strings.Contains(line, "target=") {
			// Basic extraction for: Name : source=S, target=T, type=M
			parts := strings.Split(line, ":")
			if len(parts) < 2 {
				continue
			}
			name := strings.TrimSpace(parts[0])
			attrStr := parts[1]

			mount := MountConfig{Name: name}
			attrs := strings.Split(attrStr, ",")
			for _, attr := range attrs {
				kv := strings.Split(strings.TrimSpace(attr), "=")
				if len(kv) != 2 {
					continue
				}
				key := strings.TrimSpace(kv[0])
				val := strings.Trim(strings.TrimSpace(kv[1]), "\"'; ")

				switch key {
				case "source":
					mount.Source = val
				case "target":
					mount.Target = val
				case "type":
					mount.Type = val
				}
			}
			config.Mounts = append(config.Mounts, mount)
		} else if strings.Contains(line, "host=") {
			// Basic extraction for Name : host=H
			parts := strings.Split(line, ":")
			if len(parts) < 2 {
				continue
			}
			name := strings.TrimSpace(parts[0])
			attrStr := parts[1]

			emu := EmulatorConfig{Name: name}
			kv := strings.Split(strings.TrimSpace(attrStr), "=")
			if len(kv) == 2 {
				emu.Host = strings.Trim(strings.TrimSpace(kv[1]), "\"'; ")
				config.Emulators = append(config.Emulators, emu)
			}
		}
	}

	return config, nil
}

// applyMountsToContainer injects the declarative mounts into a Dagger container
func (m *AihubForge) applyMountsToContainer(client *dagger.Client, container *dagger.Container, config *ProvisionConfig) *dagger.Container {
	for _, mount := range config.Mounts {
		// Resolve relative source path to fleet root
		hostDir := client.Host().Directory(mount.Source)
		container = container.WithMountedDirectory(mount.Target, hostDir)
		fmt.Printf("   📦 Mounted: %s -> %s [%s]\n", mount.Source, mount.Target, mount.Type)
	}

	// Inject Emulator ENV variables
	for _, emu := range config.Emulators {
		envName := strings.ToUpper(emu.Name) + "_HOST"
		container = container.WithEnvVariable(envName, emu.Host)
		fmt.Printf("   🔌 Emulator: %s -> %s\n", envName, emu.Host)
	}

	return container
}

// generatePodmanRun creates the 'podman run' command with named mounts
func (m *AihubForge) generatePodmanRun(workspace string, config *ProvisionConfig) string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("podman run -d --name %s \\\n", strings.ToLower(workspace)))

	for _, mount := range config.Mounts {
		mode := "ro"
		if mount.Type == "ReadWrite" {
			mode = "rw"
		}
		sb.WriteString(fmt.Sprintf("  -v %s:%s:%s \\\n", mount.Source, mount.Target, mode))
	}

	for _, emu := range config.Emulators {
		envName := strings.ToUpper(emu.Name) + "_HOST"
		sb.WriteString(fmt.Sprintf("  -e %s=%s \\\n", envName, emu.Host))
	}

	sb.WriteString(fmt.Sprintf("  localhost/%s:latest", strings.ToLower(workspace)))
	return sb.String()
}
