package service

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"

	"mm-go-agent/internal/adapter/llm"
	"mm-go-agent/internal/repository"
	"mm-go-agent/pkg/mermaid"
)

// DiagramService handles the generation of Mermaid diagrams
type DiagramService interface {
	GenerateDiagram(ctx context.Context, filePath string, diagramType string) (string, error)
	GenerateComponentDiagram(ctx context.Context, componentSpec string, diagramType string) (string, error)
	GenerateProjectDiagram(ctx context.Context, diagramType string) (string, error)
}

// diagramService implements DiagramService
type diagramService struct {
	fileRepo   repository.FileRepository
	llmAdapter llm.LLMAdapter
}

// NewDiagramService creates a new diagram service
func NewDiagramService(fileRepo repository.FileRepository, llmAdapter llm.LLMAdapter) DiagramService {
	return &diagramService{
		fileRepo:   fileRepo,
		llmAdapter: llmAdapter,
	}
}

// GenerateDiagram generates a Mermaid diagram from Go code in the specified file
func (s *diagramService) GenerateDiagram(ctx context.Context, filePath string, diagramType string) (string, error) {
	// Read the file content
	codeContent, err := s.fileRepo.ReadGoFile(filePath)
	if err != nil {
		return "", fmt.Errorf("failed to read Go file: %w", err)
	}

	// Create prompt based on diagram type
	dt := parseDiagramType(diagramType)
	prompt := mermaid.CreatePrompt(codeContent, dt)

	// Generate diagram using LLM
	diagram, err := s.llmAdapter.GenerateCompletion(ctx, prompt)
	if err != nil {
		return "", fmt.Errorf("failed to generate diagram: %w", err)
	}

	// Format and return the diagram
	return mermaid.FormatOutput(diagram), nil
}

// GenerateComponentDiagram generates a Mermaid diagram for a specific component (service, repository, etc.)
func (s *diagramService) GenerateComponentDiagram(ctx context.Context, componentSpec string, diagramType string) (string, error) {
	// Parse component type and name
	parts := strings.Split(componentSpec, ":")
	if len(parts) != 2 {
		return "", fmt.Errorf("invalid component specification: %s (should be 'type:name')", componentSpec)
	}

	componentType := parts[0]
	componentName := parts[1]

	// Get component files
	files, err := s.fileRepo.FindComponentFiles(componentType, componentName)
	if err != nil {
		return "", fmt.Errorf("failed to find %s %s files: %w", componentType, componentName, err)
	}

	if len(files) == 0 {
		return "", fmt.Errorf("no files found for %s %s", componentType, componentName)
	}

	// Read all files content
	var codeContents []string
	for _, file := range files {
		content, err := s.fileRepo.ReadGoFile(file)
		if err != nil {
			return "", fmt.Errorf("failed to read file %s: %w", file, err)
		}
		codeContents = append(codeContents, content)
	}

	// Join all code contents
	allCode := strings.Join(codeContents, "\n\n")

	// Create prompt for component
	dt := parseDiagramType(diagramType)
	prompt := fmt.Sprintf("Please create a %s Mermaid diagram for the %s '%s' from this Go code:\n\n```go\n%s\n```\n\nProvide only the Mermaid diagram syntax without any explanation or markdown formatting.",
		dt, componentType, componentName, allCode)

	// Generate diagram using LLM
	diagram, err := s.llmAdapter.GenerateCompletion(ctx, prompt)
	if err != nil {
		return "", fmt.Errorf("failed to generate diagram: %w", err)
	}

	return mermaid.FormatOutput(diagram), nil
}

