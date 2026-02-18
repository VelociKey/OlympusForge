package transformer

import (
	"fmt"
	"strings"

	"OlympusForge/20000-Context-Bridges/000-OlympusFabric/P0000-pkg/000-ir"
)

// Transformer handles conversion between different grammar notations
type Transformer struct {
	warnings []string
}

// NewTransformer creates a new transformer instance
func NewTransformer() *Transformer {
	return &Transformer{
		warnings: []string{},
	}
}

// EBNFToJeBNF converts an IR tree from standard EBNF to jeBNF
func (t *Transformer) EBNFToJeBNF(root *ir.Node) (*ir.Node, []string) {
	t.warnings = []string{}
	newRoot := root.Clone()

	// 1. Structural Normalization (Ensure sequences are flat where appropriate)
	t.normalizeStructure(newRoot)

	// 2. Choice Promotion (Convert | to / PEG semantics)
	t.transformChoices(newRoot)

	return newRoot, t.warnings
}

func (t *Transformer) normalizeStructure(n *ir.Node) {
	// Recursive normalization logic
	for _, child := range n.Children {
		t.normalizeStructure(child)
	}
}

func (t *Transformer) transformChoices(n *ir.Node) {
	// Rule: Convert Choice to Ordered Choice
	if n.Type == ir.NodeChoice {
		if val, _ := n.GetAttribute("ordered"); val != true {
			n.SetAttribute("ordered", true)
			// Add warning only for the first one to avoid spam
			if len(t.warnings) == 0 {
				t.warnings = append(t.warnings, "Promoted all Choices (|) to Ordered Choices (/).")
			}
		}
		t.validateOrderedChoice(n)
	}

	for _, child := range n.Children {
		t.transformChoices(child)
	}
}

func (t *Transformer) validateOrderedChoice(n *ir.Node) {
	// Simple overlap detection (e.g., "int" before "float" if both start the same)
	// This is sophisticated to do on an IR, but we can do basic checks if children are literals
	if len(n.Children) < 2 {
		return
	}

	for i := 0; i < len(n.Children)-1; i++ {
		curr := n.Children[i]
		next := n.Children[i+1]

		if curr.Type == ir.NodeLiteral && next.Type == ir.NodeLiteral {
			currVal := curr.GetAttributeString("value")
			nextVal := next.GetAttributeString("value")

			if strings.HasPrefix(nextVal, currVal) && currVal != nextVal {
				t.warnings = append(t.warnings, fmt.Sprintf("Literal overlap: '%s' precedes '%s'. In ordered choice, '%s' will shadow '%s'.", currVal, nextVal, currVal, nextVal))
			}
		}
	}
}

// GetWarnings returns any warnings generated during the last transformation
func (t *Transformer) GetWarnings() []string {
	return t.warnings
}
