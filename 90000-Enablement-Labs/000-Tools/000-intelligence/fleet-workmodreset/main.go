package main

import (
	"bufio"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
)

// webPrefixes is a list of prefixes to remove according to mandates.
var webPrefixes = []string{
	"github.com/VelociKey/",
}

// scratchDir is the mandated ephemeral scratch space.
const scratchDir = "00SDLC/Olympus2/C0990-Ephemeral-Scratch"

func main() {
	startDir := flag.String("start", ".", "Starting directory for the process")
	topDir := flag.String("top", ".", "The 'top' of the workspace to finalize work at")
	flag.Parse()

	absStart, _ := filepath.Abs(*startDir)
	absTop, _ := filepath.Abs(*topDir)

	// Ensure scratch dir exists
	os.MkdirAll(scratchDir, 0755)

	scratch1 := filepath.Join(scratchDir, "scratch1.txt")
	scratch2 := filepath.Join(scratchDir, "scratch2.txt")

	// Clear previous scratch files if any
	os.Remove(scratch1)
	os.Remove(scratch2)

	// Perform DFS
	err := processDFS(absStart, absTop, scratch1, scratch2)
	if err != nil {
		slog.Error("Process failed", "error", err)
		os.Exit(1)
	}

	// Final step at the starting directory
	finalizeWorkspace(absStart, scratch2)

	fmt.Println("âœ¨ Successfully completed WorkMod Reset")
}

func processDFS(current, top, s1, s2 string) error {
	// Print a statement for each directory visited
	fmt.Printf("ðŸ“‚ Visiting: %s\n", current)

	entries, err := os.ReadDir(current)
	if err != nil {
		return err
	}

	hasGo := false
	for _, e := range entries {
		if !e.IsDir() && strings.HasSuffix(e.Name(), ".go") {
			hasGo = true
		}
		if e.IsDir() {
			sub := filepath.Join(current, e.Name())
			// Skip standard ignore dirs
			if e.Name() == ".git" || e.Name() == "node_modules" || e.Name() == "gen" || e.Name() == "proto" {
				continue
			}
			err := processDFS(sub, top, s1, s2)
			if err != nil {
				return err
			}
		}
	}

	if hasGo {
		err := handleGoModule(current, s1)
		if err != nil {
			// Silent fail or low-level log for robustness
		}
	}

	if current == top {
		finalizeWorkspace(current, s1)
		appendAndClear(s1, s2)
	}

	return nil
}

func handleGoModule(dir, s1 string) error {
	// 1) Remove webPrefixes from imports in all .go files
	files, _ := os.ReadDir(dir)
	for _, f := range files {
		if !f.IsDir() && strings.HasSuffix(f.Name(), ".go") {
			_ = fixImports(filepath.Join(dir, f.Name()))
		}
	}

	// 2) go mod init [basename]
	os.Remove(filepath.Join(dir, "go.mod"))
	os.Remove(filepath.Join(dir, "go.sum"))
	modName := filepath.Base(dir)
	
	cmd := exec.Command("go", "mod", "init", modName)
	cmd.Dir = dir
	_ = cmd.Run()

	// 3) go build (populates go.mod)
	cmdBuild := exec.Command("go", "build", "./...")
	cmdBuild.Dir = dir
	_ = cmdBuild.Run()

	// 4) Add path to go.mod to scratch1
	f, err := os.OpenFile(s1, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer f.Close()
	_, _ = f.WriteString(dir + "\n")

	return nil
}

func fixImports(path string) error {
	content, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	newContent := content
	for _, prefix := range webPrefixes {
		// More robust regex to catch imports in Various formats
		re := regexp.MustCompile(`"(` + regexp.QuoteMeta(prefix) + `)([^"]+)"`)
		newContent = re.ReplaceAll(newContent, []byte(`"$2"`))
	}

	if string(newContent) != string(content) {
		return os.WriteFile(path, newContent, 0644)
	}
	return nil
}

func finalizeWorkspace(dir, scratchFile string) {
	if _, err := os.Stat(scratchFile); os.IsNotExist(err) {
		return
	}

	goWork := filepath.Join(dir, "go.work")
	if _, err := os.Stat(goWork); os.IsNotExist(err) {
		cmd := exec.Command("go", "work", "init")
		cmd.Dir = dir
		_ = cmd.Run()
	}

	// Read paths from scratch file
	file, _ := os.Open(scratchFile)
	defer file.Close()
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		rel, err := filepath.Rel(dir, scanner.Text())
		if err == nil {
			cmd := exec.Command("go", "work", "use", rel)
			cmd.Dir = dir
			_ = cmd.Run()
		}
	}
}

func appendAndClear(src, dst string) {
	data, err := os.ReadFile(src)
	if err != nil {
		return
	}

	f, err := os.OpenFile(dst, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return
	}
	defer f.Close()
	_, _ = f.Write(data)

	os.Remove(src)
}
