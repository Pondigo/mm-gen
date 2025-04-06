package service

import (
	"context"
	"fmt"
	"math/rand"
	"path/filepath"
	"strings"
	"time"

	"mm-go-agent/internal/adapter/llm"
	"mm-go-agent/internal/repository"
	"mm-go-agent/pkg/mermaid"
	"mm-go-agent/pkg/prompt"
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
	promptMgr  *prompt.TemplateManager
}

// NewDiagramService creates a new diagram service
func NewDiagramService(fileRepo repository.FileRepository, llmAdapter llm.LLMAdapter) DiagramService {
	promptMgr, err := prompt.New()
	if err != nil {
		// Fall back to empty manager if templates can't be loaded
		promptMgr = &prompt.TemplateManager{}
	}

	return &diagramService{
		fileRepo:   fileRepo,
		llmAdapter: llmAdapter,
		promptMgr:  promptMgr,
	}
}

// GenerateDiagram generates a Mermaid diagram from Go code in the specified file
func (s *diagramService) GenerateDiagram(ctx context.Context, filePath string, diagramType string) (string, error) {
	// Read the file content
	codeContent, err := s.fileRepo.ReadGoFile(filePath)
	if err != nil {
		return "", fmt.Errorf("failed to read Go file: %w", err)
	}

	// Map the diagram type string to a DiagramType
	dt := s.mapDiagramType(diagramType)

	// Generate the prompt for the diagram
	var promptText string
	if s.promptMgr != nil {
		// Use the template manager if available
		promptText, err = s.promptMgr.GetDiagramPrompt(codeContent, dt)
		if err != nil {
			// Fall back to old method if template fails
			promptText = mermaid.CreatePrompt(codeContent, dt)
		}
	} else {
		// Use the old method if template manager is not available
		promptText = mermaid.CreatePrompt(codeContent, dt)
	}

	// Generate the diagram using the LLM
	diagramText, err := s.llmAdapter.GenerateCompletion(ctx, promptText)
	if err != nil {
		return "", fmt.Errorf("error generating diagram: %w", err)
	}

	// Format the output
	formattedDiagram := mermaid.FormatOutput(diagramText)

	// Validate and fix the diagram if needed
	validationResult := mermaid.ValidateSyntax(formattedDiagram)

	if !validationResult.IsValid {
		fmt.Printf("Diagram for file %s has syntax errors, attempting to fix...\n", filePath)

		// Create validation service for fixing diagrams
		validationService := NewValidationService(llm.NewClientAdapter(s.llmAdapter))

		// Try to fix the diagram
		fixedDiagram, err := s.retryWithBackoff(ctx, "fix-file-diagram", func() (string, error) {
			return validationService.FixMermaidDiagramWithLLM(ctx, formattedDiagram, validationResult)
		})

		if err != nil {
			// If fixing failed, use the original but log the error
			fmt.Printf("Warning: Failed to fix diagram for %s: %v\n", filePath, err)
		} else {
			fmt.Printf("Successfully fixed diagram for %s\n", filePath)
			formattedDiagram = fixedDiagram
		}
	}

	return formattedDiagram, nil
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
	promptText := fmt.Sprintf("Please create a %s Mermaid diagram for the %s '%s' from this Go code:\n\n```go\n%s\n```\n\nProvide only the Mermaid diagram syntax without any explanation or markdown formatting.",
		diagramType, componentType, componentName, allCode)

	// Generate diagram using LLM
	diagramText, err := s.llmAdapter.GenerateCompletion(ctx, promptText)
	if err != nil {
		return "", fmt.Errorf("failed to generate diagram: %w", err)
	}

	formattedDiagram := mermaid.FormatOutput(diagramText)

	// Validate and fix the diagram if needed
	validationResult := mermaid.ValidateSyntax(formattedDiagram)

	if !validationResult.IsValid {
		fmt.Printf("Diagram for %s %s has syntax errors, attempting to fix...\n", componentType, componentName)

		// Create validation service for fixing diagrams
		validationService := NewValidationService(llm.NewClientAdapter(s.llmAdapter))

		// Try to fix the diagram
		fixedDiagram, err := s.retryWithBackoff(ctx, fmt.Sprintf("fix-%s-%s-diagram", componentType, componentName), func() (string, error) {
			return validationService.FixMermaidDiagramWithLLM(ctx, formattedDiagram, validationResult)
		})

		if err != nil {
			// If fixing failed, use the original but log the error
			fmt.Printf("Warning: Failed to fix diagram for %s %s: %v\n", componentType, componentName, err)
		} else {
			fmt.Printf("Successfully fixed diagram for %s %s\n", componentType, componentName)
			formattedDiagram = fixedDiagram
		}
	}

	return formattedDiagram, nil
}

