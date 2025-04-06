package diagram

import (
	"strings"
)

// Processor handles diagram content processing and cleanup
type Processor struct{}

// NewProcessor creates a new diagram processor
func NewProcessor() *Processor {
	return &Processor{}
}

// CleanDiagramOutput processes the diagram output to fix formatting issues
func (p *Processor) CleanDiagramOutput(diagram string) string {
	// Remove outer classDiagram wrapper if there are nested diagrams
	lines := strings.Split(diagram, "\n")

	// Check if we have nested mermaid diagrams with ```mermaid markers
	if strings.Contains(diagram, "```mermaid") {
		var result strings.Builder

		// Extract and merge all mermaid code blocks
		inMermaidBlock := false
		firstDiagramType := ""

		for _, line := range lines {
			// Skip lines with percentage comments
			if strings.HasPrefix(strings.TrimSpace(line), "%") {
				continue
			}

			if strings.Contains(line, "```mermaid") {
				inMermaidBlock = true
				continue
			} else if strings.HasPrefix(line, "```") && inMermaidBlock {
				inMermaidBlock = false
				result.WriteString("\n")
				continue
			}

			if inMermaidBlock {
				// Capture the first diagram type we encounter
				if firstDiagramType == "" && (strings.HasPrefix(strings.TrimSpace(line), "classDiagram") ||
					strings.HasPrefix(strings.TrimSpace(line), "sequenceDiagram") ||
					strings.HasPrefix(strings.TrimSpace(line), "flowchart") ||
					strings.HasPrefix(strings.TrimSpace(line), "graph")) {
					firstDiagramType = strings.TrimSpace(line)
				}

				// Don't add duplicate diagram type declarations
				if strings.TrimSpace(line) != firstDiagramType || firstDiagramType == "" {
					result.WriteString(line)
					result.WriteString("\n")
				}
			}
		}

		// Ensure we have the diagram type at the beginning
		if firstDiagramType != "" {
			return firstDiagramType + "\n" + result.String()
		}

		return result.String()
	}

	// No nested diagrams, just clean up any percentage comments
	var result strings.Builder
	for _, line := range lines {
		if !strings.HasPrefix(strings.TrimSpace(line), "%") {
			result.WriteString(line)
			result.WriteString("\n")
		}
	}

	return strings.TrimSpace(result.String())
}

// ExtractComponentSections extracts different component sections from a project map diagram
func (p *Processor) ExtractComponentSections(diagram string) map[string]string {
	sections := make(map[string]string)

	lines := strings.Split(diagram, "\n")
	currentComponent := ""
	var currentContent strings.Builder

	// Look for section markers like "% MODEL components" or "% SERVICE components"
	for _, line := range lines {
		trimmedLine := strings.TrimSpace(line)

		// Check if this is a component section marker
		if strings.HasPrefix(trimmedLine, "%") && strings.Contains(strings.ToLower(trimmedLine), "component") {
			// If we were already in a component section, save it
			if currentComponent != "" && currentContent.Len() > 0 {
				sections[currentComponent] = currentContent.String()
				currentContent.Reset()
			}

			// Extract the component name from the marker
			parts := strings.Fields(trimmedLine)
			if len(parts) >= 2 {
				currentComponent = strings.ToLower(strings.TrimSuffix(parts[1], "s"))
			}

			continue
		}

		// If we're in a component section, add this line to the content
		if currentComponent != "" {
			currentContent.WriteString(line)
			currentContent.WriteString("\n")
		}
	}

	// Save the last component section if there is one
	if currentComponent != "" && currentContent.Len() > 0 {
		sections[currentComponent] = currentContent.String()
	}

	return sections
}
