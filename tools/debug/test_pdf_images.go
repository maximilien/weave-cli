// SPDX-License-Identifier: MIT
// Copyright (c) 2025 dr.max

package main

import (
	"fmt"
	"os"

	"github.com/maximilien/weave-cli/src/pkg/pdf"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintf(os.Stderr, "Usage: %s <pdf-file>\n", os.Args[0])
		os.Exit(1)
	}

	pdfPath := os.Args[1]

	fmt.Printf("🔍 Testing PDF image extraction for: %s\n\n", pdfPath)

	// Test PDF extraction
	textData, imageData, err := pdf.ExtractPDFContent(pdfPath, 1000, false, 0, false)
	if err != nil {
		fmt.Fprintf(os.Stderr, "❌ Error extracting PDF content: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("📝 Text Chunks Found: %d\n", len(textData))
	if len(textData) > 0 {
		fmt.Printf("   First chunk sample: %.100s...\n\n", textData[0].Content)
	}

	fmt.Printf("🖼️  Images Found: %d\n", len(imageData))
	if len(imageData) > 0 {
		for i, img := range imageData {
			fmt.Printf("   Image %d:\n", i+1)
			fmt.Printf("      - ID: %s\n", img.ID)
			fmt.Printf("      - Size: %d bytes\n", len(img.ImageData))
			fmt.Printf("      - Source: %s\n", img.SourcePDF)
			fmt.Printf("      - Page: %d\n", img.PageNumber)
			fmt.Printf("      - Caption: %s\n", img.Caption)
			if img.Metadata != nil {
				if format, ok := img.Metadata["image_format"]; ok {
					fmt.Printf("      - Format: %s\n", format)
				}
			}
			fmt.Println()
		}
	} else {
		fmt.Println("   ⚠️  No images were extracted from this PDF")
		fmt.Println()
		fmt.Println("   Possible reasons:")
		fmt.Println("   1. PDF contains no embedded images")
		fmt.Println("   2. Images are in unsupported format (e.g., CMYK JPEG without APP14)")
		fmt.Println("   3. Images are embedded as XObjects that pdfcpu cannot extract")
		fmt.Println("   4. PDF uses image compression not supported by extraction library")
		fmt.Println()
		fmt.Println("   To verify if PDF contains images:")
		fmt.Println("   - Open PDF in a viewer and check for images")
		fmt.Println("   - Run: pdfimages -list <file.pdf>  # requires poppler-utils")
		fmt.Println("   - Run: pdfcpu info <file.pdf>")
	}

	fmt.Printf("\n✅ Extraction test complete!\n")
}
