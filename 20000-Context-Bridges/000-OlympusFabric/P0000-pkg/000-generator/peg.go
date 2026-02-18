package generator

import (
	"fmt"
	"strings"

	"OlympusForge/20000-Context-Bridges/000-OlympusFabric/P0000-pkg/000-ir"
)

type PEGGenerator struct{}

func NewPEGGenerator() *PEGGenerator {
	return &PEGGenerator{}
}

func (g *PEGGenerator) FormatName() string {
	return "peg"
}

func (g *PEGGenerator) FileExtension() string {
	return "peg"
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

func (g *PEGGenerator) generateRule(node *ir.Node) (string, error) {
	if node.Type != ir.NodeRule {
		return "", fmt.Errorf("expected rule node, got %s", node.Type)
	}

	ruleName, _ := node.Attributes["name"].(string)
	var expressions []string

	for _, child := range node.Children {
		expr, err := g.generateExpression(child)
		if err != nil {
			return "", err
		}
		expressions = append(expressions, expr)
	}

	return fmt.Sprintf("%s = %s", ruleName, strings.Join(expressions, " / ")), nil
}

func (g *PEGGenerator) generateExpression(node *ir.Node) (string, error) {
	switch node.Type {
	case ir.NodeChoice:
		var parts []string
		for _, child := range node.Children {
			part, err := g.generateExpression(child)
			if err != nil {
				return "", err
			}
			parts = append(parts, part)
		}
		return strings.Join(parts, " / "), nil

	case ir.NodeSequence:
		var parts []string
		for _, child := range node.Children {
			part, err := g.generateExpression(child)
			if err != nil {
				return "", err
			}
			parts = append(parts, part)
		}
		return strings.Join(parts, " "), nil

	case ir.NodeLiteral:
		val, _ := node.Attributes["value"].(string)
		return fmt.Sprintf("\"%s\"", val), nil

	case ir.NodeReference:
		name, _ := node.Attributes["name"].(string)
		return name, nil

	case ir.NodeOptional:
		if len(node.Children) != 1 {
			return "", fmt.Errorf("optional node must have exactly one child")
		}
		child, err := g.generateExpression(node.Children[0])
		if err != nil {
			return "", err
		}
		return fmt.Sprintf("(%s)?", child), nil

	case ir.NodeRepeat:
		if len(node.Children) != 1 {
			return "", fmt.Errorf("repeat node must have exactly one child")
		}
		child, err := g.generateExpression(node.Children[0])
		if err != nil {
			return "", err
		}
		return fmt.Sprintf("(%s)*", child), nil

	default:
		return "", fmt.Errorf("unsupported IR node type for PEG: %s", node.Type)
	}
}
