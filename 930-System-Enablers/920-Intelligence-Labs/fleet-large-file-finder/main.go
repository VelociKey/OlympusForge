package main

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

const sizeThreshold = 1 * 1024 * 1024 // 1MB

type LargeFile struct {
	Path      string `json:"path"`
	Size      int64  `json:"size_bytes"`
	SizeMB    string `json:"size_mb"`
	Workspace string `json:"workspace"`
}

func main() {
	root, err := os.Getwd()
	if err != nil {
		log.Fatalf("failed to get current directory: %v", err)
	}

	workspaces := getWorkspaces(root)
	var largeFiles []LargeFile

	for _, ws := range workspaces {
		wsPath := filepath.Join(root, ws)
		if isReadOnly(wsPath) {
			continue
		}

		err := filepath.WalkDir(wsPath, func(path string, d fs.DirEntry, err error) error {
			if err != nil {
				return nil // skip errors
			}
			if d.IsDir() {
				// Skip .git and other hidden folders to speed up
				if d.Name() == ".git" || d.Name() == ".dagger" {
					return filepath.SkipDir
				}
				return nil
			}

			info, err := d.Info()
			if err != nil {
				return nil
			}

			if info.Size() > sizeThreshold {
				// Check if ignored by git
				if !isGitIgnored(wsPath, path) {
					relPath, _ := filepath.Rel(root, path)
					largeFiles = append(largeFiles, LargeFile{
						Path:      relPath,
						Size:      info.Size(),
						SizeMB:    fmt.Sprintf("%.2f MB", float64(info.Size())/(1024*1024)),
						Workspace: ws,
					})
				}
			}
			return nil
		})
		if err != nil {
			fmt.Fprintf(os.Stderr, "error walking workspace %s: %v\n", ws, err)
		}
	}

	if len(largeFiles) == 0 {
		fmt.Println("No large unignored files found.")
		return
	}

	output, _ := json.MarshalIndent(largeFiles, "", "  ")
	fmt.Println(string(output))
}

func getWorkspaces(root string) []string {
	entries, err := os.ReadDir(root)
	if err != nil {
		log.Fatalf("failed to read root directory: %v", err)
	}

	var workspaces []string
	for _, entry := range entries {
		if entry.IsDir() && !strings.HasPrefix(entry.Name(), ".") {
			workspaces = append(workspaces, entry.Name())
		}
	}
	return workspaces
}

func isReadOnly(path string) bool {
	readmePath := filepath.Join(path, "README.md")
	data, err := os.ReadFile(readmePath)
	if err != nil {
		return false
	}
	content := strings.ToUpper(string(data))
	return strings.Contains(content, "READ-ONLY RESEARCH") ||
		strings.Contains(content, "WORKSPACE STATUS: READ-ONLY")
}

func isGitIgnored(wsPath, filePath string) bool {
	cmd := exec.Command("git", "-C", wsPath, "check-ignore", filePath)
	err := cmd.Run()
	return err == nil
}
