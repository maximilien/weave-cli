package image

import (
	"fmt"
	"path/filepath"
	"strings"
	"time"
)

// GenerateTimestampedPath generates a timestamped relative path for an image
// Format: YYYYMMDD_HHMMSS_microseconds_originalfilename.ext
// Example: 20251015_143052_123456_yorkshire-terrier.jpg
func GenerateTimestampedPath(originalPath string) string {
	// Get the filename from the original path
	filename := filepath.Base(originalPath)

	// Get current time
	now := time.Now()

	// Format: YYYYMMDD_HHMMSS_microseconds
	timestamp := fmt.Sprintf("%04d%02d%02d_%02d%02d%02d_%06d",
		now.Year(),
		now.Month(),
		now.Day(),
		now.Hour(),
		now.Minute(),
		now.Second(),
		now.Nanosecond()/1000, // Convert to microseconds
	)

	// Clean filename: lowercase, replace spaces with hyphens
	cleanFilename := strings.ToLower(filename)
	cleanFilename = strings.ReplaceAll(cleanFilename, " ", "-")

	// Combine timestamp with filename
	timestampedFilename := fmt.Sprintf("%s_%s", timestamp, cleanFilename)

	// Return relative path with images/ prefix
	return filepath.Join("images", timestampedFilename)
}

// GenerateTimestampedPathWithPrefix generates a timestamped path with custom prefix
func GenerateTimestampedPathWithPrefix(originalPath, prefix string) string {
	filename := filepath.Base(originalPath)
	now := time.Now()

	timestamp := fmt.Sprintf("%04d%02d%02d_%02d%02d%02d_%06d",
		now.Year(),
		now.Month(),
		now.Day(),
		now.Hour(),
		now.Minute(),
		now.Second(),
		now.Nanosecond()/1000,
	)

	cleanFilename := strings.ToLower(filename)
	cleanFilename = strings.ReplaceAll(cleanFilename, " ", "-")

	timestampedFilename := fmt.Sprintf("%s_%s", timestamp, cleanFilename)

	return filepath.Join(prefix, timestampedFilename)
}

// ExtractTimestampFromPath extracts the timestamp from a timestamped path
// Returns zero time if the path doesn't have a valid timestamp
func ExtractTimestampFromPath(path string) (time.Time, error) {
	filename := filepath.Base(path)

	// Expected format: YYYYMMDD_HHMMSS_microseconds_filename.ext
	parts := strings.Split(filename, "_")
	if len(parts) < 3 {
		return time.Time{}, fmt.Errorf("invalid timestamped path format")
	}

	// Parse YYYYMMDD
	dateStr := parts[0]
	if len(dateStr) != 8 {
		return time.Time{}, fmt.Errorf("invalid date format")
	}

	// Parse HHMMSS
	timeStr := parts[1]
	if len(timeStr) != 6 {
		return time.Time{}, fmt.Errorf("invalid time format")
	}

	// Parse microseconds
	microStr := parts[2]
	if len(microStr) != 6 {
		return time.Time{}, fmt.Errorf("invalid microseconds format")
	}

	// Combine into a parseable string
	timestampStr := dateStr + timeStr
	timestamp, err := time.Parse("20060102150405", timestampStr)
	if err != nil {
		return time.Time{}, fmt.Errorf("failed to parse timestamp: %w", err)
	}

	return timestamp, nil
}

// IsTimestampedPath checks if a path follows the timestamped format
func IsTimestampedPath(path string) bool {
	_, err := ExtractTimestampFromPath(path)
	return err == nil
}

// GetOriginalFilename extracts the original filename from a timestamped path
func GetOriginalFilename(timestampedPath string) string {
	filename := filepath.Base(timestampedPath)

	// Expected format: YYYYMMDD_HHMMSS_microseconds_filename.ext
	parts := strings.Split(filename, "_")
	if len(parts) < 4 {
		return filename // Return as-is if not in expected format
	}

	// Join everything after the third underscore
	originalParts := parts[3:]
	return strings.Join(originalParts, "_")
}
