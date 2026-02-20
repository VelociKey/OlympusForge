package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	fleet "OlympusForge/00000-Identity-Foundations/P0000-pkg/000-fleet"
)

func main() {
	if len(os.Args) < 2 {
		printUsage()
		return
	}

	command := os.Args[1]
	cwd, _ := os.Getwd()
	var root string

	// Traverse up from CWD until we find the folder containing "Olympus2"
	// This ensures we always find the absolute root of aAntigravitySpace
	curr := cwd
	for {
		if _, err := os.Stat(filepath.Join(curr, "Olympus2")); err == nil {
			root = curr
			break
		}
		parent := filepath.Dir(curr)
		if parent == curr {
			root = cwd // Fallback
			break
		}
		curr = parent
	}

	switch command {
	case "init":
		runInit(root)
	case "tidy":
		runTidy(root)
	case "create":
		if len(os.Args) < 3 {
			fmt.Println("Usage: sovereign create <moduleName>")
			os.Exit(1)
		}
		runCreate(root, os.Args[2])
	case "check":
		runCheck(root)
	case "start":
		runStart(root)
	case "migrate":
		if len(os.Args) < 4 {
			fmt.Println("Usage: sovereign migrate <oldPath> <newPath>")
			os.Exit(1)
		}
		runMigrate(root, os.Args[2], os.Args[3])
	case "help":
		printUsage()
	default:
		fmt.Printf("Unknown command: %s\n", command)
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Println("Sovereign Fleet CLI")
	fmt.Println("Usage: sovereign <command> [args]")
	fmt.Println("\nCommands:")
	fmt.Println("  init    Synchronize all go.mod and go.work files")
	fmt.Println("  tidy    Run 'go mod tidy' across all modules")
	fmt.Println("  create  Scaffold a new Go module")
	fmt.Println("  check   Verify fleet integrity and structural consistency")
	fmt.Println("  start   Start the Sovereign Workstation Cloud (Infrastructure + Bridges)")
	fmt.Println("  migrate Update import paths and replacements across the fleet")
	fmt.Println("  help    Show this message")
}

func runStart(root string) {
	fmt.Println("🚀 Starting Sovereign Cloud...")
	cmd := exec.Command("go", "run", "orchestrate.go")
	cmd.Dir = filepath.Join(root, "OlympusInfrastructure")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		fmt.Printf("Error starting cloud: %v\n", err)
		os.Exit(1)
	}
}

func runCheck(root string) {
	if err := fleet.VerifyFleet(root); err != nil {
		os.Exit(1)
	}
}

func runCreate(root, moduleName string) {
	if err := fleet.ScaffoldModule(root, moduleName); err != nil {
		fmt.Printf("Error creating module: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("✅ Module scaffolded. Initializing fleet...")
	runInit(root)
	fmt.Println("🧹 Module ready. running tidy...")
	runTidy(root)
}

func runMigrate(root, oldPath, newPath string) {
	fmt.Printf("🚀 Migrating imports from %s to %s...\n", oldPath, newPath)
	if err := fleet.MigrateImports(root, oldPath, newPath); err != nil {
		fmt.Printf("Error during migration: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("✅ Migration complete. Tidying dependencies...")
	runTidy(root)
}

func runInit(root string) {
	fmt.Println("🚀 Initializing Sovereign Fleet...")

	modules, err := fleet.FindGoModules(root)
	if err != nil {
		fmt.Printf("Error scanning modules: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Found %d modules. Syncing go.mod files...\n", len(modules))
	for _, mod := range modules {
		if err := fleet.SyncGoMod(mod, modules, root); err != nil {
			fmt.Printf("Error syncing %s: %v\n", mod.Name, err)
		}
	}

	fmt.Println("Syncing go.work...")
	if err := fleet.SyncGoWork(root, modules); err != nil {
		fmt.Printf("Error syncing go.work: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("✅ Fleet initialization complete.")
}

func runTidy(root string) {
	fmt.Println("🧹 Tidying Fleet Dependencies...")

	modules, err := fleet.FindGoModules(root)
	if err != nil {
		fmt.Printf("Error scanning modules: %v\n", err)
		os.Exit(1)
	}

	for _, mod := range modules {
		fmt.Printf("Tidying %s...\n", mod.Name)
		cmd := exec.Command("go", "mod", "tidy")
		cmd.Dir = mod.Path
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			fmt.Printf("Warning: 'go mod tidy' failed in %s: %v\n", mod.Name, err)
		}
	}

	fmt.Println("Synchronizing workspace...")
	cmd := exec.Command("go", "work", "sync")
	cmd.Dir = root
	if err := cmd.Run(); err != nil {
		fmt.Printf("Error: 'go work sync' failed: %v\n", err)
	}

	fmt.Println("✅ Fleet tidying complete.")
}
