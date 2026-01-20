// SPDX-License-Identifier: MIT
// Copyright (c) 2025 dr.max

package document

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/maximilien/weave-cli/src/pkg/pdf"
	"github.com/spf13/cobra"
)

// InspectCmd represents the document inspect command
var InspectCmd = &cobra.Command{
	Use:   "inspect FILE",
	Short: "Inspect a document file and show extraction preview",
	Long: `Inspect a document file to preview what would be extracted without uploading to a database.

This is useful for:
- Verifying PDF image extraction works
- Checking text chunk sizes before ingestion
- Debugging document processing issues
- Understanding what metadata will be created

Supported file types:
- PDF (.pdf) - Shows text chunks and embedded images
- Text (.txt, .md, .json, .yaml)
- Images (.jpg, .png, .gif)

Examples:
  # Inspect a PDF and see what would be extracted
  weave docs inspect document.pdf

  # Inspect with custom chunk size
  weave docs inspect document.pdf --chunk-size 2000

  # Show only summary (no detailed output)
  weave docs inspect document.pdf --summary`,
	Args: cobra.ExactArgs(1),
	RunE: runInspect,
}

var (
	inspectChunkSize int
	inspectSummary   bool
)

func init() {
	InspectCmd.Flags().IntVar(&inspectChunkSize, "chunk-size", 1000, "Text chunk size for PDFs")
	InspectCmd.Flags().BoolVar(&inspectSummary, "summary", false, "Show only summary, not detailed output")
}

func runInspect(cmd *cobra.Command, args []string) error {
	filePath := args[0]

	// Check file exists
	if _, err := os.Stat(filePath); err != nil {
		return fmt.Errorf("file does not exist: %s", filePath)
	}

	// Get file extension
	ext := filepath.Ext(filePath)

	fmt.Printf("🔍 Inspecting: %s\n", filePath)
	fmt.Printf("📁 File type: %s\n\n", ext)

	switch ext {
	case ".pdf":
		return inspectPDF(filePath)
	case ".txt", ".md", ".json", ".yaml", ".yml":
		return inspectText(filePath)
	case ".jpg", ".jpeg", ".png", ".gif", ".bmp":
		return inspectImage(filePath)
	default:
		return fmt.Errorf("unsupported file type: %s", ext)
	}
}

func inspectPDF(filePath string) error {
	fmt.Println("📄 Extracting PDF content...")

	textData, imageData, err := pdf.ExtractPDFContent(filePath, inspectChunkSize, false, 0, 2000, true)
	if err != nil {
		return fmt.Errorf("failed to extract PDF content: %w", err)
	}

	// Show summary
	fmt.Printf("\n📊 Extraction Summary:\n")
	fmt.Printf("   Text Chunks: %d\n", len(textData))
	fmt.Printf("   Images:      %d\n", len(imageData))

	if len(textData) > 0 {
		totalChars := 0
		for _, chunk := range textData {
			totalChars += len(chunk.Content)
		}
		avgChunkSize := totalChars / len(textData)
		fmt.Printf("   Avg Chunk:   %d characters\n", avgChunkSize)
	}

	if len(imageData) > 0 {
		totalSize := 0
		for _, img := range imageData {
			totalSize += len(img.ImageData)
		}
		avgImageSize := totalSize / len(imageData)
		fmt.Printf("   Avg Image:   %d bytes\n", avgImageSize)
	}

	// Detailed output (unless --summary)
	if !inspectSummary {
		if len(textData) > 0 {
			fmt.Printf("\n📝 Text Chunks Preview (showing first 3):\n")
			for i, chunk := range textData {
				if i >= 3 {
					break
				}
				preview := chunk.Content
				if len(preview) > 200 {
					preview = preview[:200] + "..."
				}
				fmt.Printf("\n   Chunk %d (page %d):\n", i+1, chunk.PageNumber)
				fmt.Printf("   %s\n", preview)
			}
			if len(textData) > 3 {
				fmt.Printf("\n   ... and %d more text chunks\n", len(textData)-3)
			}
		}

		if len(imageData) > 0 {
			fmt.Printf("\n🖼️  Images Preview (showing first 5):\n")
			for i, img := range imageData {
				if i >= 5 {
					break
				}
				fmt.Printf("\n   Image %d:\n", i+1)
				fmt.Printf("      - Page:    %d\n", img.PageNumber)
				fmt.Printf("      - Size:    %d bytes\n", len(img.ImageData))
				fmt.Printf("      - Caption: %s\n", img.Caption)
				if format, ok := img.Metadata["image_format"]; ok {
					fmt.Printf("      - Format:  %s\n", format)
				}
			}
			if len(imageData) > 5 {
				fmt.Printf("\n   ... and %d more images\n", len(imageData)-5)
			}
		}
	}

	fmt.Printf("\n✅ Inspection complete!\n")
	fmt.Printf("\n💡 To upload this PDF to a database:\n")
	fmt.Printf("   weave docs create <collection> %s\n", filepath.Base(filePath))
	if len(imageData) > 0 {
		fmt.Printf("   weave docs create <collection> %s --image-collection <image-col>\n", filepath.Base(filePath))
	}

	return nil
}

func inspectText(filePath string) error {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("failed to read file: %w", err)
	}

	fmt.Printf("📊 File Statistics:\n")
	fmt.Printf("   Size:       %d bytes\n", len(data))
	fmt.Printf("   Characters: %d\n", len(string(data)))

	// Estimate chunks
	chunkSize := inspectChunkSize
	numChunks := (len(data) + chunkSize - 1) / chunkSize
	fmt.Printf("   Chunks:     %d (with chunk size %d)\n", numChunks, chunkSize)

	if !inspectSummary && len(data) > 0 {
		preview := string(data)
		if len(preview) > 500 {
			preview = preview[:500] + "..."
		}
		fmt.Printf("\n📝 Content Preview:\n%s\n", preview)
	}

	fmt.Printf("\n✅ Inspection complete!\n")
	fmt.Printf("\n💡 To upload this file to a database:\n")
	fmt.Printf("   weave docs create <collection> %s\n", filepath.Base(filePath))

	return nil
}

func inspectImage(filePath string) error {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("failed to read file: %w", err)
	}

	fmt.Printf("📊 Image Statistics:\n")
	fmt.Printf("   Size:   %d bytes\n", len(data))
	fmt.Printf("   Format: %s\n", filepath.Ext(filePath))

	fmt.Printf("\n✅ Inspection complete!\n")
	fmt.Printf("\n💡 To upload this image to a database:\n")
	fmt.Printf("   weave docs create <image-collection> %s\n", filepath.Base(filePath))

	return nil
}
