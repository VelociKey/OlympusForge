package main

import (
	"context"
	"crypto/sha256"
	"flag"
	"fmt"
	"io/fs"
	"log/slog"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	search "olympus.fleet/00SDLC/Olympus2/90000-Enablement-Labs/P0900-Labs/190-Search"
	"google.golang.org/genai"
)

// WorkspaceCategory mirrors the Workspace Taxonomy standard.
type WorkspaceCategory string

const (
	CategorySDLC     WorkspaceCategory = "SDLC"
	CategoryProject  WorkspaceCategory = "Project"
	CategoryResearch WorkspaceCategory = "Research"
	CategorySaaS     WorkspaceCategory = "SaaS"
)

// WorkspaceEntry represents one workspace in the fleet.
type WorkspaceEntry struct {
	Name     string
	Category WorkspaceCategory
	Nodes    []SymbolNode
	Hash     string // SHA-256 of the combined node paths (structural fingerprint)
}

// SymbolNode represents a single populated directory entry.
type SymbolNode struct {
	Workspace     string    `json:"workspace"`
	Path          string    `json:"path"` // Relative path from workspace root
	TaxonomyCode  string    `json:"taxonomy_code"`
	TaxonomyLabel string    `json:"taxonomy_label"`
	FileCount     int       `json:"file_count"`
	FileTypes     []string  `json:"file_types"`
	Hash          string    `json:"hash"` // SHA-256 of the path + files
	Embedding     []float32 `json:"embedding,omitempty"`
}

// skipDirs are directory names to always skip during scanning.
var skipDirs = map[string]bool{
	".git": true, ".dart_tool": true, ".gradle": true,
	"node_modules": true, "vendor": true, "build": true,
	".idea": true, "__pycache__": true,
	"Runner.xcodeproj": true, "Runner.xcworkspace": true,
}

// ignoredFiles are files that don't count toward "non-empty".
var ignoredFiles = map[string]bool{
	".gitkeep":     true,
	".antigravity": true,
	".DS_Store":    true,
	"Thumbs.db":    true,
}

// classifyWorkspace uses naming conventions to categorize workspaces.
func classifyWorkspace(name string) WorkspaceCategory {
	switch {
	case name == "Olympus2" || name == "OlympusFabric" || name == "OlympusForge" ||
		name == "OlympusGrammar" || name == "OlympusVision" || name == "OlympusAssurance" ||
		name == "OlympusAtelier" || name == "OlympusInfrastructure" || name == "OlympusAscent" ||
		name == "OlympusMuse" || name == "OlympusMCP" || name == "OlympusActors-Cognition" ||
		name == "OlympusActors-Delegation" || strings.HasPrefix(name, "OlympusGCP-"):
		return CategorySDLC
	case name == "George" || name == "SovereignWeb":
		return CategoryProject
	case strings.HasPrefix(name, "R0000"):
		return CategoryResearch
	default:
		return CategorySaaS
	}
}

// extractTaxonomyCode pulls the leading code from a dir name like "10000-Autonomous-Actors".
func extractTaxonomyCode(name string) (code string, label string) {
	idx := strings.Index(name, "-")
	if idx > 0 && idx <= 6 {
		return name[:idx], name[idx+1:]
	}
	return "", name
}

// computeStructuralHash generates a SHA-256 fingerprint of all node paths for change detection.
func computeStructuralHash(nodes []SymbolNode) string {
	h := sha256.New()
	for _, n := range nodes {
		h.Write([]byte(n.Path))
		h.Write([]byte{'\n'})
	}
	return fmt.Sprintf("%x", h.Sum(nil))
}

