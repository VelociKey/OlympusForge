package main

import (
	"context"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"

	"olympus.fleet/00SDLC/Olympus2/90000-Enablement-Labs/P0000-pkg/000-search"
)

func main() {
	queryPtr := flag.String("q", "", "Search query")
	rootPtr := flag.String("root", ".", "Root directory of AntigravitySpace")
	limitPtr := flag.Int("limit", 5, "Number of results to return")
	semanticPtr := flag.Bool("semantic", true, "Use semantic search")
	ollamaPtr := flag.Bool("ollama", false, "Use local Ollama for semantic search")
	flag.Parse()

	if *queryPtr == "" {
		fmt.Println("Usage: fleet-search -q \"your query\" [--ollama]")
		os.Exit(1)
	}

	indexPath := filepath.Join(*rootPtr, "Olympus2", "80000-System-Governance", "000-Symbols", "vectors")
	idx, err := search.OpenIndex(indexPath)
	if err != nil {
		slog.Error("Failed to open index", "error", err)
		os.Exit(1)
	}
	defer idx.Close()

	if *semanticPtr {
		var embedder search.Embedder
		
		if *ollamaPtr {
			embedder = search.NewOllamaEmbedder("", "")
		} else {
			apiKey := os.Getenv("GEMINI_API_KEY")
			if apiKey == "" {
				slog.Error("GEMINI_API_KEY required for cloud semantic search. Use --ollama for local.")
				os.Exit(1)
			}

			ctx := context.Background()
			client, err := genai.NewClient(ctx, &genai.ClientConfig{APIKey: apiKey})
			if err != nil {
				slog.Error("Failed to create GenAI client", "error", err)
				os.Exit(1)
			}
			embedder = search.NewGeminiEmbedder(client, "")
		}

		vector, err := embedder.Embed(context.Background(), *queryPtr)
		if err != nil {
			slog.Error("Failed to generate embedding", "error", err)
			os.Exit(1)
		}

		results, err := idx.SearchSemantic(vector, *limitPtr)
		if err != nil {
			slog.Error("Semantic search failed", "error", err)
			os.Exit(1)
		}

		fmt.Printf("\n[SEMANTIC RESULTS] for: \"%s\"\n\n", *queryPtr)
		for _, res := range results {
			icon := "📁"
			if strings.HasPrefix(res.ID, "file:") {
				icon = "📄"
			}
			fmt.Printf("[%0.4f] %s %s\n", res.Score, icon, res.ID)
		}
	} else {
		res, err := idx.Search(*queryPtr, *limitPtr)
		if err != nil {
			slog.Error("Full-text search failed", "error", err)
			os.Exit(1)
		}

		fmt.Printf("\n[FULL-TEXT RESULTS] for: \"%s\"\n\n", *queryPtr)
		for _, hit := range res.Hits {
			fmt.Printf("[%0.4f] %s\n", hit.Score, hit.ID)
		}
	}
}
