package pipeline

import (
	"context"
	"crypto/sha256"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

// FileScanner handles file discovery and filtering
type FileScanner struct {
	root      string
	pattern   string
	exclude   []string
	recursive bool
}

// NewFileScanner creates a new file scanner
func NewFileScanner(root, pattern string, exclude []string, recursive bool) *FileScanner {
	return &FileScanner{
		root:      root,
		pattern:   pattern,
		exclude:   exclude,
		recursive: recursive,
	}
}

// Scan discovers files matching the criteria
func (s *FileScanner) Scan(ctx context.Context) ([]FileInfo, error) {
	var files []FileInfo

	err := filepath.WalkDir(s.root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		// Check context cancellation
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		// Skip directories
		if d.IsDir() {
			// Check if we should skip this directory
			if !s.recursive && path != s.root {
				return filepath.SkipDir
			}
			return nil
		}

		// Check exclusion patterns
		for _, pattern := range s.exclude {
			if matched, _ := filepath.Match(pattern, filepath.Base(path)); matched {
				return nil
			}
			// Also check against relative path for patterns like "drafts/**"
			relPath, _ := filepath.Rel(s.root, path)
			if matched, _ := filepath.Match(pattern, relPath); matched {
				return nil
			}
		}

		// Check glob pattern if specified
		if s.pattern != "" {
			// Support both simple patterns and ** patterns
			if strings.Contains(s.pattern, "**") {
				// Handle ** glob pattern (match any directory depth)
				pattern := strings.ReplaceAll(s.pattern, "**", "*")
				relPath, _ := filepath.Rel(s.root, path)
				if matched, _ := filepath.Match(pattern, relPath); !matched {
					return nil
				}
			} else {
				// Simple pattern matching on filename
				if matched, _ := filepath.Match(s.pattern, filepath.Base(path)); !matched {
					return nil
				}
			}
		}

		// Get file info
		info, err := d.Info()
		if err != nil {
			return nil // Skip files we can't stat
		}

		// Detect file type
		fileType := detectFileType(path)

		// Calculate hash (for deduplication)
		hash, _ := calculateFileHash(path)

		files = append(files, FileInfo{
			Path:    path,
			Size:    info.Size(),
			ModTime: info.ModTime(),
			Type:    fileType,
			Hash:    hash,
		})

		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to scan files: %w", err)
	}

	return files, nil
}

// detectFileType determines the file type based on extension
func detectFileType(path string) FileType {
	ext := strings.ToLower(filepath.Ext(path))
	switch ext {
	case ".pdf":
		return FileTypePDF
	case ".txt":
		return FileTypeTXT
	case ".md", ".markdown":
		return FileTypeMD
	case ".json":
		return FileTypeJSON
	case ".yaml", ".yml":
		return FileTypeYAML
	case ".png", ".jpg", ".jpeg", ".gif", ".bmp":
		return FileTypeImage
	default:
		return FileTypeUnknown
	}
}

// calculateFileHash calculates SHA256 hash of file
func calculateFileHash(path string) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer f.Close()

	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return "", err
	}

	return fmt.Sprintf("%x", h.Sum(nil)), nil
}
