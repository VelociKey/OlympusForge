package generator

import (
	"encoding/json"

	"OlympusForge/20000-Context-Bridges/000-OlympusFabric/P0000-pkg/000-ir"
	"gopkg.in/yaml.v3"
)

// YAMLGenerator produces YAML representations of the IR (for debugging or data DSLs)
type YAMLGenerator struct{}

func NewYAMLGenerator() *YAMLGenerator {
	return &YAMLGenerator{}
}

func (g *YAMLGenerator) FormatName() string {
	return "yaml"
}

func (g *YAMLGenerator) FileExtension() string {
	return ".yaml"
}

func (g *YAMLGenerator) Generate(node *ir.Node) (string, error) {
	// Simple strategy: Convert Node tree to generic map/slice then Marshal
	data := g.nodeToMap(node)
	bytes, err := yaml.Marshal(data)
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}

func (g *YAMLGenerator) nodeToMap(n *ir.Node) interface{} {
	m := make(map[string]interface{})
	m["type"] = n.Type
	if len(n.Attributes) > 0 {
		m["attributes"] = n.Attributes
	}
	if len(n.Children) > 0 {
		children := make([]interface{}, len(n.Children))
		for i, c := range n.Children {
			children[i] = g.nodeToMap(c)
		}
		m["children"] = children
	}
	return m
}

// JSONGenerator produces JSON representation
type JSONGenerator struct{}

func NewJSONGenerator() *JSONGenerator {
	return &JSONGenerator{}
}

func (g *JSONGenerator) FormatName() string {
	return "json"
}

func (g *JSONGenerator) FileExtension() string {
	return ".json"
}

func (g *JSONGenerator) Generate(node *ir.Node) (string, error) {
	bytes, err := json.MarshalIndent(node, "", "  ")
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}
