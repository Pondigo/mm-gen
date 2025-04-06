package renderer

// SVGRenderer is an alias for the MermaidRenderer
// This is kept for backward compatibility
type SVGRenderer struct {
	renderer Renderer
}

// NewSVGRenderer creates a new SVG renderer (using MermaidRenderer)
func NewSVGRenderer() *SVGRenderer {
	return &SVGRenderer{
		renderer: NewMermaidRenderer(),
	}
}

// DefaultSVGRenderer creates a default SVG renderer (using MermaidRenderer)
func DefaultSVGRenderer() *SVGRenderer {
	return NewSVGRenderer()
}

// ConvertToSVG converts a mermaid diagram to SVG format using MermaidRenderer
func (r *SVGRenderer) ConvertToSVG(mermaidDiagram string) (string, error) {
	return r.renderer.ConvertToSVG(mermaidDiagram)
}
