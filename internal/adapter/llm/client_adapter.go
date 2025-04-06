package llm

import (
	"context"
	pkgllm "mm-go-agent/pkg/llm"
)

// ClientAdapter is an adapter that implements the pkg/llm.Client interface using the internal LLMAdapter
type ClientAdapter struct {
	adapter LLMAdapter
}

// NewClientAdapter creates a new client adapter
func NewClientAdapter(adapter LLMAdapter) pkgllm.Client {
	return &ClientAdapter{
		adapter: adapter,
	}
}

// GenerateText implements the Client.GenerateText method using the underlying LLMAdapter
func (a *ClientAdapter) GenerateText(ctx context.Context, prompt string) (string, error) {
	return a.adapter.GenerateCompletion(ctx, prompt)
}
