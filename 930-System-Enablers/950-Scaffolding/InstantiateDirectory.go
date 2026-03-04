package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// Symantec Web 1 - Scaffolding Implementation (v1.2)
// This program materializes the Sovereign Node structure on an empty workspace.
// Version 1.2.2: Added -DryRun support (default true).

type NodeConfig struct {
	Name        string    `json:"name"`
	Version     string    `json:"version"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"created_at"`
}

func main() {
	// 1. Get the directory where the program is located
	exePath, err := os.Executable()
	if err != nil {
		fmt.Printf("❌ Error determining executable path: %v\n", err)
		return
	}
	exeDir := filepath.Dir(exePath)

	// 2. Setup Flags
	workspaceName := flag.String("name", "SovereignNode-Alpha", "The name of the workspace node")
	pathFlag := flag.String("path", ".wraith", "The target path for the workspace root (defaults to .wraith at program level)")
	dryRun := flag.Bool("DryRun", true, "Perform a dry run without creating directories or files")
	flag.Parse()

	// 3. Resolve Path: If not absolute, anchor it to the program directory
	var absPath string
	if filepath.IsAbs(*pathFlag) {
		absPath = *pathFlag
	} else {
		absPath = filepath.Join(exeDir, *pathFlag)
	}

	absPath = filepath.Clean(absPath)

	if *dryRun {
		fmt.Println("⚠️  DRY RUN MODE: No directories or files will be created.")
	}

	fmt.Printf("🕸️  Materializing Symantec Web 1 Node: %s\n", *workspaceName)
	fmt.Printf("📍 Target: %s\n\n", absPath)

	// 4. Define the Protocol Handshake & Primary Domains
	structure := []string{
		".well-known",
		"olympus.fleet/00SDLC/Olympus2/00000-Identity/010-vision",
		"olympus.fleet/00SDLC/Olympus2/00000-Identity/020-blueprints",
		"olympus.fleet/00SDLC/Olympus2/00000-Identity/030-usage",
		"olympus.fleet/00SDLC/Olympus2/00000-Identity/040-decisions",
		"olympus.fleet/00SDLC/Olympus2/10000-Agents/110-architect",
		"olympus.fleet/00SDLC/Olympus2/10000-Agents/120-builder",
		"olympus.fleet/00SDLC/Olympus2/10000-Agents/150-inference-compute",
		"olympus.fleet/00SDLC/Olympus2/10000-Agents/160-mesh-negotiation",
		"olympus.fleet/00SDLC/Olympus2/20000-MCP-Servers/210-git-mcp",
		"olympus.fleet/00SDLC/Olympus2/20000-MCP-Servers/220-cloud-mcp",
		"olympus.fleet/00SDLC/Olympus2/20000-MCP-Servers/240-async-primitives",
		"olympus.fleet/00SDLC/Olympus2/30000-Services/310-core",
		"olympus.fleet/00SDLC/Olympus2/30000-Services/320-pkg/010-telemetry",
		"olympus.fleet/00SDLC/Olympus2/30000-Services/320-pkg/020-crypto",
		"olympus.fleet/00SDLC/Olympus2/30000-Services/330-workers",
		"olympus.fleet/00SDLC/Olympus2/40000-Interfaces/430-grpc-definitions",
		"olympus.fleet/00SDLC/Olympus2/40000-Interfaces/440-cli-automation",
		"olympus.fleet/00SDLC/Olympus2/40000-Interfaces/460-agent-gateway",
		"olympus.fleet/00SDLC/Olympus2/40000-Interfaces/470-economic-surface",
		"olympus.fleet/00SDLC/Olympus2/50000-Intelligence/510-grammars",
		"olympus.fleet/00SDLC/Olympus2/50000-Intelligence/520-prompts",
		"olympus.fleet/00SDLC/Olympus2/50000-Intelligence/530-ontologies",
		"olympus.fleet/00SDLC/Olympus2/60000-Data-Storage/610-schemas",
		"olympus.fleet/00SDLC/Olympus2/60000-Data-Storage/630-semantic-telemetry",
		"olympus.fleet/00SDLC/Olympus2/70000-Environment/710-opentofu-infra/010-workstation",
		"olympus.fleet/00SDLC/Olympus2/70000-Environment/710-opentofu-infra/020-gcp",
		"olympus.fleet/00SDLC/Olympus2/70000-Environment/720-dagger-modules/010-build",
		"olympus.fleet/00SDLC/Olympus2/70000-Environment/720-dagger-modules/020-deploy",
		"olympus.fleet/00SDLC/Olympus2/70000-Environment/730-provisioning",
		"olympus.fleet/00SDLC/Olympus2/80000-Governance/810-policies",
		"olympus.fleet/00SDLC/Olympus2/80000-Governance/820-visual-ui",
		"olympus.fleet/00SDLC/Olympus2/80000-Governance/830-testing",
		"olympus.fleet/00SDLC/Olympus2/80000-Governance/840-audit",
		"olympus.fleet/00SDLC/Olympus2/80000-Governance/850-trust-registries",
		"olympus.fleet/00SDLC/Olympus2/80000-Governance/860-sdlc-process",
		"olympus.fleet/00SDLC/Olympus2/90000-Enablement/910-benchmarks",
		"olympus.fleet/00SDLC/Olympus2/90000-Enablement/920-labs",
	}

	// 5. Define the Cognizant Operational Layer
	cognizant := []string{
		"olympus.fleet/00SDLC/Olympus2/C0100-Config/.antigravity",
		"olympus.fleet/00SDLC/Olympus2/C0200-Execution/100-campaigns",
		"olympus.fleet/00SDLC/Olympus2/C0200-Execution/200-dagger-runs",
		"olympus.fleet/00SDLC/Olympus2/C0400-Artifacts/100-oci-images",
		"olympus.fleet/00SDLC/Olympus2/C0400-Artifacts/200-binaries",
		"olympus.fleet/00SDLC/Olympus2/C0400-Artifacts/400-changesets",
		"olympus.fleet/00SDLC/Olympus2/C0500-Agent-Out/100-reasoning",
		"olympus.fleet/00SDLC/Olympus2/C0500-Agent-Out/200-dialogue",
		"olympus.fleet/00SDLC/Olympus2/C0700-System-Work/100-sockets",
		"olympus.fleet/00SDLC/Olympus2/C0990-Scratch",
	}

	// Create Primary Silos
	for _, dir := range structure {
		createDir(filepath.Join(absPath, dir), *dryRun)
	}

	// Create Working Memory
	for _, dir := range cognizant {
		createDir(filepath.Join(absPath, dir), *dryRun)
	}

	// 6. Initialize Sentinel Protocols
	initAgentCard(absPath, *workspaceName, *dryRun)
	initNodeManifest(absPath, *workspaceName, *dryRun)
	initAntigravity(absPath, *dryRun)
	initLicense(absPath, *dryRun)

	if *dryRun {
		fmt.Println("\n✅ Dry run complete. No changes made.")
	} else {
		fmt.Println("\n✅ Scaffolding Complete. Symantec Web 1 Node is anchored.")
		fmt.Printf("olympus.fleet/00SDLC/Olympus2/🚀 Next: Define your vision in %s\n", filepath.Join(absPath, "olympus.fleet/00SDLC/Olympus2/00000-Identity", "olympus.fleet/00SDLC/Olympus2/010-vision"))
	}
}

