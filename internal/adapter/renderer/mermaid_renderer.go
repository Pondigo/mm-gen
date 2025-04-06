package renderer

import (
	"github.com/anz-bank/mermaid-go/mermaid"
)

// MermaidRenderer implements Renderer using mermaid-go library
type MermaidRenderer struct {
	generator *mermaid.Generator
}

// NewMermaidRenderer creates a new MermaidRenderer
func NewMermaidRenderer() *MermaidRenderer {
	return &MermaidRenderer{
		generator: mermaid.Init(),
	}
}

// ConvertToSVG converts Mermaid diagram syntax to SVG format
func (r *MermaidRenderer) ConvertToSVG(mermaidContent string) (string, error) {
	svg := r.generator.Execute(mermaidContent)
	return svg, nil
}
