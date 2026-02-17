package main

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"time"
)

// TreeExtracter: High-Performance Directory Mapping (Go)
// Replaces: TreeExtracter.ps1

func main() {
	target := "."
	if len(os.Args) > 1 {
		target = os.Args[1]
	}

	absPath, _ := filepath.Abs(target)
	dirName := filepath.Base(absPath)
	outputFile := fmt.Sprintf("zDirectory_Tree_%s.txt", dirName)

	f, err := os.Create(outputFile)
	if err != nil {
		fmt.Printf("❌ Failed to create output: %v\n", err)
		return
	}
	defer f.Close()

	fmt.Fprintf(f, "Directory Tree for: %s\n", absPath)
	fmt.Fprintf(f, "Generated (Go) on: %s\n", time.Now().Format("2006-01-02 15:04:05"))
	fmt.Fprintf(f, "========================================================\n\n")

	walk(absPath, "", true, f)

	fmt.Printf("✅ Tree generated: %s\n", outputFile)
}

func walk(path string, prefix string, isLast bool, f *os.File) {
	entries, err := os.ReadDir(path)
	if err != nil {
		return
	}

	// Filter and Sort: Directories first, then files
	sort.Slice(entries, func(i, j int) bool {
		if entries[i].IsDir() != entries[j].IsDir() {
			return entries[i].IsDir()
		}
		return entries[i].Name() < entries[j].Name()
	})

	for i, entry := range entries {
		if entry.Name() == ".git" || entry.Name() == "node_modules" {
			continue
		}

		last := i == len(entries)-1
		branch := "├── "
		if last {
			branch = "└── "
		}

		line := prefix + branch + entry.Name()
		if entry.IsDir() {
			line += "\\"
		}
		fmt.Fprintln(f, line)

		if entry.IsDir() {
			newPrefix := prefix
			if last {
				newPrefix += "    "
			} else {
				newPrefix += "│   "
			}
			walk(filepath.Join(path, entry.Name()), newPrefix, last, f)
		}
	}
}
