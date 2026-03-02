package main

import (
	"fmt"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"strings"
)

var translationMap = map[string]string{
	"protoimpl \"google.golang.org/protobuf/reflect/protoreflect\"": "protoimpl \"google.golang.org/protobuf/runtime/protoimpl\"",
}

func main() {
	rootDir := "."
	if len(os.Args) > 1 {
		rootDir = os.Args[1]
	}

	fmt.Printf("🚀 Fixing Corrupted Protoimpl Imports in: %s\n", rootDir)

	err := filepath.WalkDir(rootDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if !d.IsDir() && strings.HasSuffix(path, ".go") {
			return processFile(path)
		}

		return nil
	})

	if err != nil {
		log.Fatalf("Fatal: Proto fix failed: %v", err)
	}

	fmt.Println("✅ Protoimpl Imports Fixed.")
}

func processFile(path string) error {
	content, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	s := string(content)
	modified := false

	for old, new := range translationMap {
		if strings.Contains(s, old) {
			s = strings.ReplaceAll(s, old, new)
			modified = true
		}
	}

	if modified {
		fmt.Printf("  Fixed: %s\n", path)
		return os.WriteFile(path, []byte(s), 0644)
	}

	return nil
}
