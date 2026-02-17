package generator

import (
	"fmt"
	"strings"

	"github.com/VelociKey/OlympusFabric/20000-MCP-Servers/OlympusFabric/pkg/ir"
)

// PEGGenerator produces PEG/Pigeon compatible grammar files
type PEGGenerator struct{}

func NewPEGGenerator() *PEGGenerator {
	return &PEGGenerator{}
}

func (g *PEGGenerator) FormatName() string {
	return "peg"
}

func (g *PEGGenerator) FileExtension() string {
	return ".peg"
}

func (g *PEGGenerator) Generate(node *ir.Node) (string, error) {
	if node.Type != ir.NodeRoot {
		return "", fmt.Errorf("expected root node, got %s", node.Type)
	}

	var sb strings.Builder
	sb.WriteString("{ \npackage parser\n}\n\n")

	for _, rule := range node.Children {
		if rule.Type == ir.NodeRule {
			content, err := g.generateRule(rule)
			if err != nil {
				return "", err
			}
			sb.WriteString(content)
			sb.WriteString("\n")
		}
	}

	return sb.String(), nil
}

func (g *PEGGenerator) generateRule(n *ir.Node) (string, error) {
	name := n.GetAttributeString("name")
	if len(n.Children) == 0 {
		return "", fmt.Errorf("rule %s has no body", name)
	}

	body, err := g.generateExpression(n.Children[0])
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("%s <- %s", name, body), nil
}

func (g *PEGGenerator) generateExpression(n *ir.Node) (string, error) {
	switch n.Type {
	case ir.NodeChoice:
		var parts []string
		for _, child := range n.Children {
			c, err := g.generateExpression(child)
			if err != nil {
				return "", err
			}
			parts = append(parts, c)
		}
		// In PEG, choice is ALWAYS ordered (/)
		return strings.Join(parts, " / "), nil

	case ir.NodeSequence:
		var parts []string
		for _, child := range n.Children {
			c, err := g.generateExpression(child)
			if err != nil {
				return "", err
			}
			parts = append(parts, c)
		}
		return strings.Join(parts, " "), nil

	case ir.NodeReference:
		return n.GetAttributeString("name"), nil

	case ir.NodeLiteral:
		val := n.GetAttributeString("value")
		// Detect range
		if to, ok := n.GetAttribute("range_to"); ok {
			return fmt.Sprintf("[%s-%s]", val, to), nil
		}
		return fmt.Sprintf(`"%s"`, val), nil

	case ir.NodeOptional:
		inner, err := g.generateExpression(n.Children[0])
		if err != nil {
			return "", err
		}
		return fmt.Sprintf("(%s)?", inner), nil

	case ir.NodeRepeat:
		inner, err := g.generateExpression(n.Children[0])
		if err != nil {
			return "", err
		}
		if n.GetAttributeString("at_least_one") == "true" {
			return fmt.Sprintf("(%s)+", inner), nil
		}
		return fmt.Sprintf("(%s)*", inner), nil

	default:
		return "", fmt.Errorf("unknown node type in PEG generator: %s", n.Type)
	}
}
