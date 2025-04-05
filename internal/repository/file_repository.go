package repository

import (
	"fmt"
	"os"
	"path/filepath"
)

// FileRepository defines the interface for file operations
type FileRepository interface {
	ReadGoFile(path string) (string, error)
	ValidateGoFile(path string) error
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
