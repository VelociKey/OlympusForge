package main

import (
	"fmt"
	"os"
	"regexp"
	"strings"
)

// ParseRegistry extracts ToolDefinitions from a jeBNF registry file.
func ParseRegistry(path string) ([]ToolDefinition, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var tools []ToolDefinition
	sContent := string(content)

	// regex to find all Tool definitions: Tool name { body }
	toolBlockRegex := regexp.MustCompile(`Tool\s+([a-z_]+)\s+\{([^}]+)\}`)
	fieldRegex := regexp.MustCompile(`([a-z_]+)\s*=\s*"([^"]+)"`)

	// Find the positions of categories
	categories := []string{"Foundation", "Engineering", "Security", "Intelligence", "Infrastructure", "SovereignSubstrate", "RustFriends"}
	
	matches := toolBlockRegex.FindAllStringSubmatchIndex(sContent, -1)
	for _, m := range matches {
		toolStart := m[0]
		toolName := strings.ReplaceAll(sContent[m[2]:m[3]], "_", "-")
		body := sContent[m[4]:m[5]]

		tool := ToolDefinition{
			Name: toolName,
		}

		// Determine category by finding the nearest category header above the tool
		tool.Category = "unknown"
		nearestPos := -1
		for _, cat := range categories {
			pos := strings.LastIndex(sContent[:toolStart], cat)
			if pos > nearestPos {
				nearestPos = pos
				tool.Category = strings.ToLower(cat)
			}
		}

		fields := fieldRegex.FindAllStringSubmatch(body, -1)
		for _, f := range fields {
			switch f[1] {
			case "version":
				tool.Version = f[2]
			case "origin":
				tool.Origin = f[2]
			case "package":
				tool.Package = f[2]
			case "binary":
				tool.Binary = f[2]
			}
		}
		tools = append(tools, tool)
	}

	if len(tools) == 0 {
		return nil, fmt.Errorf("no tools parsed from %s", path)
	}

	return tools, nil
}
