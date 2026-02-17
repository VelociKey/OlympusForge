package main

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: workspace-creator <workspace-name>")
		os.Exit(1)
	}
	name := os.Args[1]
	cwd, _ := os.Getwd()

	// 1. Create VelociKey/<workspace> repository (Simulating the "Origin")
	velociPath := filepath.Join(cwd, "VelociKey", name)
	fmt.Printf("Initializing VelociKey repository at: %s\n", velociPath)
	if err := os.MkdirAll(velociPath, 0755); err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	runGit(velociPath, "init")
	os.WriteFile(filepath.Join(velociPath, "README.md"), []byte("# "+name+"\n"), 0644)
	runGit(velociPath, "add", "README.md")
	runGit(velociPath, "commit", "-m", "Initial semantic blueprint commit")
	runGit(velociPath, "checkout", "-b", "production")
	runGit(velociPath, "checkout", "-b", "staging")
	runGit(velociPath, "checkout", "-b", "development")

	// 2. Create the actual workspace (Cloning from the VelociKey "Origin")
	workspacePath := filepath.Join(cwd, name)
	fmt.Printf("Creating workspace at: %s\n", workspacePath)
	if err := exec.Command("git", "clone", "-b", "development", velociPath, workspacePath).Run(); err != nil {
		fmt.Printf("Error cloning workspace: %v\n", err)
		os.Exit(1)
	}

	// 3. Apply directory structure and needed files from OlympusFabric
	fabricPath := filepath.Join(cwd, "OlympusFabric")
	fmt.Printf("Applying semantic structure from: %s\n", fabricPath)

	// Copy .gitignore from Olympus2 as a default
	gitIgnoreSrc := filepath.Join(cwd, "Olympus2", ".gitignore")
	if _, err := os.Stat(gitIgnoreSrc); err == nil {
		copyFile(gitIgnoreSrc, filepath.Join(workspacePath, ".gitignore"))
	}

	// Recursive copy of OlympusFabric contents (intent-based structure)
	if err := copyDir(fabricPath, workspacePath); err != nil {
		fmt.Printf("Error applying Fabric structure: %v\n", err)
	}

	fmt.Printf("Successfully created workspace: %s\n", name)
}

func runGit(dir string, args ...string) {
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	cmd.Run()
}

func copyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil { return err }
	defer in.Close()
	out, err := os.Create(dst)
	if err != nil { return err }
	defer out.Close()
	_, err = io.Copy(out, in)
	return err
}

func copyDir(src, dst string) error {
	return filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
		if err != nil { return err }
		if path == src { return nil }
		if info.IsDir() && info.Name() == ".git" { return filepath.SkipDir }

		rel, _ := filepath.Rel(src, path)
		target := filepath.Join(dst, rel)

		if info.IsDir() {
			return os.MkdirAll(target, info.Mode())
		}
		return copyFile(path, target)
	})
}
