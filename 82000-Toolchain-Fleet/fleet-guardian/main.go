package main

import (
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
)

const SizeLimit = 50 * 1024 * 1024 // 50MB

func main() {
	// 1. Get staged files
	cmd := exec.Command("git", "diff", "--cached", "--name-only")
	output, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Printf("Error getting staged files: %v\n", err)
		os.Exit(0) // Don't block commit if git fails for some reason
	}

	files := strings.Split(strings.TrimSpace(string(output)), "\n")
	if len(files) == 1 && files[0] == "" {
		os.Exit(0)
	}

	foundTooBig := false
	for _, file := range files {
		if file == "" {
			continue
		}

		info, err := os.Stat(file)
		if err != nil {
			// File might be deleted or renamed in index, skip
			continue
		}

		if info.Size() > SizeLimit {
			fmt.Printf("❌ ERROR: File too large to commit: %s (%s MB)\n", file, strconv.FormatFloat(float64(info.Size())/(1024*1024), 'f', 2, 64))
			fmt.Printf("👉 REMEDY: Use 'git rm --cached %s' and add to .gitignore.\n", file)
			foundTooBig = true
		}
	}

	if foundTooBig {
		fmt.Println("\n🚨 COMMIT BLOCKED: Sovereign integrity requires binaries to be externalized.")
		os.Exit(1)
	}

	os.Exit(0)
}
