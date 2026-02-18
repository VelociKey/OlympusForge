// Filename: main.go
// Copyright: © 2026 VelociKey LLC. All Rights Reserved.
// Version: 1.1.0
// Author: Joseph A. White, III
// Status: Approved
// Approver: Joseph A. White, III
// Timestamp: 2026-02-13T14:40:00Z
// License: Proprietary

package main

import (
	"fmt"
	"os"

	"OlympusForge/20000-Context-Bridges/000-OlympusFabric/000-internal/000-parser"
	"OlympusForge/20000-Context-Bridges/000-OlympusFabric/P0000-pkg/000-engine"
	"OlympusForge/20000-Context-Bridges/000-OlympusFabric/P0000-pkg/000-generator"
	"OlympusForge/20000-Context-Bridges/000-OlympusFabric/P0000-pkg/000-llm" // Added llm import
	"OlympusForge/20000-Context-Bridges/000-OlympusFabric/P0000-pkg/000-transformer"
)

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	command := os.Args[1]

	switch command {
	case "assist":
		if len(os.Args) < 3 {
			fmt.Println("Usage: fabric assist <prompt> [--type grammar|program|explain] [--context <file>]")
			os.Exit(1)
		}
		runAssist(os.Args[2:])
	case "parse", "validate", "transform", "compile", "init":
		if len(os.Args) < 3 {
			printUsage()
			os.Exit(1)
		}
		runFileCommand(command, os.Args[2:])
	default:
		fmt.Printf("Unknown command: %s\n", command)
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Println("Usage: fabric <command> [args]")
	fmt.Println("Commands:")
	fmt.Println("  assist    <prompt> [--type T] [--context C]  Get AI assistance (Gemma 3)")
	fmt.Println("  parse     <file>                          Parse grammar to IR")
	fmt.Println("  validate  <file>                          Validate grammar syntax")
	fmt.Println("  transform <file>                          Transform EBNF to jeBNF IR")
	fmt.Println("  compile   <file> <target>                 Compile to target (peg, yaml, json, ps1)")
	fmt.Println("  init      <file>                          Materialize workspace from blueprint")
}

func runCompile(content string, target string) {
	if target == "" {
		target = "peg" // default
	}

	l := parser.NewLexer(content)
	p := parser.NewParser(l)

	root, err := p.ParseGrammar()
	if err != nil {
		fmt.Printf("❌ Parsing failed: %v\n", err)
		os.Exit(1)
	}

	// Transform to jeBNF IR (ordered choices, etc.)
	t := transformer.NewTransformer()
	jebnfRoot, _ := t.EBNFToJeBNF(root)

	// Setup Registry
	reg := generator.NewRegistry()
	reg.Register(generator.NewPEGGenerator())
	reg.Register(generator.NewYAMLGenerator())
	reg.Register(generator.NewJSONGenerator())
	reg.Register(generator.NewPSGenerator())

	gen, ok := reg.Get(target)
	if !ok {
		fmt.Printf("❌ Unknown target: %s. Available: %v\n", target, reg.List())
		os.Exit(1)
	}

	output, err := gen.Generate(jebnfRoot)
	if err != nil {
		fmt.Printf("❌ Generation failed: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("✅ Compilation to '%s' successful!\n", target)
	fmt.Println(output)
}

func runTransform(content string) {
	l := parser.NewLexer(content)
	p := parser.NewParser(l)

	root, err := p.ParseGrammar()
	if err != nil {
		fmt.Printf("❌ Parsing failed (required for transform): %v\n", err)
		os.Exit(1)
	}

	t := transformer.NewTransformer()
	newRoot, warnings := t.EBNFToJeBNF(root)

	if len(warnings) > 0 {
		fmt.Println("⚠️  Transformation Warnings:")
		for _, w := range warnings {
			fmt.Printf("  • %s\n", w)
		}
	}

	json, _ := newRoot.ToJSON()
	fmt.Println("✅ Transformation successful!")
	fmt.Println(json)
}

func runParse(content string) {
	l := parser.NewLexer(content)
	p := parser.NewParser(l)

	root, err := p.ParseGrammar()
	if err != nil {
		fmt.Printf("❌ Parsing failed: %v\n", err)
		os.Exit(1)
	}

	json, _ := root.ToJSON()
	fmt.Println("✅ Parsing successful!")
	fmt.Println(json)
}

func runInit(content string) {
	l := parser.NewLexer(content)
	p := parser.NewParser(l)

	root, err := p.ParseGrammar()
	if err != nil {
		fmt.Printf("❌ Parsing failed: %v\n", err)
		os.Exit(1)
	}

	// Transform to jeBNF IR
	t := transformer.NewTransformer()
	jebnfRoot, _ := t.EBNFToJeBNF(root)

	// Materialize natively in Go
	mat := engine.NewMaterializer(".", "", false)
	fmt.Println("🚀 Starting Go-native Materialization...")
	if err := mat.Materialize(jebnfRoot); err != nil {
		fmt.Printf("❌ Materialization failed: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("✅ Workspace structure synchronized successfully.")
}

func runValidate(content string) {
	l := parser.NewLexer(content)
	p := parser.NewParser(l)

	_, err := p.ParseGrammar()
	if err != nil {
		fmt.Printf("❌ Validation failed: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("✅ Grammar is valid!")
}

func runFileCommand(command string, args []string) {
	filename := args[0]
	target := ""
	if len(args) > 1 {
		target = args[1]
	}

	content, err := os.ReadFile(filename)
	if err != nil {
		fmt.Printf("Error reading file: %v\n", err)
		os.Exit(1)
	}

	switch command {
	case "parse":
		runParse(string(content))
	case "validate":
		runValidate(string(content))
	case "transform":
		runTransform(string(content))
	case "compile":
		runCompile(string(content), target)
	case "init":
		runInit(string(content))
	}
}

func runAssist(args []string) {
	prompt := args[0]
	pType := llm.PromptExplanation
	model := "gemma3:27b"
	var context []string

	for i := 1; i < len(args); i++ {
		if args[i] == "--type" && i+1 < len(args) {
			pType = llm.PromptType(args[i+1])
			i++
		} else if args[i] == "--model" && i+1 < len(args) {
			model = args[i+1]
			i++
		} else if args[i] == "--context" && i+1 < len(args) {
			ctxPath := args[i+1]
			content, err := os.ReadFile(ctxPath)
			if err == nil {
				context = append(context, string(content))
			}
			i++
		}
	}

	client := llm.NewClient("", model)
	factory := llm.NewPromptFactory()

	fullPrompt := factory.Create(pType, prompt, context...)

	fmt.Printf("🤖 Thinking (Gemma 3)... \n\n")
	response, err := client.Generate(fullPrompt)
	if err != nil {
		fmt.Printf("❌ LLM Error: %v\n", err)
		os.Exit(1)
	}

	fmt.Println(response)
}



