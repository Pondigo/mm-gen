package llm

import (
	"context"
)

// Client is an interface for text generation using LLMs
type Client interface {
	// GenerateText generates text from a prompt
	GenerateText(ctx context.Context, prompt string) (string, error)
}
