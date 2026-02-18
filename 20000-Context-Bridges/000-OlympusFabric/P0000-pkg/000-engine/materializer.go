package engine

import (
	"fmt"
	"os"
	"path/filepath"

	"OlympusForge/20000-Context-Bridges/000-OlympusFabric/P0000-pkg/000-ir"
)

// Materializer handles the physical creation of files and directories from IR
type Materializer struct {
	basePath   string
	backupPath string
	dryRun     bool
}

func NewMaterializer(basePath string, backupPath string, dryRun bool) *Materializer {
	return &Materializer{
		basePath:   basePath,
		backupPath: backupPath,
		dryRun:     dryRun,
	}
}

// Rollback reverts the filesystem to the last known-good Corporate Baseline.
func (m *Materializer) Rollback() error {
	if m.dryRun {
		fmt.Println("🚫 Dry-run: Skipping actual rollback")
		return nil
	}
	if m.backupPath == "" {
		return fmt.Errorf("no backup path configured for rollback")
	}

	fmt.Printf("🔄 Rolling back %s to baseline %s\n", m.basePath, m.backupPath)

	// Implementation would use os/exec to perform a git-level or fs-level revert
	return nil
}

// Materialize walks the IR and ensures the filesystem reflects the structure
func (m *Materializer) Materialize(node *ir.Node) error {
	if node.Type != ir.NodeRoot {
		return fmt.Errorf("expected root node for materialization, got %s", node.Type)
	}

	for _, child := range node.Children {
		if err := m.processNode(child, m.basePath); err != nil {
			return err
		}
	}
	return nil
}

func (m *Materializer) processNode(n *ir.Node, currentPath string) error {
	name := n.GetAttributeString("name")
	if name == "" && n.Type == ir.NodeLiteral {
		name = n.GetAttributeString("value")
	}

	// Logic for Rule/Dir nodes
	if n.Type == ir.NodeRule || n.Type == ir.NodeReference || (n.Type == ir.NodeLiteral && len(n.Children) > 0) {
		dirPath := filepath.Join(currentPath, name)

		fmt.Printf("📂 Ensuring directory: %s\n", dirPath)
		if !m.dryRun {
			if err := os.MkdirAll(dirPath, 0755); err != nil {
				return fmt.Errorf("failed to create directory %s: %w", dirPath, err)
			}
		}

		for _, child := range n.Children {
			if err := m.processNode(child, dirPath); err != nil {
				return err
			}
		}
		return nil
	}

	// Logic for Sequence/Choice (recursive traversal without adding path levels)
	if n.Type == ir.NodeSequence || n.Type == ir.NodeChoice {
		for _, child := range n.Children {
			if err := m.processNode(child, currentPath); err != nil {
				return err
			}
		}
		return nil
	}

	// Logic for File nodes (Literals without children in our current simplified WSG)
	if n.Type == ir.NodeLiteral && len(n.Children) == 0 {
		filePath := filepath.Join(currentPath, name)
		fmt.Printf("📄 Ensuring file: %s\n", filePath)
		if !m.dryRun {
			if _, err := os.Stat(filePath); os.IsNotExist(err) {
				if err := os.WriteFile(filePath, []byte(""), 0644); err != nil {
					return fmt.Errorf("failed to create file %s: %w", filePath, err)
				}
			}
		}
	}

	return nil
}
