package main

import (
	"context"
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"mm-go-agent/internal/adapter/llm"
	"mm-go-agent/internal/repository"
	"mm-go-agent/internal/service"
)

func main() {
	var rootCmd = &cobra.Command{
		Use:   "mm-gen [diagram-type] [file]",
		Short: "Generate Mermaid diagrams from Go code",
		Args:  cobra.ExactArgs(2),
		Run: func(cmd *cobra.Command, args []string) {
			diagramType := args[0]
			filePath := args[1]

			// Initialize dependencies
			fileRepo := repository.NewFileRepository()
			llmAdapter, err := llm.NewClaudeAdapter("")
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error initializing LLM: %v\n", err)
				os.Exit(1)
			}

			// Initialize service
			diagramService := service.NewDiagramService(fileRepo, llmAdapter)

			// Generate diagram
			ctx := context.Background()
			diagram, err := diagramService.GenerateDiagram(ctx, filePath, diagramType)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error: %v\n", err)
				os.Exit(1)
			}

			// Print diagram
			fmt.Println(diagram)
		},
	}

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}
}
