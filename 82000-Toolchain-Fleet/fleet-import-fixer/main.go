package main

import (
	"fmt"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"strings"
)

// TranslationMap associates old prefixes with new Sovereign prefixes.
var translationMap = map[string]string{
	// External Sovereign Mappings
	"olympus.fleet/00SDLC/OlympusForge/90000-Enablement-Labs/000-Tools/000-external/connect": "olympus.fleet/ext/connectrpc/connect-go",
	"olympus.fleet/00SDLC/OlympusForge/90000-Enablement-Labs/000-Tools/000-external/protobuf": "olympus.fleet/ext/protocolbuffers/protobuf-go",
	"olympus.fleet/00SDLC/OlympusForge/90000-Enablement-Labs/000-Tools/000-external/x/":       "golang.org/x/",
	"olympus.fleet/00SDLC/OlympusForge/ZC0400-Sovereign-Source/connectrpc":                   "olympus.fleet/ext/connectrpc/connect-go",
	"olympus.fleet/00SDLC/OlympusForge/ZC0400-Sovereign-Source/google-genai":                 "olympus.fleet/google-genai",
	"olympus.fleet/00SDLC/OlympusForge/81000-Tools/ext/":                                     "olympus.fleet/ext/",

	// Internal Family Mappings (00SDLC)
	"Olympus2/":               "olympus.fleet/00SDLC/Olympus2/",
	"OlympusForge/":           "olympus.fleet/00SDLC/OlympusForge/",
	"OlympusMCP/":             "olympus.fleet/00SDLC/OlympusMCP/",
	"OlympusGrammar/":         "olympus.fleet/00SDLC/OlympusGrammar/",
	"AssessAgent/":           "olympus.fleet/00SDLC/AssessAgent/",
	"OlympusActors-Cognition/": "olympus.fleet/00SDLC/OlympusActors-Cognition/",
	"OlympusActors-Delegation/": "olympus.fleet/00SDLC/OlympusActors-Delegation/",
	"OlympusAscent/":          "olympus.fleet/00SDLC/OlympusAscent/",
	"OlympusAssurance/":       "olympus.fleet/00SDLC/OlympusAssurance/",
	"OlympusAtelier/":         "olympus.fleet/00SDLC/OlympusAtelier/",
	"OlympusFabric/":          "olympus.fleet/00SDLC/OlympusFabric/",
	"OlympusInfrastructure/":  "olympus.fleet/00SDLC/OlympusInfrastructure/",
	"OlympusVision/":          "olympus.fleet/00SDLC/OlympusVision/",

	// Internal Family Mappings (10GNDT)
	"GND-Clearinghouse/": "olympus.fleet/10GNDT/GND-Clearinghouse/",
	"GND-Customs/":       "olympus.fleet/10GNDT/GND-Customs/",
	"GND-Freight/":       "olympus.fleet/10GNDT/GND-Freight/",
	"GND-Rails/":         "olympus.fleet/10GNDT/GND-Rails/",
	"GND-Registry/":      "olympus.fleet/10GNDT/GND-Registry/",
	"GND-Substrate/":     "olympus.fleet/10GNDT/GND-Substrate/",
	"GNDSAssurance/":     "olympus.fleet/10GNDT/GNDSAssurance/",

	// Internal Family Mappings (20POSI, 30INFR, etc)
	"George/":            "olympus.fleet/20POSI/George/",
	"Pinnacle/":          "olympus.fleet/30INFR/Pinnacle/",
	"PinnacleAssurance/": "olympus.fleet/30INFR/PinnacleAssurance/",
	"PublicOutreach/":    "olympus.fleet/99PUBL/PublicOutreach/",

	// Legacy/Short-Hand Mappings (Literal replacements)
	"\"connect\"":        "\"olympus.fleet/ext/connectrpc/connect-go\"",
	"\"golang.org/x/net":            "\"golang.org/x/net",
	"\"golang.org/x/sys":            "\"golang.org/x/sys",
	"\"golang.org/x/text":           "\"golang.org/x/text",
	"\"olympus.fleet/ext/protocolbuffers/protobuf-go":  "\"olympus.fleet/ext/protocolbuffers/protobuf-go",
}

func main() {
	rootDir := "."
	if len(os.Args) > 1 {
		rootDir = os.Args[1]
	}

	fmt.Printf("🚀 Starting Extended Sovereign Import Normalization in: %s\n", rootDir)

	err := filepath.WalkDir(rootDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if !d.IsDir() && strings.HasSuffix(path, ".go") {
			return processFile(path)
		}

		if d.IsDir() && (d.Name() == ".git" || d.Name() == "node_modules" || d.Name() == ".pub-cache") {
			return filepath.SkipDir
		}

		return nil
	})

	if err != nil {
		log.Fatalf("Fatal: Refactor failed: %v", err)
	}

	fmt.Println("✅ Extended Import Normalization Complete.")
}

func processFile(path string) error {
	content, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	lines := strings.Split(string(content), "\n")
	modified := false

	for i, line := range lines {
		// Only target lines containing quotes or backticks (imports)
		if strings.Contains(line, "\"") || strings.Contains(line, "`") {
			for oldPrefix, newPrefix := range translationMap {
				if strings.Contains(line, oldPrefix) {
					// Guard against re-applying to already fixed lines
					if !strings.Contains(line, newPrefix) || strings.HasPrefix(oldPrefix, "\"") {
						lines[i] = strings.Replace(line, oldPrefix, newPrefix, -1)
						modified = true
					}
				}
			}
		}
	}

	if modified {
		fmt.Printf("  Refactored: %s\n", path)
		return os.WriteFile(path, []byte(strings.Join(lines, "\n")), 0644)
	}

	return nil
}
