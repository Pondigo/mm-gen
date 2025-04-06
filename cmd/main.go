package main

import (
	"context"
	"fmt"
	"io"
	"os"
	"strconv"

	"github.com/spf13/cobra"

	"mm-go-agent/internal/adapter/llm"
	"mm-go-agent/internal/repository"
	"mm-go-agent/internal/service"
	pkgllm "mm-go-agent/pkg/llm"
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

	// Command for validating Mermaid diagram syntax
	var validateCmd = &cobra.Command{
		Use:   "validate [file]",
		Short: "Validate Mermaid diagram syntax",
		Long:  "Validate Mermaid diagram syntax and report any errors. Can read from a file or stdin.",
		Args:  cobra.MaximumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			var diagram string

			// Get diagram content from file or stdin
			if len(args) > 0 {
				// Read from file
				content, err := os.ReadFile(args[0])
				if err != nil {
					fmt.Fprintf(os.Stderr, "Error reading file: %v\n", err)
					os.Exit(1)
				}
				diagram = string(content)
			} else {
				// Read from stdin
				content, err := io.ReadAll(os.Stdin)
				if err != nil {
					fmt.Fprintf(os.Stderr, "Error reading from stdin: %v\n", err)
					os.Exit(1)
				}
				diagram = string(content)
			}

			validateDiagram(diagram, cmd)
		},
	}

	// Add the explain flag to provide a more detailed explanation of errors
	explainFlag := false
	validateCmd.Flags().BoolVarP(&explainFlag, "explain", "e", false, "Provide a detailed explanation of syntax errors")

	// Add the fix flag to attempt to fix the diagram
	fixFlag := false
	validateCmd.Flags().BoolVarP(&fixFlag, "fix", "f", false, "Attempt to fix syntax errors in the diagram")

	// Add verbose flag to show more information about the fixing process
	verboseFlag := false
	validateCmd.Flags().BoolVarP(&verboseFlag, "verbose", "v", false, "Show verbose output for the fixing process")

	// Add retries flag to set the maximum number of retries
	retriesFlag := 0
	validateCmd.Flags().IntVarP(&retriesFlag, "retries", "r", 0, "Maximum number of retries for fixing (0 = use default/env var)")

	rootCmd.AddCommand(fileCmd, componentCmd, mapCmd, validateCmd)

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

// validateDiagram validates a Mermaid diagram and outputs the result
func validateDiagram(diagram string, cmd *cobra.Command) {
	// Get flags
	explainFlag, _ := cmd.Flags().GetBool("explain")
	fixFlag, _ := cmd.Flags().GetBool("fix")
	verboseFlag, _ := cmd.Flags().GetBool("verbose")
	retriesFlag, _ := cmd.Flags().GetInt("retries")

	// If retries flag is set, use it to override the environment variable
	if retriesFlag > 0 {
		os.Setenv("MERMAID_FIX_RETRIES", strconv.Itoa(retriesFlag))
		if verboseFlag {
			fmt.Printf("Setting maximum retries to %d\n", retriesFlag)
		}
	}

	// Initialize Claude adapter if we need to fix or explain errors
	var llmClient pkgllm.Client
	if explainFlag || fixFlag {
		claudeAdapter, err := llm.NewClaudeAdapter("")
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error initializing Claude LLM: %v\n", err)
			os.Exit(1)
		}

		// Create a client adapter that bridges LLMAdapter and Client interfaces
		llmClient = llm.NewClientAdapter(claudeAdapter)
	}

	// Create validation service
	validationService := service.NewValidationService(llmClient)

	// Validate the diagram
	validationResult, validationErr := validationService.ValidateMermaidDiagram(diagram)
	if validationErr != nil {
		fmt.Fprintf(os.Stderr, "Error validating diagram: %v\n", validationErr)
		os.Exit(1)
	}

	// Format and print the validation result
	fmt.Println(validationService.FormatValidationResult(validationResult))

	ctx := context.Background()

	// If diagram is invalid and fix flag is set, try to fix it
	if !validationResult.IsValid && fixFlag {
		if verboseFlag {
			fmt.Println("\nAttempting to fix diagram with up to",
				os.Getenv("MERMAID_FIX_RETRIES"), "retries...")
		}

		fixedDiagram, fixErr := validationService.FixMermaidDiagramWithLLM(ctx, diagram, validationResult)

		if fixErr != nil {
			fmt.Fprintf(os.Stderr, "Error fixing diagram: %v\n", fixErr)

			// Check if we have a partially fixed diagram to show
			if fixedDiagram != "" && fixedDiagram != diagram {
				fmt.Println("\nPartially fixed diagram (still has errors):")
				fmt.Println(fixedDiagram)
			}
		} else {
			if verboseFlag {
				fmt.Println("Successfully fixed diagram!")
			}
			fmt.Println("\nFixed diagram:")
			fmt.Println(fixedDiagram)

			// Re-validate to show the fixed version is valid
			if verboseFlag {
				fixedResult, _ := validationService.ValidateMermaidDiagram(fixedDiagram)
				if fixedResult.IsValid {
					fmt.Println("Validation confirmed: The fixed diagram is valid.")
				}
			}
		}
	}

	// If diagram is invalid and explain flag is set, explain the errors
	if !validationResult.IsValid && explainFlag {
		explanation, explainErr := validationService.ExplainMermaidDiagramErrors(ctx, validationResult)
		if explainErr != nil {
			fmt.Fprintf(os.Stderr, "Error explaining diagram errors: %v\n", explainErr)
		} else {
			fmt.Println("\nExplanation of errors:")
			fmt.Println(explanation)
		}
	}

	// Exit with non-zero code if the diagram is invalid
	if !validationResult.IsValid {
		os.Exit(1)
	}
}