// scanWorkspace walks a workspace and returns all non-empty directories.
func scanWorkspace(wsName, wsPath string) ([]SymbolNode, error) {
	var nodes []SymbolNode

	err := filepath.WalkDir(wsPath, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil // skip inaccessible dirs
		}

		if !d.IsDir() {
			return nil
		}

		// Skip hidden and excluded directories
		if strings.HasPrefix(d.Name(), ".") || skipDirs[d.Name()] {
			return filepath.SkipDir
		}

		// Count real files in this directory (non-recursive, just immediate children)
		entries, readErr := os.ReadDir(path)
		if readErr != nil {
			return nil
		}

		fileCount := 0
		fileTypeSet := make(map[string]bool)
		for _, entry := range entries {
			if entry.IsDir() {
				continue
			}
			if ignoredFiles[entry.Name()] {
				continue
			}
			fileCount++
			ext := strings.ToLower(filepath.Ext(entry.Name()))
			if ext != "" {
				fileTypeSet[ext] = true
			}
		}

		if fileCount == 0 {
			return nil
		}

		relPath, _ := filepath.Rel(wsPath, path)
		if relPath == "." {
			return nil
		}

		relPath = filepath.ToSlash(relPath)

		firstSeg := strings.Split(relPath, "/")[0]
		code, label := extractTaxonomyCode(firstSeg)

		var fileTypes []string
		for ft := range fileTypeSet {
			fileTypes = append(fileTypes, ft)
		}
		sort.Strings(fileTypes)

		// Generate Node Hash for mPSH
		nh := sha256.New()
		nh.Write([]byte(relPath))
		for _, ft := range fileTypes {
			nh.Write([]byte(ft))
		}
		hash := fmt.Sprintf("%x", nh.Sum(nil))

		nodes = append(nodes, SymbolNode{
			Workspace:     wsName,
			Path:          relPath,
			TaxonomyCode:  code,
			TaxonomyLabel: label,
			FileCount:     fileCount,
			FileTypes:     fileTypes,
			Hash:          hash,
		})

		return nil
	})

	sort.Slice(nodes, func(i, j int) bool {
		return nodes[i].Path < nodes[j].Path
	})

	return nodes, err
}

// generateJeBNF writes the fleet-wide symbol table in jeBNF format.
func generateJeBNF(workspaces []WorkspaceEntry, rootPath string, outputPath string) error {
	var sb strings.Builder

	sb.WriteString("# Fleet Symbol Table (jeBNF-mPSH)\n")
	sb.WriteString(fmt.Sprintf("generated_at = \"%s\"\n", time.Now().Format(time.RFC3339)))
	sb.WriteString(fmt.Sprintf("root = \"%s\"\n", filepath.ToSlash(rootPath)))
	sb.WriteString("\n")

	// Fleet summary by category
	catCounts := map[WorkspaceCategory]int{}
	catNodes := map[WorkspaceCategory]int{}
	for _, ws := range workspaces {
		catCounts[ws.Category]++
		catNodes[ws.Category] += len(ws.Nodes)
	}

	sb.WriteString("Summary {\n")
	for _, cat := range []WorkspaceCategory{CategorySDLC, CategoryProject, CategoryResearch, CategorySaaS} {
		if catCounts[cat] > 0 {
			sb.WriteString(fmt.Sprintf("  %s workspaces=%d nodes=%d\n", cat, catCounts[cat], catNodes[cat]))
		}
	}
	sb.WriteString("}\n\n")

	// Per-workspace blocks
	for _, ws := range workspaces {
		sb.WriteString(fmt.Sprintf("Workspace \"%s\" %s %s %d {\n",
			ws.Name, ws.Category, ws.Hash[:16], len(ws.Nodes)))

		for _, node := range ws.Nodes {
			taxCode := node.TaxonomyCode
			if taxCode == "" {
				taxCode = "_"
			}
			types := strings.Join(node.FileTypes, " ")
			if types == "" {
				types = "_"
			}
			// Format: Node taxonomy_code "path" file_count [types]
			sb.WriteString(fmt.Sprintf("  Node %s \"%s\" %d [%s]\n",
				taxCode, node.Path, node.FileCount, types))
		}

		sb.WriteString("}\n\n")
	}

	return os.WriteFile(outputPath, []byte(sb.String()), 0644)
}

// generateWorkspaceJeBNF writes a single workspace symbol table in jeBNF format.
func generateWorkspaceJeBNF(ws WorkspaceEntry, outputDir string) error {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("# %s — Symbol Table (jeBNF-mPSH)\n", ws.Name))
	sb.WriteString(fmt.Sprintf("generated_at = \"%s\"\n", time.Now().Format(time.RFC3339)))
	sb.WriteString(fmt.Sprintf("category = %s\n", ws.Category))
	sb.WriteString(fmt.Sprintf("hash = %s\n", ws.Hash[:16]))
	sb.WriteString(fmt.Sprintf("nodes = %d\n\n", len(ws.Nodes)))

	if len(ws.Nodes) == 0 {
		sb.WriteString("# No populated directories\n")
	} else {
		for _, node := range ws.Nodes {
			taxCode := node.TaxonomyCode
			if taxCode == "" {
				taxCode = "_"
			}
			types := strings.Join(node.FileTypes, " ")
			if types == "" {
				types = "_"
			}
			sb.WriteString(fmt.Sprintf("Node %s \"%s\" %d [%s]\n",
				taxCode, node.Path, node.FileCount, types))
		}
	}

	outPath := filepath.Join(outputDir, ws.Name+".jebnf")
	return os.WriteFile(outPath, []byte(sb.String()), 0644)
}

