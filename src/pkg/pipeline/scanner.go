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

	"github.com/bmatcuk/doublestar/v4"
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

		// Get relative path for pattern matching
		relPath, err := filepath.Rel(s.root, path)
		if err != nil {
			relPath = filepath.Base(path)
		}

		// Check exclusion patterns
		for _, pattern := range s.exclude {
			// Try doublestar match first (supports ** patterns)
			if matched, _ := doublestar.Match(pattern, relPath); matched {
				return nil
			}
			// Also try matching against basename for simple patterns
			if matched, _ := doublestar.Match(pattern, filepath.Base(path)); matched {
				return nil
			}
		}

		// Check glob pattern if specified
		if s.pattern != "" {
			matched := false

			// Try doublestar match first (supports ** and directory separators)
			if m, _ := doublestar.Match(s.pattern, relPath); m {
				matched = true
			}

			// Also try matching against basename for simple patterns like "*.pdf"
			if !matched {
				if m, _ := doublestar.Match(s.pattern, filepath.Base(path)); m {
					matched = true
				}
			}

			if !matched {
				return nil
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
