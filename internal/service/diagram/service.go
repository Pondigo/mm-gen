package diagram

import (
	"fmt"
	"path/filepath"

	"mm-go-agent/internal/adapter/renderer"
	fileOutputRepo "mm-go-agent/internal/repository/file"
)

// OutputService coordinates diagram processing and output
type OutputService struct {
	processor *Processor
	renderer  renderer.Renderer
	fileRepo  *fileOutputRepo.OutputRepository
}

// NewOutputService creates a new diagram output service
func NewOutputService(processor *Processor, renderer renderer.Renderer, fileRepo *fileOutputRepo.OutputRepository) *OutputService {
	return &OutputService{
		processor: processor,
		renderer:  renderer,
		fileRepo:  fileRepo,
	}
}

// GenerateAndSaveDiagram generates a diagram and saves it to the specified output directory
func (s *OutputService) GenerateAndSaveDiagram(diagramType, filePath, target, outDir string, svgFormat bool) error {
	// If outDir is not specified, just return
	if outDir == "" {
		return nil
	}

	// Create filename based on diagram type and target
	var filename string
	if filePath != "" {
		baseName := filepath.Base(filePath)
		filename = fmt.Sprintf("%s_%s", baseName, diagramType)
	} else if target == "map" {
		filename = fmt.Sprintf("project_%s", diagramType)
	} else {
		filename = fmt.Sprintf("component_%s", diagramType)
	}

	return s.SaveDiagram(filename, outDir, filePath, svgFormat)
}

// SaveDiagram saves a diagram to the specified output directory
func (s *OutputService) SaveDiagram(filename, outDir, content string, svgFormat bool) error {
	// Clean the diagram content
	cleanedContent := s.processor.CleanDiagramOutput(content)

	// If SVG format is requested, convert the diagram
	if svgFormat {
		svgContent, err := s.renderer.ConvertToSVG(cleanedContent)
		if err != nil {
			return fmt.Errorf("error converting to SVG: %v", err)
		}

		// Save both MMD and SVG files
		return s.fileRepo.SaveDiagramFiles(outDir, filename, cleanedContent, svgContent)
	}

	// Otherwise, just save the MMD file
	return s.fileRepo.SaveDiagramFile(outDir, filename, cleanedContent, "mmd")
}

// SaveSplitDiagram splits a project map diagram into component sections and saves them
func (s *OutputService) SaveSplitDiagram(diagram, diagramType, outDir string, svgFormat bool) error {
	// Extract component sections from the diagram
	componentSections := s.processor.ExtractComponentSections(diagram)

	// If no sections were found, save the whole diagram
	if len(componentSections) == 0 {
		cleanedDiagram := s.processor.CleanDiagramOutput(diagram)
		filename := fmt.Sprintf("project_%s", diagramType)

		if svgFormat {
			svgDiagram, err := s.renderer.ConvertToSVG(cleanedDiagram)
			if err != nil {
				return fmt.Errorf("error converting to SVG: %v", err)
			}
			return s.fileRepo.SaveDiagramFiles(outDir, filename, cleanedDiagram, svgDiagram)
		}

		return s.fileRepo.SaveDiagramFile(outDir, filename, cleanedDiagram, "mmd")
	}

	// Save each component section to its own file
	for component, content := range componentSections {
		cleanedContent := s.processor.CleanDiagramOutput(content)
		filename := fmt.Sprintf("%s_%s", component, diagramType)

		if svgFormat {
			svgContent, err := s.renderer.ConvertToSVG(cleanedContent)
			if err != nil {
				fmt.Printf("Warning: Error converting %s to SVG: %v\n", component, err)
				// Continue with other components
				continue
			}
			if err := s.fileRepo.SaveDiagramFiles(outDir, filename, cleanedContent, svgContent); err != nil {
				fmt.Printf("Warning: Error saving %s: %v\n", component, err)
			}
		} else {
			if err := s.fileRepo.SaveDiagramFile(outDir, filename, cleanedContent, "mmd"); err != nil {
				fmt.Printf("Warning: Error saving %s: %v\n", component, err)
			}
		}
	}

	// Also save a full combined diagram
	cleanedFullDiagram := s.processor.CleanDiagramOutput(diagram)
	fullFilename := fmt.Sprintf("project_%s_full", diagramType)

	if svgFormat {
		svgDiagram, err := s.renderer.ConvertToSVG(cleanedFullDiagram)
		if err != nil {
			fmt.Printf("Warning: Could not convert combined diagram to SVG: %v\n", err)
			// Still save the MMD version
			return s.fileRepo.SaveDiagramFile(outDir, fullFilename, cleanedFullDiagram, "mmd")
		}
		return s.fileRepo.SaveDiagramFiles(outDir, fullFilename, cleanedFullDiagram, svgDiagram)
	}

	return s.fileRepo.SaveDiagramFile(outDir, fullFilename, cleanedFullDiagram, "mmd")
}