// generateMarkdown writes a per-workspace symbol table in Markdown (for human consumption).
func generateMarkdown(ws WorkspaceEntry, outputDir string) error {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("# %s — Symbol Table (mPSH)\n\n", ws.Name))
	sb.WriteString(fmt.Sprintf("**Category:** %s  \n", ws.Category))
	sb.WriteString(fmt.Sprintf("**Generated:** %s  \n", time.Now().Format(time.RFC3339)))
	sb.WriteString(fmt.Sprintf("**Hash:** `%s`  \n", ws.Hash[:16]))
	sb.WriteString(fmt.Sprintf("**Populated Nodes:** %d  \n\n", len(ws.Nodes)))

	if len(ws.Nodes) == 0 {
		sb.WriteString("_No populated directories found._\n")
	} else {
		sb.WriteString("| Taxonomy | Path | Files | Types |\n")
		sb.WriteString("| :--- | :--- | :--- | :--- |\n")
		for _, node := range ws.Nodes {
			taxLabel := node.TaxonomyCode
			if taxLabel == "" {
				taxLabel = "—"
			}
			types := strings.Join(node.FileTypes, ", ")
			if types == "" {
				types = "—"
			}
			sb.WriteString(fmt.Sprintf("| `%s` | `%s` | %d | %s |\n",
				taxLabel, node.Path, node.FileCount, types))
		}
	}
	sb.WriteString("\n")

	outPath := filepath.Join(outputDir, ws.Name+".md")
	return os.WriteFile(outPath, []byte(sb.String()), 0644)
}

// generateFleetSummary writes the fleet-wide overview in Markdown.
func generateFleetSummary(workspaces []WorkspaceEntry, outputDir string) error {
	var sb strings.Builder

	sb.WriteString("# Fleet Symbol Table — Overview\n\n")
	sb.WriteString(fmt.Sprintf("**Generated:** %s  \n", time.Now().Format(time.RFC3339)))
	sb.WriteString(fmt.Sprintf("**Total Workspaces:** %d  \n", len(workspaces)))
	sb.WriteString(fmt.Sprintf("**Format:** jeBNF-mPSH (primary) + Markdown (human reference)  \n\n"))

	catCounts := map[WorkspaceCategory]int{}
	catNodes := map[WorkspaceCategory]int{}
	for _, ws := range workspaces {
		catCounts[ws.Category]++
		catNodes[ws.Category] += len(ws.Nodes)
	}

	sb.WriteString("## Categories\n\n")
	sb.WriteString("| Category | Workspaces | Populated Nodes |\n")
	sb.WriteString("| :--- | :--- | :--- |\n")
	for _, cat := range []WorkspaceCategory{CategorySDLC, CategoryProject, CategoryResearch, CategorySaaS} {
		if catCounts[cat] > 0 {
			sb.WriteString(fmt.Sprintf("| **%s** | %d | %d |\n", cat, catCounts[cat], catNodes[cat]))
		}
	}

	sb.WriteString("\n## Workspace Index\n\n")
	sb.WriteString("| Workspace | Category | Nodes | Hash | jeBNF | Markdown |\n")
	sb.WriteString("| :--- | :--- | :--- | :--- | :--- | :--- |\n")
	for _, ws := range workspaces {
		sb.WriteString(fmt.Sprintf("| **%s** | %s | %d | `%s` | [.jebnf](./%s.jebnf) | [.md](./%s.md) |\n",
			ws.Name, ws.Category, len(ws.Nodes), ws.Hash[:16], ws.Name, ws.Name))
	}
	sb.WriteString("\n")

	return os.WriteFile(filepath.Join(outputDir, "FLEET_SYMBOLS.md"), []byte(sb.String()), 0644)
}

