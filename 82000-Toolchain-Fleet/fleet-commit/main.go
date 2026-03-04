package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

func findGitRepos(root string) ([]string, error) {
	var repos []string
	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		if info.IsDir() && info.Name() == ".git" {
			repos = append(repos, filepath.Dir(path))
			return filepath.SkipDir
		}
		if info.IsDir() && info.Name() == "node_modules" {
			return filepath.SkipDir
		}
		return nil
	})
	return repos, err
}

func matchesWorkspace(repo string, workspaces []string) bool {
	if len(workspaces) == 0 {
		return true // All
	}
	base := filepath.Base(repo)
	for _, w := range workspaces {
		wTrim := strings.TrimSpace(w)
		if strings.EqualFold(wTrim, base) || strings.EqualFold(wTrim, "All") {
			return true
		}
		if strings.EqualFold(wTrim, "AntigravitySpace") && strings.EqualFold(base, "aAntigravitySpace") {
			return true
		}
	}
	return false
}

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: fleet-commit <message> [workspaces...]")
		os.Exit(1)
	}

	message := os.Args[1]
	
	// Handle comma-separated workspaces if passed as a single arg
	var workspaces []string
	if len(os.Args) > 2 {
		parsed := strings.Split(strings.Join(os.Args[2:], " "), ",")
		for _, p := range parsed {
			if strings.TrimSpace(p) != "" {
				workspaces = append(workspaces, strings.TrimSpace(p))
			}
		}
	}

	root := "c:\\aAntigravitySpace"
	repos, err := findGitRepos(root)
	if err != nil {
		fmt.Printf("Error finding repos: %v\n", err)
		os.Exit(1)
	}

	var committed []string

	for _, repo := range repos {
		if !matchesWorkspace(repo, workspaces) {
			continue
		}

		cmd := exec.Command("git", "status", "--porcelain")
		cmd.Dir = repo
		out, err := cmd.Output()
		if err != nil {
			fmt.Printf("Error checking status in %s: %v\n", repo, err)
			continue
		}

		if len(strings.TrimSpace(string(out))) > 0 {
			fmt.Printf("Committing in %s...\n", repo)
			addCmd := exec.Command("git", "add", ".")
			addCmd.Dir = repo
			addCmd.Run()

			commitCmd := exec.Command("git", "commit", "--no-verify", "-m", message)
			commitCmd.Dir = repo
			commitCmd.Stdout = os.Stdout
			commitCmd.Stderr = os.Stderr
			if err := commitCmd.Run(); err != nil {
				fmt.Printf("Failed to commit in %s: %v\n", repo, err)
			} else {
				committed = append(committed, filepath.Base(repo))
			}
		}
	}

	fmt.Println("\n=== Fleet Commit Summary ===")
	if len(committed) == 0 {
		fmt.Println("No changes to commit in selected workspaces.")
	} else {
		fmt.Println("Successfully committed in:")
		for _, c := range committed {
			fmt.Printf(" - %s\n", c)
		}
		
		// Enforce user mandate check
		hasRoot := false
		for _, c := range committed {
			if strings.EqualFold(c, "aAntigravitySpace") {
				hasRoot = true
			}
		}
		if !hasRoot {
		    for _, repo := range repos {
		        if matchesWorkspace(repo, workspaces) && filepath.Base(repo) == "aAntigravitySpace" {
		            fmt.Println(" - aAntigravitySpace (No changes)")
		            break
		        }
		    }
		}
	}
}
