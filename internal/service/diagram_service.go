package service

import (
	"context"
	"fmt"

	"mm-go-agent/internal/adapter/llm"
	"mm-go-agent/internal/repository"
	"mm-go-agent/pkg/mermaid"
)

// DiagramService handles the generation of Mermaid diagrams
type DiagramService interface {
	GenerateDiagram(ctx context.Context, filePath string, diagramType string) (string, error)
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
	default:
		return mermaid.Basic
	}
}
