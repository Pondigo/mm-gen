package service

import (
	"context"
	"fmt"
	"mm-go-agent/pkg/llm"
	"mm-go-agent/pkg/mermaid"
	"os"
	"strconv"
)

// MaxFixRetries is the default number of retries for fixing Mermaid diagrams
const MaxFixRetries = 3

// ValidationService handles the validation of Mermaid diagrams
type ValidationService struct {
	llmClient  llm.Client
	maxRetries int
}

// NewValidationService creates a new validation service with the given LLM client
func NewValidationService(llmClient llm.Client) *ValidationService {
	// Get max retries from environment variable or use default
	maxRetries := MaxFixRetries
	if envRetries := os.Getenv("MERMAID_FIX_RETRIES"); envRetries != "" {
		if val, err := strconv.Atoi(envRetries); err == nil && val > 0 {
			maxRetries = val
		}
	}

	return &ValidationService{
		llmClient:  llmClient,
		maxRetries: maxRetries,
	}
}

// ValidateMermaidDiagram validates a Mermaid diagram and returns the validation result
func (s *ValidationService) ValidateMermaidDiagram(diagram string) (mermaid.ValidationResult, error) {
	result := mermaid.ValidateSyntax(diagram)
	return result, nil
}

// FormatValidationResult formats the validation result as a user-friendly string
func (s *ValidationService) FormatValidationResult(result mermaid.ValidationResult) string {
	return mermaid.FormatLinterOutput(result)
}

// FixMermaidDiagramWithLLM uses an LLM to fix a Mermaid diagram with syntax errors
// It will retry fixing the diagram up to the configured number of retries
func (s *ValidationService) FixMermaidDiagramWithLLM(ctx context.Context, diagram string, validationResult mermaid.ValidationResult) (string, error) {
	if validationResult.IsValid {
		return diagram, nil
	}

	// Current diagram to fix
	currentDiagram := diagram
	currentResult := validationResult
	fixAttempts := 0

	// Try to fix the diagram up to the maximum number of retries
	for fixAttempts < s.maxRetries {
		fixAttempts++

		// Create context for the LLM from the validation result
		validationContext := mermaid.ValidationResultAsContext(currentResult)

		// Add information about previous attempts if this is a retry
		retryInfo := ""
		if fixAttempts > 1 {
			retryInfo = fmt.Sprintf("\nThis is fix attempt %d/%d. Previous attempts still had errors.", fixAttempts, s.maxRetries)
		}

		prompt := fmt.Sprintf(`
You are a Mermaid diagram syntax expert. Fix the following Mermaid diagram that has syntax errors.
Here are the errors identified by the validator:%s

%s

Here is the diagram to fix:

%s

Please provide a complete, fixed version of the diagram that resolves all syntax errors.
Only respond with the corrected Mermaid diagram code, without any explanations or markdown formatting.
`, retryInfo, validationContext, currentDiagram)

		response, err := s.llmClient.GenerateText(ctx, prompt)
		if err != nil {
			return "", fmt.Errorf("error generating fixed diagram (attempt %d): %w", fixAttempts, err)
		}

		// Re-validate the fixed diagram
		fixedDiagram := response
		fixedResult := mermaid.ValidateSyntax(fixedDiagram)

		// If the diagram is now valid, return it
		if fixedResult.IsValid {
			return fixedDiagram, nil
		}

		// Update current diagram and result for the next attempt
		currentDiagram = fixedDiagram
		currentResult = fixedResult

		// If we've reached the maximum number of retries, return the best attempt with an error
		if fixAttempts >= s.maxRetries {
			return currentDiagram, fmt.Errorf("could not fix diagram after %d attempts, %d errors remain: %s",
				fixAttempts, len(currentResult.Errors), mermaid.FormatLinterOutput(currentResult))
		}
	}

	// This should not be reached due to the check inside the loop
	return currentDiagram, fmt.Errorf("diagram still has errors: %s", mermaid.FormatLinterOutput(currentResult))
}

// ExplainMermaidDiagramErrors uses an LLM to explain Mermaid syntax errors in a user-friendly way
func (s *ValidationService) ExplainMermaidDiagramErrors(ctx context.Context, validationResult mermaid.ValidationResult) (string, error) {
	if validationResult.IsValid {
		return "The Mermaid diagram is valid. No errors to explain.", nil
	}

	// Create context for the LLM from the validation result
	validationContext := mermaid.ValidationResultAsContext(validationResult)

	prompt := fmt.Sprintf(`
You are a Mermaid diagram syntax expert. Explain the following errors in a Mermaid diagram in a clear, user-friendly way.
Here are the errors identified by the validator:

%s

Here is the diagram with errors:

%s

Please provide a detailed explanation of what's wrong with the diagram and how to fix each error.
Use a friendly, educational tone as if you're teaching someone about Mermaid syntax.
`, validationContext, validationResult.Diagram)

	response, err := s.llmClient.GenerateText(ctx, prompt)
	if err != nil {
		return "", fmt.Errorf("error generating error explanation: %w", err)
	}

	return response, nil
}
