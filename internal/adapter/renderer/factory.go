package renderer

// NewRenderer creates a new Renderer implementation using mermaid-go
func NewRenderer() Renderer {
	return NewMermaidRenderer()
}
