package generator

import (
	"github.com/VelociKey/OlympusFabric/20000-MCP-Servers/OlympusFabric/pkg/ir"
)

// Generator is the interface that all output formatters must implement
type Generator interface {
	// Generate converts an IR tree into a target string format
	Generate(node *ir.Node) (string, error)

	// FormatName returns the identifier for this generator (e.g., "peg", "yaml")
	FormatName() string

	// FileExtension returns the preferred file extension for the output
	FileExtension() string
}

// Registry manages the available generators
type Registry struct {
	generators map[string]Generator
}

func NewRegistry() *Registry {
	return &Registry{
		generators: make(map[string]Generator),
	}
}

func (r *Registry) Register(g Generator) {
	r.generators[g.FormatName()] = g
}

func (r *Registry) Get(name string) (Generator, bool) {
	g, ok := r.generators[name]
	return g, ok
}

func (r *Registry) List() []string {
	list := make([]string, 0, len(r.generators))
	for name := range r.generators {
		list = append(list, name)
	}
	return list
}