func createDir(path string, dryRun bool) {
	if dryRun {
		fmt.Printf("📁 [DryRun] Would create: %s\n", path)
		return
	}
	if err := os.MkdirAll(path, 0755); err != nil {
		fmt.Printf("❌ Error creating %s: %v\n", path, err)
	} else {
		fmt.Printf("📁 Materialized: %s\n", path)
	}
}

func initAgentCard(root, name string, dryRun bool) {
	path := filepath.Join(root, ".well-known", "agent.json")
	card := map[string]interface{}{
		"agent_manifest_version": "1.0",
		"identity": map[string]string{
			"did":  fmt.Sprintf("did:wba:%s", name),
			"name": name,
		},
		"capabilities": []string{"discovery", "mesh-negotiation"},
		"endpoints": map[string]string{
			"mcp": "/20000-MCP-Servers",
			"a2a": "/460-agent-gateway",
		},
	}
	writeData(path, card, dryRun)
}

func initNodeManifest(root, name string, dryRun bool) {
	path := filepath.Join(root, "olympus.fleet/00SDLC/Olympus2/00000-Identity", "olympus.fleet/00SDLC/Olympus2/020-blueprints", "olympus.fleet/00SDLC/Olympus2/node-manifest.dsl")
	content := fmt.Sprintf("// Symantec Web 1 - Node Manifest\n// Node: %s\n\nnode_capability {\n  type = \"sovereign-hub\"\n  protocols = [\"mcp\", \"a2a\", \"anp\"]\n  mesh_role = \"arbiter\"\n}\n", name)

	if dryRun {
		fmt.Printf("📄 [DryRun] Would write manifest: %s\n", path)
		return
	}
	os.WriteFile(path, []byte(content), 0644)
}

