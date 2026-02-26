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

	matches := toolBlockRegex.FindAllStringSubmatch(sContent, -1)
	for _, m := range matches {
		toolName := strings.ReplaceAll(m[1], "_", "-")
		body := m[2]

		tool := ToolDefinition{
			Name: toolName,
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
		
		// Determine category by simple text scanning (crude but effective for this format)
		if strings.Contains(sContent, "Foundation") && strings.Contains(sContent, m[0]) { tool.Category = "foundation" }
		if strings.Contains(sContent, "Engineering") && strings.Contains(sContent, m[0]) { tool.Category = "engineering" }
		if strings.Contains(sContent, "Security") && strings.Contains(sContent, m[0]) { tool.Category = "security" }
		if strings.Contains(sContent, "Intelligence") && strings.Contains(sContent, m[0]) { tool.Category = "intelligence" }
		if strings.Contains(sContent, "Infrastructure") && strings.Contains(sContent, m[0]) { tool.Category = "infrastructure" }

		tools = append(tools, tool)
	}

	if len(tools) == 0 {
		return nil, fmt.Errorf("no tools parsed from %s", path)
	}

	return tools, nil
}
