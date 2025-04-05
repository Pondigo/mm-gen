package llm

import (
	"context"
	"fmt"

	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/llms/anthropic"
)

// LLMAdapter defines the interface for LLM interactions
type LLMAdapter interface {
	GenerateCompletion(ctx context.Context, prompt string) (string, error)
}

// claudeAdapter implements LLMAdapter for Claude
type claudeAdapter struct {
	model string
	llm   llms.LLM
}

// NewClaudeAdapter creates a new Claude adapter
func NewClaudeAdapter(model string) (LLMAdapter, error) {
	if model == "" {
		model = "claude-3-7-sonnet-20250219" // Default to Claude 3.7 Sonnet
	}

	llm, err := anthropic.New(
		anthropic.WithModel(model),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize Claude: %w", err)
	}

	return &claudeAdapter{
		model: model,
		llm:   llm,
	}, nil
}

// GenerateCompletion generates a completion from Claude
func (a *claudeAdapter) GenerateCompletion(ctx context.Context, prompt string) (string, error) {
	completion, err := llms.GenerateFromSinglePrompt(ctx, a.llm, prompt)
	if err != nil {
		return "", fmt.Errorf("error generating completion: %w", err)
	}

	return completion, nil
}