func initAntigravity(root string, dryRun bool) {
	path := filepath.Join(root, "olympus.fleet/00SDLC/Olympus2/C0100-Config", "olympus.fleet/00SDLC/Olympus2/.antigravity", "olympus.fleet/00SDLC/Olympus2/workspace.json")
	config := map[string]string{
		"managed_by": "antigravity",
		"taxonomy":   "symantec-web-1.2",
		"status":     "initialized",
	}
	writeData(path, config, dryRun)
}

func writeData(path string, data interface{}, dryRun bool) {
	if dryRun {
		fmt.Printf("📄 [DryRun] Would initialize protocol: %s\n", path)
		return
	}
	if _, err := os.Stat(path); os.IsNotExist(err) {
		bytes, _ := json.MarshalIndent(data, "", "  ")
		os.WriteFile(path, bytes, 0644)
		fmt.Printf("📄 Initialized Protocol: %s\n", path)
	}
}

func initLicense(root string, dryRun bool) {
	path := filepath.Join(root, "LICENSE")
	content := `# 🌌 Sovereign Fleet Proprietary License

**Version:** 1.0.0
**Classification:** STRICTLY INTERNAL - CONFIDENTIAL

## 1. Proprietary Nature
This software and all associated documentation ("the Works") are the exclusive property of the Sovereign Fleet and its designated operators. The Works contain highly confidential and trade secret information.

## 2. Usage Restrictions
Access to and use of the Works are strictly limited to authenticated members of the Sovereign Fleet. 
- **NO EXTERNAL DISTRIBUTION:** Distribution, publication, or disclosure of the Works to any secondary party outside the designated Mesh is strictly prohibited.
- **INTERNAL USE ONLY:** The Works may only be used for internal research, development, and operational tasks within authorized Sovereign environments.
- **NO REVERSE ENGINEERING:** Unauthorized reverse engineering, de-compilation, or disassembly of any constituent components is prohibited.

## 3. Compliance and Attestation
Any operator accessing these Works must maintain Level 0 Continuous Attestation status. Failure to comply with these terms will result in immediate revocation of Mesh identity and potential legal/diplomatic countermeasures.

## 4. Notice
(C) 2026 VelociKey / Sovereign Fleet. All rights reserved.
Proprietary and Confidential.
`
	if dryRun {
		fmt.Printf("📄 [DryRun] Would write LICENSE: %s\n", path)
		return
	}
	if _, err := os.Stat(path); os.IsNotExist(err) {
		os.WriteFile(path, []byte(content), 0644)
		fmt.Printf("📄 Initialized License: %s\n", path)
	}
}