// GenerateProjectDiagram generates a project-wide Mermaid diagram
func (s *diagramService) GenerateProjectDiagram(ctx context.Context, diagramType string) (string, error) {
	// Validate diagram type
	if !isValidProjectDiagramType(diagramType) {
		return "", fmt.Errorf("invalid project diagram type: %s (should be 'sequence', 'class', 'config', or 'adapters')", diagramType)
	}

	// Find all relevant files based on diagram type
	var files []string
	var err error

	switch diagramType {
	case "sequence":
		// Get service, repository, and adapter files for sequence diagram
		files, err = s.fileRepo.FindAllComponentFiles([]string{"service", "repository", "adapter"})
	case "class":
		// Get all component files for class diagram
		files, err = s.fileRepo.FindAllComponentFiles([]string{"service", "repository", "adapter", "model", "config"})
	case "config":
		// Get only config files
		files, err = s.fileRepo.FindAllComponentFiles([]string{"config"})
	case "adapters":
		// Get adapter files
		files, err = s.fileRepo.FindAllComponentFiles([]string{"adapter"})
	default:
		// Fallback to all files
		files, err = s.fileRepo.FindAllComponentFiles([]string{"service", "repository", "adapter", "model", "config"})
	}

	if err != nil {
		return "", fmt.Errorf("failed to find files: %w", err)
	}

	if len(files) == 0 {
		return "", fmt.Errorf("no relevant files found for diagram type: %s", diagramType)
	}

	// Read all files content
	var codeContents []string
	for _, file := range files {
		content, err := s.fileRepo.ReadGoFile(file)
		if err != nil {
			return "", fmt.Errorf("failed to read file %s: %w", file, err)
		}
		codeContents = append(codeContents, fmt.Sprintf("// File: %s\n%s", filepath.Base(file), content))
	}

	// Join all code contents
	allCode := strings.Join(codeContents, "\n\n")

	// Create prompt for project diagram
	var prompt string
	switch diagramType {
	case "sequence":
		prompt = fmt.Sprintf("Please create a sequence diagram showing the interactions between all components (services, repositories, adapters) in this Go project. Focus on the flow of calls between different components and how they interact:\n\n```go\n%s\n```\n\nProvide only the Mermaid diagram syntax without any explanation or markdown formatting.", allCode)
	case "class":
		prompt = fmt.Sprintf("Please create a comprehensive class diagram showing all components (services, repositories, adapters, models, config) in this Go project. Show their structs, interfaces, methods, and relationships between them:\n\n```go\n%s\n```\n\nProvide only the Mermaid diagram syntax without any explanation or markdown formatting.", allCode)
	case "config":
		prompt = fmt.Sprintf("Please create a diagram showing how configuration is structured and accessed throughout the application. Show config structs and how other components interact with them:\n\n```go\n%s\n```\n\nProvide only the Mermaid diagram syntax without any explanation or markdown formatting.", allCode)
	case "adapters":
		prompt = fmt.Sprintf("Please create a diagram showing all inbound and outbound communications in the application. Focus on adapter components and how they interact with external systems and internal components:\n\n```go\n%s\n```\n\nProvide only the Mermaid diagram syntax without any explanation or markdown formatting.", allCode)
	default:
		prompt = fmt.Sprintf("Please create a diagram showing the overall architecture of this Go project based on the following code:\n\n```go\n%s\n```\n\nProvide only the Mermaid diagram syntax without any explanation or markdown formatting.", allCode)
	}

	// Generate diagram using LLM
	diagram, err := s.llmAdapter.GenerateCompletion(ctx, prompt)
	if err != nil {
		return "", fmt.Errorf("failed to generate diagram: %w", err)
	}

	return mermaid.FormatOutput(diagram), nil
}

// parseDiagramType converts a string to a DiagramType
func parseDiagramType(dt string) mermaid.DiagramType {
	switch dt {
	case "basic":
		return mermaid.Basic
	case "sequence":
		return mermaid.Sequence
	case "class":
		return mermaid.Class
	case "flowchart":
		return mermaid.Flowchart
	case "project":
		return mermaid.Project
	case "config":
		return mermaid.Config
	case "adapters":
		return mermaid.Adapters
	default:
		return mermaid.Basic
	}
}

// isValidProjectDiagramType checks if a diagram type is valid for project-wide diagrams
func isValidProjectDiagramType(dt string) bool {
	validTypes := []string{"sequence", "class", "config", "adapters"}
	for _, t := range validTypes {
		if dt == t {
			return true
		}
	}
	return false
}
