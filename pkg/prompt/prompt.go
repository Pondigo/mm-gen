package prompt

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"mm-go-agent/pkg/mermaid"
)

// TemplateManager handles the loading and execution of prompt templates
type TemplateManager struct {
	templates *template.Template
}

// New creates a new TemplateManager
func New() (*TemplateManager, error) {
	// Get the current working directory
	cwd, err := os.Getwd()
	if err != nil {
		return nil, fmt.Errorf("failed to get current directory: %w", err)
	}

	// Find the templates directory relative to the current directory
	// Try different relative paths to handle various test environments
	templatesDir := filepath.Join(cwd, "templates")
	if _, err := os.Stat(templatesDir); os.IsNotExist(err) {
		// If running from project root
		templatesDir = filepath.Join(cwd, "pkg", "prompt", "templates")
		if _, err := os.Stat(templatesDir); os.IsNotExist(err) {
			// If running from a test directory
			templatesDir = filepath.Join(cwd, "..", "prompt", "templates")
		}
	}

	// Load all templates from the directory
	templates, err := template.ParseGlob(filepath.Join(templatesDir, "*.tmpl"))
	if err != nil {
		return nil, fmt.Errorf("failed to parse templates: %w", err)
	}

	return &TemplateManager{
		templates: templates,
	}, nil
}

// DiagramPromptData contains the data for generating a diagram prompt
type DiagramPromptData struct {
	CodeContent string
	DiagramType string
}

// ValidationPromptData contains the data for generating a validation prompt
type ValidationPromptData struct {
	Diagram          string
	ValidationResult string
	RetryInfo        string
	AttemptNum       int
	MaxRetries       int
}

// GetDiagramPrompt generates a prompt for creating a mermaid diagram
func (m *TemplateManager) GetDiagramPrompt(codeContent string, diagramType mermaid.DiagramType) (string, error) {
	templateName := fmt.Sprintf("%s_diagram.tmpl", diagramType)

	data := DiagramPromptData{
		CodeContent: codeContent,
		DiagramType: string(diagramType),
	}

	var buf strings.Builder
	if err := m.templates.ExecuteTemplate(&buf, templateName, data); err != nil {
		return "", fmt.Errorf("failed to execute template %q: %w", templateName, err)
	}

	return buf.String(), nil
}

// GetFixPrompt generates a prompt for fixing a mermaid diagram
func (m *TemplateManager) GetFixPrompt(diagram string, validationResult mermaid.ValidationResult, attemptNum, maxRetries int) (string, error) {
	retryInfo := ""
	if attemptNum > 1 {
		retryInfo = fmt.Sprintf("This is fix attempt %d/%d. Previous attempts still had errors.", attemptNum, maxRetries)
	}

	data := ValidationPromptData{
		Diagram:          diagram,
		ValidationResult: mermaid.ValidationResultAsContext(validationResult),
		RetryInfo:        retryInfo,
		AttemptNum:       attemptNum,
		MaxRetries:       maxRetries,
	}

	var buf strings.Builder
	if err := m.templates.ExecuteTemplate(&buf, "fix_diagram.tmpl", data); err != nil {
		return "", fmt.Errorf("failed to execute template %q: %w", "fix_diagram.tmpl", err)
	}

	return buf.String(), nil
}

// GetExplanationPrompt generates a prompt for explaining mermaid diagram errors
func (m *TemplateManager) GetExplanationPrompt(diagram string, validationResult mermaid.ValidationResult) (string, error) {
	data := ValidationPromptData{
		Diagram:          diagram,
		ValidationResult: mermaid.ValidationResultAsContext(validationResult),
	}

	var buf strings.Builder
	if err := m.templates.ExecuteTemplate(&buf, "explain_errors.tmpl", data); err != nil {
		return "", fmt.Errorf("failed to execute template %q: %w", "explain_errors.tmpl", err)
	}

	return buf.String(), nil
}
