package registry

import (
	"fmt"
	"os"
)

// Lookup finds the absolute path to a tool binary in the Forge.
// It uses the generated symbol table for O(1) matching.
func Lookup(name string) (string, error) {
	path, ok := ToolRegistry[name]
	if !ok {
		return "", fmt.Errorf("tool %q not found in Forge registry", name)
	}

	// Verify existence on disk
	if _, err := os.Stat(path); err != nil {
		return "", fmt.Errorf("tool %q registered at %q but missing from disk", name, path)
	}

	return path, nil
}

// All returns all registered tools
func All() map[string]string {
	return ToolRegistry
}
