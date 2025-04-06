package mermaid

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// SyntaxError represents an error in the Mermaid diagram syntax
type SyntaxError struct {
	Line    int    `json:"line"`
	Message string `json:"message"`
	Text    string `json:"text,omitempty"`
}

// ValidationResult contains the result of validating a Mermaid diagram
type ValidationResult struct {
	IsValid  bool          `json:"isValid"`
	Errors   []SyntaxError `json:"errors,omitempty"`
	Diagram  string        `json:"diagram"`
	ErrorMsg string        `json:"errorMsg,omitempty"`
}

// ValidateSyntax checks if the given Mermaid diagram has valid syntax
// It returns a ValidationResult containing validation information
func ValidateSyntax(diagram string) ValidationResult {
	// Remove the mermaid markdown formatting if present
	cleanDiagram := strings.TrimPrefix(diagram, "```mermaid\n")
	cleanDiagram = strings.TrimSuffix(cleanDiagram, "\n```")

	// Create temporary file to store the diagram
	tempDir, err := os.MkdirTemp("", "mermaid-validation")
	if err != nil {
		return ValidationResult{
			IsValid:  false,
			Diagram:  diagram,
			ErrorMsg: fmt.Sprintf("Failed to create temporary directory: %v", err),
		}
	}
	defer os.RemoveAll(tempDir)

	tempFile := filepath.Join(tempDir, "diagram.mmd")
	if err := os.WriteFile(tempFile, []byte(cleanDiagram), 0644); err != nil {
		return ValidationResult{
			IsValid:  false,
			Diagram:  diagram,
			ErrorMsg: fmt.Sprintf("Failed to write temporary file: %v", err),
		}
	}

	// Check if mmdc (Mermaid CLI) is installed
	if _, err := exec.LookPath("mmdc"); err != nil {
		// If not installed, fall back to basic syntax checking
		return basicSyntaxCheck(cleanDiagram)
	}

	// Use mmdc to validate the syntax
	cmd := exec.Command("mmdc", "-i", tempFile, "-o", filepath.Join(tempDir, "output.svg"))
	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		// Parse the error output to extract syntax errors
		errorOutput := stderr.String()
		return parseErrorOutput(errorOutput, cleanDiagram)
	}

	return ValidationResult{
		IsValid: true,
		Diagram: diagram,
	}
}

// basicSyntaxCheck performs a basic syntax check without external tools
func basicSyntaxCheck(diagram string) ValidationResult {
	result := ValidationResult{
		IsValid: true,
		Diagram: diagram,
	}

	lines := strings.Split(diagram, "\n")

	// Check for essential Mermaid elements
	if len(lines) == 0 {
		result.IsValid = false
		result.Errors = append(result.Errors, SyntaxError{
			Line:    0,
			Message: "Empty diagram",
		})
		return result
	}

	// Check if the first line declares a valid diagram type
	firstLine := strings.TrimSpace(lines[0])
	validTypes := []string{"graph ", "flowchart ", "sequenceDiagram", "classDiagram", "stateDiagram",
		"erDiagram", "journey", "gantt", "pie", "requirementDiagram", "gitGraph"}

	isValidType := false
	for _, t := range validTypes {
		if strings.HasPrefix(firstLine, t) {
			isValidType = true
			break
		}
	}

	if !isValidType {
		result.IsValid = false
		result.Errors = append(result.Errors, SyntaxError{
			Line:    1,
			Message: "Invalid or missing diagram type declaration",
			Text:    firstLine,
		})
	}

	// Check for common syntax errors
	for i, line := range lines {
		trimmedLine := strings.TrimSpace(line)

		// Skip comments and empty lines
		if trimmedLine == "" || strings.HasPrefix(trimmedLine, "%%") {
			continue
		}

		// Check for unclosed quotes
		if strings.Count(trimmedLine, "\"")%2 != 0 {
			result.IsValid = false
			result.Errors = append(result.Errors, SyntaxError{
				Line:    i + 1,
				Message: "Unclosed quotes",
				Text:    trimmedLine,
			})
		}

		// Check for mismatched brackets
		if strings.Count(trimmedLine, "[") != strings.Count(trimmedLine, "]") {
			result.IsValid = false
			result.Errors = append(result.Errors, SyntaxError{
				Line:    i + 1,
				Message: "Mismatched square brackets",
				Text:    trimmedLine,
			})
		}

		if strings.Count(trimmedLine, "(") != strings.Count(trimmedLine, ")") {
			result.IsValid = false
			result.Errors = append(result.Errors, SyntaxError{
				Line:    i + 1,
				Message: "Mismatched parentheses",
				Text:    trimmedLine,
			})
		}

		if strings.Count(trimmedLine, "{") != strings.Count(trimmedLine, "}") {
			result.IsValid = false
			result.Errors = append(result.Errors, SyntaxError{
				Line:    i + 1,
				Message: "Mismatched curly braces",
				Text:    trimmedLine,
			})
		}
	}

	return result
}

