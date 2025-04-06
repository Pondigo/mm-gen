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
		Use:   "mm-gen",
		Short: "Generate Mermaid diagrams from Go code",
	}

	// Command for generating diagram for a single file
	var fileCmd = &cobra.Command{
		Use:   "file [diagram-type] [file]",
		Short: "Generate Mermaid diagram from a single Go file",
		Args:  cobra.ExactArgs(2),
		Run: func(cmd *cobra.Command, args []string) {
			diagramType := args[0]
			filePath := args[1]

			generateAndPrintDiagram(diagramType, filePath, "")
		},
	}

	// Command for generating diagram for a component
	var componentCmd = &cobra.Command{
		Use:   "component [diagram-type] [component-type] [component-name]",
		Short: "Generate Mermaid diagram for a specific component (service, repository, adapter, etc.)",
		Long:  "Generate Mermaid diagram for a specific component. Component types: service, repository, adapter, config, model",
		Args:  cobra.ExactArgs(3),
		Run: func(cmd *cobra.Command, args []string) {
			diagramType := args[0]
			componentType := args[1]
			componentName := args[2]

			generateAndPrintDiagram(diagramType, "", fmt.Sprintf("%s:%s", componentType, componentName))
		},
	}

	// Command for mapping the project
	var mapCmd = &cobra.Command{
		Use:   "map [diagram-type]",
		Short: "Generate project-wide Mermaid diagrams",
		Long:  "Generate project-wide Mermaid diagrams. Diagram types: sequence (component interactions), class (all components), config (config interactions), adapters (inbound/outbound communications)",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			diagramType := args[0]

			generateAndPrintDiagram(diagramType, "", "map")
		},
	}

	rootCmd.AddCommand(fileCmd, componentCmd, mapCmd)

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}
}

func generateAndPrintDiagram(diagramType, filePath, target string) {
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
	var diagram string

	if filePath != "" {
		// Generate diagram for a single file
		diagram, err = diagramService.GenerateDiagram(ctx, filePath, diagramType)
	} else if target == "map" {
		// Generate project-wide diagram
		diagram, err = diagramService.GenerateProjectDiagram(ctx, diagramType)
	} else {
		// Generate component diagram
		diagram, err = diagramService.GenerateComponentDiagram(ctx, target, diagramType)
	}

	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	// Print diagram
	fmt.Println(diagram)
}
