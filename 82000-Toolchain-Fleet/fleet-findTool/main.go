package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"hash/fnv"
	"io/fs"
	"log/slog"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// ToolEntry represents a discovered tool and its location.
type ToolEntry struct {
	Name     string `json:"name"`
	Path     string `json:"path"`
	Category string `json:"category"`
}

// MPHF represents a Minimum Perfect Hash Function.
type MPHF struct {
	NumKeys uint32   `json:"num_keys"`
	M       uint32   `json:"m"`      // Number of buckets
	Seeds   []uint32 `json:"seeds"`  // Seeds for each bucket
	Table   []string `json:"table"`  // Final mapping of indices to tool names (verification)
	Values  []string `json:"values"` // Final mapping of indices to paths
}

func hash1(key string) uint32 {
	h := fnv.New32a()
	h.Write([]byte(key))
	return h.Sum32()
}

func hash2(key string, seed uint32) uint32 {
	h := fnv.New32a()
	h.Write([]byte(fmt.Sprintf("%s:%d", key, seed)))
	return h.Sum32()
}

// BuildMPHF builds a Minimum Perfect Hash Function from the given entries.
func BuildMPHF(entries []ToolEntry) *MPHF {
	n := uint32(len(entries))
	if n == 0 {
		return &MPHF{}
	}

	// r = average bucket size. A typical value is 2-4.
	r := uint32(2)
	m := (n + r - 1) / r
	buckets := make([][]string, m)
	entryMap := make(map[string]string)

	for _, e := range entries {
		idx := hash1(e.Name) % m
		buckets[idx] = append(buckets[idx], e.Name)
		entryMap[e.Name] = e.Path
	}

	// Sort buckets by size (descending) to fill the table more easily.
	type bucketInfo struct {
		idx  uint32
		keys []string
	}
	sortedBuckets := make([]bucketInfo, m)
	for i := uint32(0); i < m; i++ {
		sortedBuckets[i] = bucketInfo{idx: i, keys: buckets[i]}
	}
	sort.Slice(sortedBuckets, func(i, j int) bool {
		return len(sortedBuckets[i].keys) > len(sortedBuckets[j].keys)
	})

	seeds := make([]uint32, m)
	table := make([]string, n)
	values := make([]string, n)
	occupied := make([]bool, n)

	for _, b := range sortedBuckets {
		if len(b.keys) == 0 {
			continue
		}

		seed := uint32(0)
		for {
			collision := false
			currentIndices := make([]uint32, 0, len(b.keys))
			for _, k := range b.keys {
				idx := hash2(k, seed) % n
				if occupied[idx] {
					collision = true
					break
				}
				for _, prevIdx := range currentIndices {
					if idx == prevIdx {
						collision = true
						break
					}
				}
				if collision {
					break
				}
				currentIndices = append(currentIndices, idx)
			}

			if !collision {
				// Success! Mark as occupied and save seed.
				for i, k := range b.keys {
					idx := currentIndices[i]
					occupied[idx] = true
					table[idx] = k
					values[idx] = entryMap[k]
				}
				seeds[b.idx] = seed
				break
			}
			seed++
			if seed > 1000000 { // Safety break
				panic("Failed to find perfect hash seed")
			}
		}
	}

	return &MPHF{
		NumKeys: n,
		M:       m,
		Seeds:   seeds,
		Table:   table,
		Values:  values,
	}
}

func (m *MPHF) Lookup(key string) (string, bool) {
	if m.NumKeys == 0 {
		return "", false
	}
	bucketIdx := hash1(key) % m.M
	seed := m.Seeds[bucketIdx]
	idx := hash2(key, seed) % m.NumKeys
	if m.Table[idx] == key {
		return m.Values[idx], true
	}
	return "", false
}

