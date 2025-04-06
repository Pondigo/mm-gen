package prompt

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"text/template"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"mm-go-agent/pkg/mermaid"
)

// PromptTestSuite is a test suite for the prompt package
type PromptTestSuite struct {
	suite.Suite
	tempDir          string
	templateManager  *TemplateManager
	sampleDiagram    string
	validationResult mermaid.ValidationResult
	sampleGoCode     string
}

// SetupSuite prepares the test suite
func (s *PromptTestSuite) SetupSuite() {
	// Sample Go code for testing
	s.sampleGoCode = `
package main

import "fmt"

func main() {
	fmt.Println("Hello, world!")
}
`

	// Sample mermaid diagram with errors
	s.sampleDiagram = `
classDiagram
  class User {
    +string ID
    +string Name
  }
  class Order {
    +string OrderID
  }
  User --> Order
`

	// Mock validation result
	s.validationResult = mermaid.ValidationResult{
		IsValid: false,
		Diagram: s.sampleDiagram,
		Errors: []mermaid.SyntaxError{
			{
				Message: "Class 'User' not properly defined",
				Line:    3,
				Text:    "  class User {",
			},
		},
		ErrorMsg: "Syntax validation failed",
	}
}

// SetupTest runs before each test
func (s *PromptTestSuite) SetupTest() {
	// Create a temporary directory
	tempDir, err := os.MkdirTemp("", "prompt-test")
	require.NoError(s.T(), err, "Failed to create temp directory")

	// Create templates directory
	templatesDir := filepath.Join(tempDir, "templates")
	err = os.Mkdir(templatesDir, 0755)
	if err != nil {
		os.RemoveAll(tempDir)
		require.NoError(s.T(), err, "Failed to create templates directory")
	}

	// Create sample templates
	templates := map[string]string{
		"basic_diagram.tmpl":     "You are a Mermaid diagram expert\n\n# INSTRUCTIONS\nCreate a Mermaid diagram from this Go code:\n\n```go\n{{.CodeContent}}\n```",
		"class_diagram.tmpl":     "You are a Mermaid diagram expert\n\n# INSTRUCTIONS\nCreate a class diagram from this Go code:\n\n```go\n{{.CodeContent}}\n```",
		"sequence_diagram.tmpl":  "You are a Mermaid diagram expert\n\n# INSTRUCTIONS\nCreate a sequence diagram from this Go code:\n\n```go\n{{.CodeContent}}\n```",
		"flowchart_diagram.tmpl": "You are a Mermaid diagram expert\n\n# INSTRUCTIONS\nCreate a flowchart diagram from this Go code:\n\n```go\n{{.CodeContent}}\n```",
		"project_diagram.tmpl":   "You are a Mermaid diagram expert\n\n# INSTRUCTIONS\nCreate a project diagram from this Go code:\n\n```go\n{{.CodeContent}}\n```",
		"config_diagram.tmpl":    "You are a Mermaid diagram expert\n\n# INSTRUCTIONS\nCreate a config diagram from this Go code:\n\n```go\n{{.CodeContent}}\n```",
		"adapters_diagram.tmpl":  "You are a Mermaid diagram expert\n\n# INSTRUCTIONS\nCreate an adapters diagram from this Go code:\n\n```go\n{{.CodeContent}}\n```",
		"fix_diagram.tmpl":       "You are a Mermaid diagram syntax expert\n\n# CONTEXT\n{{if .RetryInfo}}\n{{.RetryInfo}}\n{{end}}\n\n{{.ValidationResult}}\n\n```mermaid\n{{.Diagram}}\n```",
		"explain_errors.tmpl":    "You are a Mermaid diagram syntax expert\n\n{{.ValidationResult}}\n\n```mermaid\n{{.Diagram}}\n```",
	}

	for filename, content := range templates {
		err := os.WriteFile(filepath.Join(templatesDir, filename), []byte(content), 0644)
		if err != nil {
			os.RemoveAll(tempDir)
			require.NoError(s.T(), err, "Failed to write template file %s", filename)
		}
	}

	s.tempDir = tempDir

	// Create template manager
	templatesDir = filepath.Join(tempDir, "templates")
	tmpl, err := template.ParseGlob(filepath.Join(templatesDir, "*.tmpl"))
	require.NoError(s.T(), err, "Failed to parse templates")

	s.templateManager = &TemplateManager{
		templates: tmpl,
	}
}

