package main

import (
	"context"
	"flag"
	"fmt"
	"os"

	"olympus.fleet/00SDLC/Olympus2/90000-Enablement-Labs/P0000-pkg/000-substrate"
)

// fleet-substrate is a multi-call Wasm/Native binary.

func main() {
	if len(os.Args) < 2 {
		usage()
		os.Exit(1)
	}

	cmd := os.Args[1]
	args := os.Args[2:]

	switch cmd {
	case "index":
		runIndex(args)
	case "materialize":
		runMaterialize(args)
	case "search":
		runSearch(args)
	case "soul-gen":
		fmt.Println("Substrate: Soul-Gen not yet fully migrated.")
	default:
		fmt.Printf("Unknown command: %s\n", cmd)
		usage()
		os.Exit(1)
	}
}

func usage() {
	fmt.Println("Usage: fleet-substrate <cmd> [args...]")
	fmt.Println("Commands: index, materialize, search, soul-gen")
}

func runIndex(args []string) {
	fs := flag.NewFlagSet("index", flag.ExitOnError)
	rootDir := fs.String("root", ".", "Root directory containing all workspaces")
	jebnfOnly := fs.Bool("jebnf", false, "Output fleet-wide jeBNF to stdout only")
	singleWS := fs.String("ws", "", "Regenerate only for a specific workspace")
	outputDir := fs.String("out", "", "Output directory")
	semantic := fs.Bool("semantic", false, "Generate semantic embeddings")
	fs.Parse(args)

	opts := substrate.IndexOptions{
		RootDir:   *rootDir,
		OutputDir: *outputDir,
		Semantic:  *semantic,
		SingleWS:  *singleWS,
		JebnfOnly: *jebnfOnly,
	}

	if err := substrate.RunIndex(context.Background(), opts); err != nil {
		fmt.Fprintf(os.Stderr, "❌ Index failed: %v\n", err)
		os.Exit(1)
	}
}

func runMaterialize(args []string) {
	fs := flag.NewFlagSet("materialize", flag.ExitOnError)
	rootDir := fs.String("root", ".", "Fleet root directory")
	sync := fs.Bool("sync", false, "Synchronize all substrate files")
	verify := fs.Bool("verify", false, "Verify substrate integrity")
	fs.Parse(args)

	opts := substrate.MaterializeOptions{
		RootDir: *rootDir,
		Sync:    *sync,
		Verify:  *verify,
	}

	if err := substrate.RunMaterialize(opts); err != nil {
		fmt.Fprintf(os.Stderr, "❌ Materialize failed: %v\n", err)
		os.Exit(1)
	}
}

func runSearch(args []string) {
	fs := flag.NewFlagSet("search", flag.ExitOnError)
	query := fs.String("q", "", "Search query")
	rootDir := fs.String("root", ".", "Root directory")
	limit := fs.Int("limit", 5, "Number of results")
	semantic := fs.Bool("semantic", true, "Use semantic search")
	ollama := fs.Bool("ollama", false, "Use local Ollama")
	fs.Parse(args)

	opts := substrate.SearchOptions{
		Query:    *query,
		RootDir:  *rootDir,
		Limit:    *limit,
		Semantic: *semantic,
		Ollama:   *ollama,
	}

	if err := substrate.RunSearch(context.Background(), opts); err != nil {
		fmt.Fprintf(os.Stderr, "❌ Search failed: %v\n", err)
		os.Exit(1)
	}
}