func main() {
	rootDir := flag.String("root", ".", "Root directory containing all workspaces")
	jebnfOnly := flag.Bool("jebnf", false, "Output fleet-wide jeBNF to stdout only")
	singleWS := flag.String("ws", "", "Regenerate only for a specific workspace")
	outputDir := flag.String("out", "", "Output directory (default: <root>/conductor/workspace-indices)")
	semantic := flag.Bool("semantic", false, "Generate semantic embeddings (requires GEMINI_API_KEY)")
	flag.Parse()

	outDir := *outputDir
	if outDir == "" {
		outDir = filepath.Join(*rootDir, "conductor", "workspace-indices")
	}

	if !*jebnfOnly {
		if err := os.MkdirAll(outDir, 0755); err != nil {
			slog.Error("Failed to create output directory", "path", outDir, "error", err)
			os.Exit(1)
		}
	}

	entries, err := os.ReadDir(*rootDir)
	if err != nil {
		slog.Error("Failed to read root directory", "path", *rootDir, "error", err)
		os.Exit(1)
	}

	var workspaces []WorkspaceEntry

	indexPath := filepath.Join(outDir, "vectors")
	idx, err := search.OpenIndex(indexPath)
	if err != nil {
		slog.Error("Failed to open search index", "path", indexPath, "error", err)
		os.Exit(1)
	}
	defer idx.Close()

	var embedder search.Embedder
	if *semantic {
		apiKey := os.Getenv("GEMINI_API_KEY")
		if apiKey == "" {
			slog.Error("GEMINI_API_KEY environment variable is required for --semantic mode")
			os.Exit(1)
		}

		ctx := context.Background()
		client, err := genai.NewClient(ctx, &genai.ClientConfig{APIKey: apiKey})
		if err != nil {
			slog.Error("Failed to create GenAI client", "error", err)
			os.Exit(1)
		}
		embedder = search.NewGeminiEmbedder(client, "")
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		name := entry.Name()

		if strings.HasPrefix(name, ".") || name == "node_modules" {
			continue
		}

		if *singleWS != "" && name != *singleWS {
			continue
		}

		wsPath := filepath.Join(*rootDir, name)
		category := classifyWorkspace(name)

		slog.Info("Scanning workspace", "name", name, "category", category)

		nodes, scanErr := scanWorkspace(name, wsPath)
		if scanErr != nil {
			slog.Warn("Error scanning workspace", "name", name, "error", scanErr)
			continue
		}

		hash := computeStructuralHash(nodes)

		ws := WorkspaceEntry{
			Name:     name,
			Category: category,
			Nodes:    nodes,
			Hash:     hash,
		}
		workspaces = append(workspaces, ws)

		if !*jebnfOnly {
			// Generate per-workspace jeBNF
			if err := generateWorkspaceJeBNF(ws, outDir); err != nil {
				slog.Error("Failed to write workspace jeBNF", "workspace", name, "error", err)
			}

			// Generate per-workspace Markdown
			if err := generateMarkdown(ws, outDir); err != nil {
				slog.Error("Failed to write workspace Markdown", "workspace", name, "error", err)
			}

			slog.Info("Indexing symbols and files", "workspace", name)

			// 1. Index Directory Nodes
			for _, node := range nodes {
				var vector []float32
				if *semantic && embedder != nil {
					content := fmt.Sprintf("Workspace: %s\nPath: %s\nLabel: %s\nTypes: %v",
						node.Workspace, node.Path, node.TaxonomyLabel, node.FileTypes)

					vector, err = embedder.Embed(context.Background(), content)
					if err != nil {
						slog.Warn("Failed to generate embedding", "path", node.Path, "error", err)
					}
				}

				docID := fmt.Sprintf("dir:%s/%s", node.Workspace, node.Path)
				if err := idx.IndexDocument(docID, node, vector); err != nil {
					slog.Warn("Failed to index document", "id", docID, "error", err)
				}
			}

			// 2. Index Individual Files (CYC-055.3)
			filepath.WalkDir(wsPath, func(path string, d fs.DirEntry, err error) error {
				if err != nil || d.IsDir() {
					return nil
				}

				ext := strings.ToLower(filepath.Ext(path))
				if ext != ".md" && ext != ".jebnf" && ext != ".go" {
					return nil
				}

				relPath, _ := filepath.Rel(*rootDir, path)
				relPath = filepath.ToSlash(relPath)

				content, err := os.ReadFile(path)
				if err != nil {
					return nil
				}

				var vector []float32
				if *semantic && embedder != nil && len(content) > 0 {
					// Use first 2000 chars for embedding to avoid token limits
					trunc := string(content)
					if len(trunc) > 2000 {
						trunc = trunc[:2000]
					}
					vector, _ = embedder.Embed(context.Background(), trunc)
				}

				docID := fmt.Sprintf("file:%s", relPath)
				meta := map[string]string{
					"workspace": name,
					"path":      relPath,
					"type":      "file",
					"extension": ext,
				}
				idx.IndexDocument(docID, meta, vector)
				return nil
			})

			// 3. Generate Local Navigator (Skip Read-Only)
			if name != "OpenClaw" && name != "Mage" {
				if err := generateLocalNavigator(ws, *rootDir); err != nil {
					slog.Error("Failed to write local Navigator", "workspace", name, "error", err)
				}
			}

			slog.Info("Generated symbol tables and local indices", "workspace", name, "nodes", len(nodes))
		}
	}

	// ... (rest of main)
	// Sort workspaces: category then name
	sort.Slice(workspaces, func(i, j int) bool {
		if workspaces[i].Category != workspaces[j].Category {
			return workspaces[i].Category < workspaces[j].Category
		}
		return workspaces[i].Name < workspaces[j].Name
	})

	if *jebnfOnly {
		// Write jeBNF to stdout
		generateJeBNFToStdout(workspaces, *rootDir)
		return
	}

	// Write fleet-wide jeBNF file
	jebnfPath := filepath.Join(outDir, "FLEET_SYMBOLS.jebnf")
	if err := generateJeBNF(workspaces, *rootDir, jebnfPath); err != nil {
		slog.Error("Failed to write fleet jeBNF", "error", err)
		os.Exit(1)
	}
	slog.Info("Generated fleet jeBNF", "output", jebnfPath)

	// Write fleet summary Markdown
	if err := generateFleetSummary(workspaces, outDir); err != nil {
		slog.Error("Failed to write fleet summary", "error", err)
		os.Exit(1)
	}
	slog.Info("Generated fleet summary", "output", filepath.Join(outDir, "FLEET_SYMBOLS.md"))

	// Write the global Persistent Mesh Symbol Hash (mPSH.jebnf)
	if err := generateMeshPersistentHash(workspaces, outDir, *rootDir); err != nil {
		slog.Error("Failed to write global mPSH", "error", err)
	}

	// Print summary
	totalNodes := 0
	for _, ws := range workspaces {
		totalNodes += len(ws.Nodes)
	}
	fmt.Printf("\n🗂️  Fleet Index Complete: %d workspaces, %d populated nodes\n", len(workspaces), totalNodes)
	fmt.Printf("📍 Output: %s\n", outDir)
	fmt.Printf("📄 Primary: FLEET_SYMBOLS.jebnf (agent-optimized)\n")
	fmt.Printf("📋 Reference: FLEET_SYMBOLS.md (human-readable)\n")
	if *semantic {
		fmt.Printf("🧠 Semantic Index: Enabled\n")
	}
}

