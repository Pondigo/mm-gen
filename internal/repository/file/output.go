package file

import (
	"fmt"
	"os"
	"path/filepath"
)

// OutputRepository handles file output operations
type OutputRepository struct{}

// NewOutputRepository creates a new file output repository
func NewOutputRepository() *OutputRepository {
	return &OutputRepository{}
}

// SaveDiagramFiles saves both MMD and SVG diagram files
func (r *OutputRepository) SaveDiagramFiles(outDir, filename, mmdContent, svgContent string) error {
	// Create output directory if it doesn't exist
	if err := os.MkdirAll(outDir, 0755); err != nil {
		return fmt.Errorf("error creating output directory: %v", err)
	}

	// Save the MMD file
	mmdOutputPath := filepath.Join(outDir, filename+".mmd")
	if err := os.WriteFile(mmdOutputPath, []byte(mmdContent), 0644); err != nil {
		return fmt.Errorf("error writing MMD file: %v", err)
	}
	fmt.Printf("Original diagram saved to %s\n", mmdOutputPath)

	// If SVG content is provided, save it as well
	if svgContent != "" {
		svgOutputPath := filepath.Join(outDir, filename+".svg")
		if err := os.WriteFile(svgOutputPath, []byte(svgContent), 0644); err != nil {
			return fmt.Errorf("error writing SVG file: %v", err)
		}
		fmt.Printf("SVG diagram saved to %s\n", svgOutputPath)
	}

	return nil
}

// SaveDiagramFile saves a single diagram file with the specified content and extension
func (r *OutputRepository) SaveDiagramFile(outDir, filename, content, extension string) error {
	// Create output directory if it doesn't exist
	if err := os.MkdirAll(outDir, 0755); err != nil {
		return fmt.Errorf("error creating output directory: %v", err)
	}

	// Create the full output path
	outputPath := filepath.Join(outDir, filename+"."+extension)

	// Write the file
	if err := os.WriteFile(outputPath, []byte(content), 0644); err != nil {
		return fmt.Errorf("error writing file: %v", err)
	}

	fmt.Printf("Diagram saved to %s\n", outputPath)
	return nil
}
