package ir

import "encoding/json"

// NodeType represents the type of IR node
type NodeType string

// Common node types
const (
	NodeRoot      NodeType = "root"
	NodeRule      NodeType = "rule"
	NodeChoice    NodeType = "choice"
	NodeSequence  NodeType = "sequence"
	NodeRepeat    NodeType = "repeat"
	NodeOptional  NodeType = "optional"
	NodeLiteral   NodeType = "literal"
	NodeReference NodeType = "reference"
	NodeProperty  NodeType = "property"
)

// Node represents a node in the IR tree
type Node struct {
	Type       NodeType       `json:"type"`
	Attributes map[string]any `json:"attributes,omitempty"`
	Children   []*Node        `json:"children,omitempty"`
}

// NewNode creates a new IR node
func NewNode(nodeType NodeType) *Node {
	return &Node{
		Type:       nodeType,
		Attributes: make(map[string]any),
		Children:   make([]*Node, 0),
	}
}

// AddChild appends a child node
func (n *Node) AddChild(child *Node) {
	n.Children = append(n.Children, child)
}

// SetAttribute sets an attribute value
func (n *Node) SetAttribute(key string, value any) {
	n.Attributes[key] = value
}

// GetAttribute retrieves an attribute value
func (n *Node) GetAttribute(key string) (any, bool) {
	val, ok := n.Attributes[key]
	return val, ok
}

// GetAttributeString retrieves an attribute as string
func (n *Node) GetAttributeString(key string) string {
	if val, ok := n.Attributes[key]; ok {
		if str, ok := val.(string); ok {
			return str
		}
	}
	return ""
}

// ToJSON converts node to JSON string
func (n *Node) ToJSON() (string, error) {
	bytes, err := json.MarshalIndent(n, "", "  ")
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}

func (n *Node) Clone() *Node {
	newNode := NewNode(n.Type)
	for k, v := range n.Attributes {
		newNode.Attributes[k] = v
	}
	for _, child := range n.Children {
		newNode.AddChild(child.Clone())
	}
	return newNode
}