// generateJeBNFToStdout writes fleet jeBNF to stdout for piping.
func generateJeBNFToStdout(workspaces []WorkspaceEntry, rootPath string) {
	fmt.Printf("# Fleet Symbol Table (jeBNF-mPSH)\n")
	fmt.Printf("generated_at = \"%s\"\n", time.Now().Format(time.RFC3339))
	fmt.Printf("root = \"%s\"\n\n", filepath.ToSlash(rootPath))

	catCounts := map[WorkspaceCategory]int{}
	catNodes := map[WorkspaceCategory]int{}
	for _, ws := range workspaces {
		catCounts[ws.Category]++
		catNodes[ws.Category] += len(ws.Nodes)
	}

	fmt.Printf("Summary {\n")
	for _, cat := range []WorkspaceCategory{CategorySDLC, CategoryProject, CategoryResearch, CategorySaaS} {
		if catCounts[cat] > 0 {
			fmt.Printf("  %s workspaces=%d nodes=%d\n", cat, catCounts[cat], catNodes[cat])
		}
	}
	fmt.Printf("}\n\n")

	for _, ws := range workspaces {
		fmt.Printf("Workspace \"%s\" %s %s %d {\n",
			ws.Name, ws.Category, ws.Hash[:16], len(ws.Nodes))
		for _, node := range ws.Nodes {
			taxCode := node.TaxonomyCode
			if taxCode == "" {
				taxCode = "_"
			}
			types := strings.Join(node.FileTypes, " ")
			if types == "" {
				types = "_"
			}
			fmt.Printf("  Node %s \"%s\" %d [%s]\n",
				taxCode, node.Path, node.FileCount, types)
		}
		fmt.Printf("}\n\n")
	}
}

