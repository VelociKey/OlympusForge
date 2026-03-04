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
	var workspaces []string
	
	if len(os.Args) > 1 {
		parsed := strings.Split(strings.Join(os.Args[1:], " "), ",")
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

	var published []string

	for _, repo := range repos {
		if !matchesWorkspace(repo, workspaces) {
			continue
		}

		fmt.Printf("Publishing in %s...\n", repo)
		pushCmd := exec.Command("git", "push")
		pushCmd.Dir = repo
		pushCmd.Stdout = os.Stdout
		pushCmd.Stderr = os.Stderr
		if err := pushCmd.Run(); err != nil {
			fmt.Printf("Failed to publish in %s: %v\n", repo, err)
		} else {
			published = append(published, filepath.Base(repo))
		}
	}

	fmt.Println("\n=== Fleet Publish Summary ===")
	if len(published) == 0 {
		fmt.Println("No repositories successfully published in selected workspaces.")
	} else {
		fmt.Println("Successfully published in:")
		for _, p := range published {
			fmt.Printf(" - %s\n", p)
		}
	}
}