// parseErrorOutput parses the error output from mmdc to extract syntax errors
func parseErrorOutput(errorOutput, diagram string) ValidationResult {
	result := ValidationResult{
		IsValid:  false,
		Diagram:  diagram,
		ErrorMsg: errorOutput,
	}

	// Extract line numbers and error messages
	// This is a simplified parser and might need adjustments based on actual mmdc output format
	lines := strings.Split(errorOutput, "\n")
	diagramLines := strings.Split(diagram, "\n")

	for _, line := range lines {
		// Look for patterns like "Error: Parse error on line X:"
		if strings.Contains(line, "Error") && strings.Contains(line, "line") {
			parts := strings.Split(line, "line")
			if len(parts) > 1 {
				lineNumStr := strings.TrimSpace(strings.Split(parts[1], ":")[0])
				var lineNum int
				fmt.Sscanf(lineNumStr, "%d", &lineNum)

				errorMsg := strings.TrimSpace(strings.Join(parts[1:], ""))

				// Get the text from the line with the error
				var lineText string
				if lineNum > 0 && lineNum <= len(diagramLines) {
					lineText = diagramLines[lineNum-1]
				}

				result.Errors = append(result.Errors, SyntaxError{
					Line:    lineNum,
					Message: errorMsg,
					Text:    lineText,
				})
			}
		}
	}

	// If no specific errors were parsed but there was an error output,
	// add a generic error
	if len(result.Errors) == 0 && errorOutput != "" {
		result.Errors = append(result.Errors, SyntaxError{
			Line:    0,
			Message: "Syntax error in diagram",
			Text:    errorOutput,
		})
	}

	return result
}

// FormatLinterOutput formats the validation result as a string
func FormatLinterOutput(result ValidationResult) string {
	if result.IsValid {
		return "Mermaid diagram syntax is valid."
	}

	var output strings.Builder
	output.WriteString("Mermaid diagram syntax validation failed:\n\n")

	for _, err := range result.Errors {
		if err.Line > 0 {
			output.WriteString(fmt.Sprintf("Line %d: %s\n", err.Line, err.Message))
			if err.Text != "" {
				output.WriteString(fmt.Sprintf("  %s\n", err.Text))
			}
		} else {
			output.WriteString(fmt.Sprintf("%s\n", err.Message))
		}
	}

	if result.ErrorMsg != "" && len(result.Errors) == 0 {
		output.WriteString(fmt.Sprintf("\nError details:\n%s\n", result.ErrorMsg))
	}

	return output.String()
}

// ValidationResultAsContext formats the validation result as context for an LLM
func ValidationResultAsContext(result ValidationResult) string {
	jsonBytes, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return fmt.Sprintf("Error creating validation context: %v", err)
	}
	return string(jsonBytes)
}
