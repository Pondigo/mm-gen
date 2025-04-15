package repository

import "mm-go-agent/pkg/mermaid"

// LoggerRepository defines the interface for logging data for training purposes
type LoggerRepository interface {
	// LogFixAttempt logs a single attempt to fix a Mermaid diagram
	LogFixAttempt(originalDiagram string, validationResult mermaid.ValidationResult, fixedDiagram string, attempt int, isSuccessful bool) error

	// LogValidation logs a validation result
	LogValidation(diagram string, validationResult mermaid.ValidationResult) error

	// LogExplanation logs an explanation of errors
	LogExplanation(validationResult mermaid.ValidationResult, explanation string) error

	// LogPromptResponse logs the prompt and response for fine-tuning purposes
	LogPromptResponse(prompt string, response string, context string, entryType string) error
}