func scanTools(root string) []ToolEntry {
	var entries []ToolEntry

	// Areas to scan (Restricted to relevant OlympusForge areas)
	areas := []struct {
		path string
		cat  string
	}{
		{filepath.Join(root, "00SDLC/OlympusForge/81000-Toolchain-External"), "external"},
		{filepath.Join(root, "00SDLC/OlympusForge/82000-Toolchain-Fleet"), "fleet"},
		{filepath.Join(root, "00SDLC/OlympusForge/90000-Enablement-Labs/000-Tools"), "internal"},
	}

	for _, area := range areas {
		slog.Info("Scanning area", "path", area.path)
		
		filepath.WalkDir(area.path, func(path string, d fs.DirEntry, err error) error {
			if err != nil {
				return nil // Skip errors
			}
			
			// Skip hidden directories and large metadata folders
			if d.IsDir() {
				name := d.Name()
				if strings.HasPrefix(name, ".") || name == "node_modules" || name == ".git" {
					return filepath.SkipDir
				}
				return nil
			}

			name := d.Name()
			isExec := false
			toolName := name

			// Check for executables
			if strings.HasSuffix(strings.ToLower(name), ".exe") {
				isExec = true
				toolName = strings.TrimSuffix(name, filepath.Ext(name))
			} else if strings.HasSuffix(strings.ToLower(name), ".cmd") || strings.HasSuffix(strings.ToLower(name), ".bat") {
				isExec = true
				toolName = strings.TrimSuffix(name, filepath.Ext(name))
			} else if !strings.Contains(name, ".") {
				// Likely a Linux/Unix style binary or a tool directory
				// In this context, we check if it's in a 'bin' folder or matches common patterns
				isExec = true 
			}

			if isExec {
				// Normalize path
				cleanPath := filepath.ToSlash(path)
				
				// Deduplicate: prefer higher-level binaries or specific patterns
				exists := false
				for i, e := range entries {
					if e.Name == toolName {
						// If we find a shorter path (higher level), replace it
						if len(cleanPath) < len(e.Path) {
							entries[i].Path = cleanPath
						}
						exists = true
						break
					}
				}

				if !exists {
					entries = append(entries, ToolEntry{
						Name:     toolName,
						Path:     cleanPath,
						Category: area.cat,
					})
				}
			}
			return nil
		})
	}

	return entries
}

func main() {
	rootPtr := flag.String("root", ".", "AntigravitySpace root")
	findPtr := flag.String("find", "", "Tool name to lookup")
	rebuildPtr := flag.Bool("rebuild", false, "Rebuild the mPSH index")
	indexPathPtr := flag.String("index", "", "Path to save/load the index")
	flag.Parse()

	level := slog.LevelInfo
	if os.Getenv("LOG_LEVEL") == "DEBUG" {
		level = slog.LevelDebug
	}
	handler := slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: level})
	slog.SetDefault(slog.New(handler))

	root, _ := filepath.Abs(*rootPtr)
	indexPath := *indexPathPtr
	if indexPath == "" {
		indexPath = filepath.Join(root, "00SDLC/OlympusForge/82000-Toolchain-Fleet/TOOL_INDEX.jebnf")
	}

	var mph *MPHF
	if *rebuildPtr || *findPtr == "" {
		entries := scanTools(root)
		slog.Info("Building Minimum Perfect Hash", "count", len(entries))
		mph = BuildMPHF(entries)

		// Save as JEBNF-wrapped JSON for high-performance loading
		data, _ := json.MarshalIndent(mph, "", "  ")
		var sb strings.Builder
		sb.WriteString("::mPSH::ToolRegistry::v1\n")
		sb.WriteString(fmt.Sprintf("generated_at = %q\n", os.Getenv("USER"))) // or some timestamp
		sb.WriteString("Data ")
		sb.WriteString(string(data))
		sb.WriteString("\n")
		
		err := os.MkdirAll(filepath.Dir(indexPath), 0755)
		if err != nil {
			slog.Error("Failed to create index directory", "error", err)
			os.Exit(1)
		}
		err = os.WriteFile(indexPath, []byte(sb.String()), 0644)
		if err != nil {
			slog.Error("Failed to write index", "error", err)
			os.Exit(1)
		}
		slog.Info("Generated mPSH Index", "path", indexPath)
	}

	if *findPtr != "" {
		if mph == nil {
			// Load from index
			content, err := os.ReadFile(indexPath)
			if err != nil {
				slog.Warn("Index not found, scanning on-the-fly", "error", err)
				entries := scanTools(root)
				mph = BuildMPHF(entries)
			} else {
				// Parse JEBNF wrapper
				sContent := string(content)
				start := strings.Index(sContent, "{")
				end := strings.LastIndex(sContent, "}")
				if start != -1 && end != -1 {
					err := json.Unmarshal([]byte(sContent[start:end+1]), &mph)
					if err != nil {
						slog.Error("Failed to unmarshal MPHF", "error", err)
					} else {
						slog.Info("Successfully loaded index", "count", mph.NumKeys)
					}
				}
			}
		}

		if mph != nil {
			path, ok := mph.Lookup(*findPtr)
			if ok {
				fmt.Println(path)
			} else {
				slog.Error("Tool not found", "tool", *findPtr)
				os.Exit(1)
			}
		}
	}
}