// TearDownTest runs after each test
func (s *PromptTestSuite) TearDownTest() {
	if s.tempDir != "" {
		os.RemoveAll(s.tempDir)
	}
}

// TestLoadTemplates tests that the template manager can be created
func (s *PromptTestSuite) TestLoadTemplates() {
	assert.NotNil(s.T(), s.templateManager, "Template manager should not be nil")
}

// TestGetDiagramPrompt tests generating prompts for different diagram types
func (s *PromptTestSuite) TestGetDiagramPrompt() {
	// Test cases for each diagram type
	testCases := []struct {
		diagramType mermaid.DiagramType
		expectError bool
	}{
		{mermaid.Basic, false},
		{mermaid.Class, false},
		{mermaid.Sequence, false},
		{mermaid.Flowchart, false},
		{mermaid.Project, false},
		{mermaid.Config, false},
		{mermaid.Adapters, false},
		{"nonexistent", true}, // This should fail as the template doesn't exist
	}

	for _, tc := range testCases {
		s.Run(fmt.Sprintf("DiagramType=%s", tc.diagramType), func() {
			prompt, err := s.templateManager.GetDiagramPrompt(s.sampleGoCode, tc.diagramType)

			if tc.expectError {
				assert.Error(s.T(), err, "Expected error for diagram type %s, but got none", tc.diagramType)
				return
			}

			assert.NoError(s.T(), err, "Unexpected error for diagram type %s", tc.diagramType)

			// Verify the prompt contains the sample code
			assert.Contains(s.T(), prompt, s.sampleGoCode, "Prompt does not contain the sample code for diagram type %s", tc.diagramType)

			// Verify the prompt contains common expected elements
			expectedElements := []string{
				"You are a Mermaid diagram expert",
				"INSTRUCTIONS",
			}

			for _, element := range expectedElements {
				assert.Contains(s.T(), prompt, element, "Prompt does not contain expected element %q for diagram type %s", element, tc.diagramType)
			}
		})
	}
}

// TestGetFixPrompt tests generating prompts for fixing diagrams
func (s *PromptTestSuite) TestGetFixPrompt() {
	// Test getting fix prompt
	prompt, err := s.templateManager.GetFixPrompt(s.sampleDiagram, s.validationResult, 1, 3)
	assert.NoError(s.T(), err, "Unexpected error getting fix prompt")

	// Verify the prompt contains important elements
	expectedElements := []string{
		"Mermaid diagram syntax expert",
		"CONTEXT",
		s.sampleDiagram,
	}

	for _, element := range expectedElements {
		assert.Contains(s.T(), prompt, element, "Fix prompt does not contain expected element %q", element)
	}

	// Test with retry info
	prompt, err = s.templateManager.GetFixPrompt(s.sampleDiagram, s.validationResult, 2, 3)
	assert.NoError(s.T(), err, "Unexpected error getting fix prompt with retry")

	assert.Contains(s.T(), prompt, "This is fix attempt 2/3", "Fix prompt does not contain retry information")
}

// TestGetExplanationPrompt tests generating prompts for explaining errors
func (s *PromptTestSuite) TestGetExplanationPrompt() {
	// Test getting explanation prompt
	prompt, err := s.templateManager.GetExplanationPrompt(s.sampleDiagram, s.validationResult)
	assert.NoError(s.T(), err, "Unexpected error getting explanation prompt")

	// Verify the prompt contains important elements
	expectedElements := []string{
		"Mermaid diagram syntax expert",
		s.sampleDiagram,
	}

	for _, element := range expectedElements {
		assert.Contains(s.T(), prompt, element, "Explanation prompt does not contain expected element %q", element)
	}
}

// TestPromptSuite runs the test suite
func TestPromptSuite(t *testing.T) {
	suite.Run(t, new(PromptTestSuite))
}
