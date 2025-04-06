package renderer

// Renderer defines the interface for generating SVG from Mermaid diagrams
type Renderer interface {
	// ConvertToSVG converts Mermaid diagram syntax to SVG format
	ConvertToSVG(mermaidContent string) (string, error)
}
