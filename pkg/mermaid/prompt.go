package mermaid

import "fmt"

// DiagramType represents the type of Mermaid diagram to generate
type DiagramType string

const (
	// Basic diagram type
	Basic DiagramType = "basic"
	// Sequence diagram type
	Sequence DiagramType = "sequence"
	// Class diagram type
	Class DiagramType = "class"
	// Flowchart diagram type
	Flowchart DiagramType = "flowchart"
	// Project diagram type for project-wide mapping
	Project DiagramType = "project"
	// Config diagram type for showing config interactions
	Config DiagramType = "config"
	// Adapters diagram type for showing inbound/outbound communications
	Adapters DiagramType = "adapters"
)

// CreatePrompt creates a prompt for generating a Mermaid diagram
func CreatePrompt(codeContent string, diagramType DiagramType) string {
	var prompt string

	switch diagramType {
	case Basic:
		prompt = fmt.Sprintf("Please create a basic Mermaid diagram that shows the main components and their relationships from this Go code:\n\n```go\n%s\n```\n\nProvide only the Mermaid diagram syntax without any explanation or markdown formatting.", codeContent)
	case Sequence:
		prompt = fmt.Sprintf("Please create a Mermaid sequence diagram that shows the flow of execution and method calls from this Go code:\n\n```go\n%s\n```\n\nProvide only the Mermaid diagram syntax without any explanation or markdown formatting.", codeContent)
	case Class:
		prompt = fmt.Sprintf("Please create a Mermaid class diagram that shows the struct definitions, their fields, methods, and relationships from this Go code:\n\n```go\n%s\n```\n\nProvide only the Mermaid diagram syntax without any explanation or markdown formatting.", codeContent)
	case Flowchart:
		prompt = fmt.Sprintf("Please create a Mermaid flowchart diagram that shows the control flow from this Go code:\n\n```go\n%s\n```\n\nProvide only the Mermaid diagram syntax without any explanation or markdown formatting.", codeContent)
	case Project:
		prompt = fmt.Sprintf("Please create a Mermaid diagram that shows the overall project architecture based on this Go code. Include all major components and their relationships:\n\n```go\n%s\n```\n\nProvide only the Mermaid diagram syntax without any explanation or markdown formatting.", codeContent)
	case Config:
		prompt = fmt.Sprintf("Please create a Mermaid diagram that shows how configuration is structured and used throughout the application based on this Go code:\n\n```go\n%s\n```\n\nProvide only the Mermaid diagram syntax without any explanation or markdown formatting.", codeContent)
	case Adapters:
		prompt = fmt.Sprintf("Please create a Mermaid diagram that shows all inbound and outbound communications in the application, focusing on adapter components and their interactions with external systems:\n\n```go\n%s\n```\n\nProvide only the Mermaid diagram syntax without any explanation or markdown formatting.", codeContent)
	default:
		prompt = fmt.Sprintf("Please create a Mermaid diagram that represents this Go code:\n\n```go\n%s\n```\n\nProvide only the Mermaid diagram syntax without any explanation or markdown formatting.", codeContent)
	}

	return prompt
}

// FormatOutput formats the diagram output with Mermaid markers
func FormatOutput(diagram string) string {
	return fmt.Sprintf("```mermaid\n%s\n```", diagram)
}
