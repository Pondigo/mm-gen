package repository

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// FileRepository defines the interface for file operations
type FileRepository interface {
	ReadGoFile(path string) (string, error)
	ValidateGoFile(path string) error
	FindComponentFiles(componentType, componentName string) ([]string, error)
	FindAllComponentFiles(componentTypes []string) ([]string, error)
}

// fileRepository implements FileRepository
type fileRepository struct{}

// NewFileRepository creates a new file repository
func NewFileRepository() FileRepository {
	return &fileRepository{}
}

// ValidateGoFile checks if a file exists and has a .go extension
func (r *fileRepository) ValidateGoFile(path string) error {
	if filepath.Ext(path) != ".go" {
		return fmt.Errorf("file must be a Go file (.go extension)")
	}

	_, err := os.Stat(path)
	if err != nil {
		return fmt.Errorf("error accessing file: %w", err)
	}

	return nil
}

// ReadGoFile reads the content of a go file
func (r *fileRepository) ReadGoFile(path string) (string, error) {
	if err := r.ValidateGoFile(path); err != nil {
		return "", err
	}

	content, err := os.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("error reading file: %w", err)
	}

	return string(content), nil
}

// FindComponentFiles finds all Go files for a specific component
func (r *fileRepository) FindComponentFiles(componentType, componentName string) ([]string, error) {
	var basePath string

	// Determine the base path for the component type
	switch componentType {
	case "service":
		basePath = "internal/services"
	case "repository":
		basePath = "internal/repositories"
	case "adapter":
		basePath = "internal/adapters"
	case "config":
		basePath = "internal/config"
	case "model":
		basePath = "internal/model"
	default:
		return nil, fmt.Errorf("unsupported component type: %s", componentType)
	}

	// Create search pattern for the component name
	namePattern := strings.ToLower(componentName)

	var files []string
	err := filepath.Walk(basePath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			// Special handling for adapters which are typically organized in subdirectories
			if componentType == "adapter" && strings.HasSuffix(path, namePattern) {
				return nil // Continue descending into matching adapter directories
			}

			// Skip non-matching directories for other component types
			if filepath.Dir(path) != basePath && path != basePath {
				return filepath.SkipDir
			}
			return nil
		}

		// Check if the file matches the component name
		fileName := strings.ToLower(info.Name())
		if filepath.Ext(fileName) == ".go" &&
			(strings.Contains(fileName, namePattern) ||
				strings.Contains(fileName, strings.ToLower(componentType))) {
			files = append(files, path)
		}

		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("error walking directory %s: %w", basePath, err)
	}

	return files, nil
}

// FindAllComponentFiles finds all Go files for the specified component types
func (r *fileRepository) FindAllComponentFiles(componentTypes []string) ([]string, error) {
	var allFiles []string

	for _, componentType := range componentTypes {
		var basePath string

		// Determine the base path for the component type
		switch componentType {
		case "service":
			basePath = "internal/services"
		case "repository":
			basePath = "internal/repositories"
		case "adapter":
			basePath = "internal/adapters"
		case "config":
			basePath = "internal/config"
		case "model":
			basePath = "internal/models"
		default:
			continue // Skip unsupported component types
		}

		// Check if the directory exists
		if _, err := os.Stat(basePath); os.IsNotExist(err) {
			continue // Skip non-existent directories
		}

		// Find all Go files in the directory
		err := filepath.Walk(basePath, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}

			if !info.IsDir() && filepath.Ext(path) == ".go" {
				allFiles = append(allFiles, path)
			}

			return nil
		})

		if err != nil {
			return nil, fmt.Errorf("error walking directory %s: %w", basePath, err)
		}
	}

	return allFiles, nil
}
