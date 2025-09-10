package file

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"mm-go-agent/pkg/mermaid"
)

// LogEntry represents a single log entry for training data
type LogEntry struct {
	Timestamp        time.Time                 `json:"timestamp"`
	Type             string                    `json:"type"`
	OriginalDiagram  string                    `json:"original_diagram,omitempty"`
	ValidationResult *mermaid.ValidationResult `json:"validation_result,omitempty"`
	FixedDiagram     string                    `json:"fixed_diagram,omitempty"`
	Attempt          int                       `json:"attempt,omitempty"`
	IsSuccessful     bool                      `json:"is_successful,omitempty"`
	Explanation      string                    `json:"explanation,omitempty"`
	Prompt           string                    `json:"prompt,omitempty"`
	Response         string                    `json:"response,omitempty"`
	SessionID        string                    `json:"session_id,omitempty"`
}

// LoggerRepository implements the repository.LoggerRepository interface
type LoggerRepository struct {
	logDir    string
	sessionID string
}

// NewLoggerRepository creates a new logger repository
func NewLoggerRepository(logDir string) (*LoggerRepository, error) {
	// Default to ~/mm-gen-logs directory if none specified
	if logDir == "" {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return nil, fmt.Errorf("failed to get user home directory: %w", err)
		}
		logDir = filepath.Join(homeDir, "mm-gen-logs")
	}

	// Create logs directory if it doesn't exist
	if err := os.MkdirAll(logDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create log directory: %w", err)
	}

	// Generate a unique session ID
	sessionID := fmt.Sprintf("%s-%d", time.Now().Format("20060102-150405"), time.Now().UnixNano()%1000)

	// Create a dedicated directory for this session
	sessionDir := filepath.Join(logDir, sessionID)
	if err := os.MkdirAll(sessionDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create session directory: %w", err)
	}

	fmt.Printf("Logging data to: %s\n", sessionDir)

	return &LoggerRepository{
		logDir:    sessionDir,
		sessionID: sessionID,
	}, nil
}

// LogFixAttempt logs a single attempt to fix a Mermaid diagram
func (r *LoggerRepository) LogFixAttempt(originalDiagram string, validationResult mermaid.ValidationResult, fixedDiagram string, attempt int, isSuccessful bool) error {
	entry := LogEntry{
		Timestamp:        time.Now(),
		Type:             "fix_attempt",
		OriginalDiagram:  originalDiagram,
		ValidationResult: &validationResult,
		FixedDiagram:     fixedDiagram,
		Attempt:          attempt,
		IsSuccessful:     isSuccessful,
		SessionID:        r.sessionID,
	}

	return r.saveLogEntry(entry)
}

// LogValidation logs a validation result
func (r *LoggerRepository) LogValidation(diagram string, validationResult mermaid.ValidationResult) error {
	entry := LogEntry{
		Timestamp:        time.Now(),
		Type:             "validation",
		OriginalDiagram:  diagram,
		ValidationResult: &validationResult,
		SessionID:        r.sessionID,
	}

	return r.saveLogEntry(entry)
}

// LogExplanation logs an explanation of errors
func (r *LoggerRepository) LogExplanation(validationResult mermaid.ValidationResult, explanation string) error {
	entry := LogEntry{
		Timestamp:        time.Now(),
		Type:             "explanation",
		OriginalDiagram:  validationResult.Diagram,
		ValidationResult: &validationResult,
		Explanation:      explanation,
		SessionID:        r.sessionID,
	}

	return r.saveLogEntry(entry)
}

// LogPromptResponse logs the prompt and response for fine-tuning purposes
func (r *LoggerRepository) LogPromptResponse(prompt string, response string, context string, entryType string) error {
	entry := LogEntry{
		Timestamp:       time.Now(),
		Type:            entryType,
		OriginalDiagram: context,
		Prompt:          prompt,
		Response:        response,
		SessionID:       r.sessionID,
	}

	return r.saveLogEntry(entry)
}

// saveLogEntry saves a log entry to a JSON file
func (r *LoggerRepository) saveLogEntry(entry LogEntry) error {
	// Create filename based on timestamp and type
	filename := fmt.Sprintf("%s_%s_%d.json",
		entry.Timestamp.Format("20060102_150405"),
		entry.Type,
		entry.Timestamp.UnixNano())

	// Create full path
	filePath := filepath.Join(r.logDir, filename)

	// Marshal entry to JSON
	data, err := json.MarshalIndent(entry, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal log entry: %w", err)
	}

	// Write to file
	if err := os.WriteFile(filePath, data, 0644); err != nil {
		return fmt.Errorf("failed to write log file: %w", err)
	}

	return nil
}