// GenerateProjectDiagram generates project-wide Mermaid diagrams
func (s *diagramService) GenerateProjectDiagram(ctx context.Context, diagramType string) (string, error) {
	// Validate diagram type
	if !isValidProjectDiagramType(diagramType) {
		return "", fmt.Errorf("invalid project diagram type: %s (should be 'sequence', 'class', 'config', or 'adapters')", diagramType)
	}

	// Find all relevant files based on diagram type
	var files []string
	var err error

	// Special handling for class diagrams to process each component type separately
	if diagramType == "class" {
		return s.generateConcurrentClassDiagram(ctx)
	}

	switch diagramType {
	case "sequence":
		// Get service, repository, and adapter files for sequence diagram
		files, err = s.fileRepo.FindAllComponentFiles([]string{"service", "repository", "adapter"})
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
	var promptText string
	switch diagramType {
	case "sequence":
		promptText = fmt.Sprintf("Please create a sequence diagram showing the interactions between all components (services, repositories, adapters) in this Go project. Focus on the flow of calls between different components and how they interact:\n\n```go\n%s\n```\n\nProvide only the Mermaid diagram syntax without any explanation or markdown formatting.", allCode)
	case "config":
		promptText = fmt.Sprintf("Please create a diagram showing how configuration is structured and accessed throughout the application. Show config structs and how other components interact with them:\n\n```go\n%s\n```\n\nProvide only the Mermaid diagram syntax without any explanation or markdown formatting.", allCode)
	case "adapters":
		promptText = fmt.Sprintf("Please create a diagram showing all inbound and outbound communications in the application. Focus on adapter components and how they interact with external systems and internal components:\n\n```go\n%s\n```\n\nProvide only the Mermaid diagram syntax without any explanation or markdown formatting.", allCode)
	default:
		promptText = fmt.Sprintf("Please create a diagram showing the overall architecture of this Go project based on the following code:\n\n```go\n%s\n```\n\nProvide only the Mermaid diagram syntax without any explanation or markdown formatting.", allCode)
	}

	// Generate diagram using LLM
	diagramText, err := s.llmAdapter.GenerateCompletion(ctx, promptText)
	if err != nil {
		return "", fmt.Errorf("failed to generate diagram: %w", err)
	}

	formattedDiagram := mermaid.FormatOutput(diagramText)

	// Validate and fix the diagram if needed
	validationResult := mermaid.ValidateSyntax(formattedDiagram)

	if !validationResult.IsValid {
		fmt.Printf("Project diagram for type %s has syntax errors, attempting to fix...\n", diagramType)

		// Create validation service for fixing diagrams
		validationService := NewValidationService(llm.NewClientAdapter(s.llmAdapter))

		// Try to fix the diagram
		fixedDiagram, err := s.retryWithBackoff(ctx, fmt.Sprintf("fix-project-%s-diagram", diagramType), func() (string, error) {
			return validationService.FixMermaidDiagramWithLLM(ctx, formattedDiagram, validationResult)
		})

		if err != nil {
			// If fixing failed, use the original but log the error
			fmt.Printf("Warning: Failed to fix project diagram for type %s: %v\n", diagramType, err)
		} else {
			fmt.Printf("Successfully fixed project diagram for type %s\n", diagramType)
			formattedDiagram = fixedDiagram
		}
	}

	return formattedDiagram, nil
}

// retryWithBackoff attempts to call the provided function with exponential backoff
// for rate limit errors (429 status code)
func (s *diagramService) retryWithBackoff(ctx context.Context, operation string, fn func() (string, error)) (string, error) {
	maxRetries := 3
	baseDelay := 2 * time.Second

	var lastErr error
	for attempt := 0; attempt < maxRetries; attempt++ {
		result, err := fn()
		if err == nil {
			return result, nil
		}

		lastErr = err
		// Check if the error is a rate limit error (contains 429 status code)
		if strings.Contains(err.Error(), "429") {
			// Calculate backoff with jitter
			delay := baseDelay * time.Duration(1<<uint(attempt))
			jitter := time.Duration(rand.Int63n(int64(delay) / 2))
			waitTime := delay + jitter

			fmt.Printf("Rate limit hit for %s operation, retrying after %v (attempt %d/%d)\n",
				operation, waitTime, attempt+1, maxRetries)

			// Wait before next attempt
			select {
			case <-time.After(waitTime):
				continue
			case <-ctx.Done():
				return "", fmt.Errorf("context canceled while waiting to retry: %w", ctx.Err())
			}
		} else {
			// For non-rate-limit errors, don't retry
			return "", err
		}
	}

	return "", fmt.Errorf("failed after %d attempts: %w", maxRetries, lastErr)
}

// generateConcurrentClassDiagram generates class diagrams for each component type concurrently,
// validates/fixes them, and then combines them into a single diagram
func (s *diagramService) generateConcurrentClassDiagram(ctx context.Context) (string, error) {
	// Component types to process separately
	componentTypes := []string{"service", "repository", "adapter", "model", "config"}

	type diagramResult struct {
		componentType string
		diagram       string
		err           error
	}

	// Create a channel to receive results
	resultCh := make(chan diagramResult, len(componentTypes))

	// Create a validation service for fixing diagrams
	validationService := NewValidationService(llm.NewClientAdapter(s.llmAdapter))

	// Create a semaphore to limit concurrent LLM API calls
	// Using a buffered channel as a semaphore
	const maxConcurrentLLMCalls = 2 // Limit to 2 concurrent calls to avoid rate limiting
	sem := make(chan struct{}, maxConcurrentLLMCalls)

	// Process each component type concurrently
	for _, compType := range componentTypes {
		go func(componentType string) {
			// Find files for this component type
			files, err := s.fileRepo.FindAllComponentFiles([]string{componentType})
			if err != nil {
				resultCh <- diagramResult{componentType: componentType, err: fmt.Errorf("failed to find %s files: %w", componentType, err)}
				return
			}

			if len(files) == 0 {
				// No files for this component type, send empty result
				resultCh <- diagramResult{componentType: componentType, diagram: ""}
				return
			}

			// Read all files for this component type
			var codeContents []string
			for _, file := range files {
				content, err := s.fileRepo.ReadGoFile(file)
				if err != nil {
					resultCh <- diagramResult{componentType: componentType, err: fmt.Errorf("failed to read file %s: %w", file, err)}
					return
				}
				codeContents = append(codeContents, fmt.Sprintf("// File: %s\n%s", filepath.Base(file), content))
			}

			// Join code contents
			allCode := strings.Join(codeContents, "\n\n")

			// Acquire semaphore before making LLM API call
			sem <- struct{}{}
			fmt.Printf("Starting diagram generation for %s component\n", componentType)

			// Create prompt for this component type
			promptText := fmt.Sprintf("Please create a class diagram for the '%s' components in this Go project. Show their structs, interfaces, methods, and relationships:\n\n```go\n%s\n```\n\nProvide only the Mermaid diagram syntax without any explanation or markdown formatting.", componentType, allCode)

			// Use retryWithBackoff for LLM calls
			operation := fmt.Sprintf("generate-%s-diagram", componentType)
			diagramText, err := s.retryWithBackoff(ctx, operation, func() (string, error) {
				return s.llmAdapter.GenerateCompletion(ctx, promptText)
			})

			// Release semaphore after LLM API call
			<-sem

			if err != nil {
				resultCh <- diagramResult{componentType: componentType, err: fmt.Errorf("failed to generate %s diagram: %w", componentType, err)}
				return
			}

			// Validate and fix the diagram if needed
			formattedDiagram := mermaid.FormatOutput(diagramText)
			validationResult := mermaid.ValidateSyntax(formattedDiagram)

			if !validationResult.IsValid {
				fmt.Printf("Diagram for %s component has syntax errors, attempting to fix...\n", componentType)

				// Try to fix the diagram - acquire semaphore before LLM call
				sem <- struct{}{}

				// Use retryWithBackoff for fixing
				fixOperation := fmt.Sprintf("fix-%s-diagram", componentType)
				fixedDiagram, err := s.retryWithBackoff(ctx, fixOperation, func() (string, error) {
					return validationService.FixMermaidDiagramWithLLM(ctx, formattedDiagram, validationResult)
				})

				// Release semaphore after LLM API call
				<-sem

				if err != nil {
					// If fixing failed, use the original but log the error
					fmt.Printf("Warning: Failed to fix %s diagram: %v\n", componentType, err)
				} else {
					fmt.Printf("Successfully fixed %s diagram\n", componentType)
					formattedDiagram = fixedDiagram
				}
			}

			resultCh <- diagramResult{componentType: componentType, diagram: formattedDiagram}
		}(compType)
	}

	// Collect results
	componentDiagrams := make(map[string]string)
	var errors []string

	for i := 0; i < len(componentTypes); i++ {
		result := <-resultCh
		if result.err != nil {
			errors = append(errors, result.err.Error())
		} else if result.diagram != "" {
			componentDiagrams[result.componentType] = result.diagram
		}
	}

	// Check for errors
	if len(errors) > 0 {
		return "", fmt.Errorf("errors generating component diagrams: %s", strings.Join(errors, "; "))
	}

	// Combine diagrams into one
	if len(componentDiagrams) == 0 {
		return "", fmt.Errorf("no valid diagrams generated for any component type")
	}

	// Generate a combined diagram that relates the components
	var combinedDiagram strings.Builder
	combinedDiagram.WriteString("classDiagram\n")

	// Extract diagram content without headers and combine
	for compType, diagram := range componentDiagrams {
		// Remove the first line (diagram type declaration) and add the content
		lines := strings.Split(diagram, "\n")
		if len(lines) > 1 {
			// Add a comment to identify component section
			combinedDiagram.WriteString(fmt.Sprintf("  %% %s components\n", strings.ToUpper(compType)))
			combinedDiagram.WriteString(strings.Join(lines[1:], "\n"))
			combinedDiagram.WriteString("\n\n")
		}
	}

	// Generate relationships between components using LLM - acquire semaphore
	relationshipPrompt := "Based on the following component definitions, please generate only the relationships between different component types (service, repository, adapter, model, config). Return only Mermaid class diagram relationship syntax (e.g., 'ClassA --> ClassB: uses'):\n\n"

	for compType, diagram := range componentDiagrams {
		relationshipPrompt += fmt.Sprintf("// %s components\n%s\n\n", compType, diagram)
	}

	// Acquire semaphore for the final LLM call
	sem <- struct{}{}
	fmt.Println("Generating cross-component relationships...")

	// Use retryWithBackoff for relationship generation
	relationships, err := s.retryWithBackoff(ctx, "generate-relationships", func() (string, error) {
		return s.llmAdapter.GenerateCompletion(ctx, relationshipPrompt)
	})

	// Release semaphore
	<-sem

	if err == nil && relationships != "" {
		// Add relationships to the combined diagram
		combinedDiagram.WriteString("  % Cross-component relationships\n")

		// Extract only relationship lines
		relationshipLines := strings.Split(relationships, "\n")
		for _, line := range relationshipLines {
			line = strings.TrimSpace(line)
			if strings.Contains(line, "-->") || strings.Contains(line, "<--") ||
				strings.Contains(line, "--o") || strings.Contains(line, "--*") ||
				strings.Contains(line, "..>") || strings.Contains(line, "<..") {
				combinedDiagram.WriteString(line)
				combinedDiagram.WriteString("\n")
			}
		}
	} else if err != nil {
		fmt.Printf("Warning: Failed to generate cross-component relationships: %v\n", err)
	}

	return combinedDiagram.String(), nil
}

// mapDiagramType converts a string to a DiagramType
func (s *diagramService) mapDiagramType(dt string) mermaid.DiagramType {
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