// generateLocalNavigator writes a NAVIGATOR.md specific to one workspace.
func generateLocalNavigator(ws WorkspaceEntry, rootPath string) error {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("# 🗺️ %s Navigator\n\n", ws.Name))
	sb.WriteString(fmt.Sprintf("**Category:** %s  \n", ws.Category))
	sb.WriteString(fmt.Sprintf("**Last Synchronized:** %s  \n\n", time.Now().Format("2006-01-02 15:04:05 MST")))

	mission := getMissionStatement(filepath.Join(rootPath, ws.Name))
	if mission != "" {
		sb.WriteString(fmt.Sprintf("> %s\n\n", mission))
	}

	sb.WriteString("This navigator lists only the active surface areas within this workspace, filtering out boilerplate placeholders.\n\n")

	sb.WriteString("| Taxonomy | Implementation Path | Mesh Hash |\n")
	sb.WriteString("| :--- | :--- | :--- |\n")

	for _, node := range ws.Nodes {
		tax := node.TaxonomyCode
		if tax == "" {
			tax = "—"
		}
		// Relative link from workspace root
		dirLink := fmt.Sprintf("[%s](./%s)", node.Path, node.Path)
		// Use node Hash as the mesh reference
		sb.WriteString(fmt.Sprintf("| `%s` | %s | `%s` |\n", tax, dirLink, node.Hash[:8]))
	}

	sb.WriteString("\n---\n*Part of the Sovereign Fleet. Local navigation generated by fleet-index.exe*\n")

	outPath := filepath.Join(rootPath, ws.Name, "NAVIGATOR.md")
	return os.WriteFile(outPath, []byte(sb.String()), 0644)
}

// generateMeshPersistentHash writes a condensed lookup for the fleet and per-workspace.
func generateMeshPersistentHash(workspaces []WorkspaceEntry, outDir string, rootPath string) error {
	var fleetSB strings.Builder
	fleetSB.WriteString("::mPSH::Fleet::v1\n")

	for _, ws := range workspaces {
		if len(ws.Nodes) == 0 {
			continue
		}

		// Entry: Workspace [Name] Category [Cat] Path [Path] Hash [Hash]
		fleetSB.WriteString(fmt.Sprintf("W %s %s %s\n", ws.Name, ws.Category, ws.Hash[:12]))

		// Per-workspace local lookup (mPSH.jebnf)
		var wsSB strings.Builder
		wsSB.WriteString(fmt.Sprintf("::mPSH::%s::v1\n", ws.Name))
		for _, node := range ws.Nodes {
			wsSB.WriteString(fmt.Sprintf("N %s %s\n", node.TaxonomyCode, node.Hash[:12]))
		}

		// Also keep a copy in the central indices
		idxPath := filepath.Join(outDir, fmt.Sprintf("%s.mpsh.jebnf", ws.Name))
		if err := os.WriteFile(idxPath, []byte(wsSB.String()), 0644); err != nil {
			slog.Warn("Failed to write central mPSH", "workspace", ws.Name, "path", idxPath, "error", err)
		}
	}

	fleetPath := filepath.Join(outDir, "FLEET.mpsh.jebnf")
	if err := os.WriteFile(fleetPath, []byte(fleetSB.String()), 0644); err != nil {
		return fmt.Errorf("failed to write fleet mPSH to %s: %w", fleetPath, err)
	}
	slog.Info("Generated Mesh Persistent Symbol Hash (mPSH)", "fleet_path", fleetPath)
	return nil
}

func getMissionStatement(wsPath string) string {
	paths := []string{"PURPOSE.md", "README.md", "BLUEPRINT.md"}
	for _, p := range paths {
		content, err := os.ReadFile(filepath.Join(wsPath, p))
		if err == nil {
			lines := strings.Split(string(content), "\n")
			for _, line := range lines {
				trimmed := strings.TrimSpace(line)
				if trimmed == "" || strings.HasPrefix(trimmed, "#") {
					continue
				}
				// Skip badges and metadata
				if strings.Contains(trimmed, "**Taxonomy:**") ||
					strings.Contains(trimmed, "**Status:**") ||
					strings.Contains(trimmed, "**Fleet:**") {
					continue
				}
				// Return first meaningful sentence/line
				if len(trimmed) > 180 {
					return trimmed[:177] + "..."
				}
				return trimmed
			}
		}
	}
	return ""
}
