// SPDX-License-Identifier: MIT
// Copyright (c) 2025 dr.max

package document

import (
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/maximilien/weave-cli/src/cmd/utils"
	"github.com/spf13/cobra"
)

// ConversionTool represents the tool to use for PDF conversion
type ConversionTool int

const (
	ToolGhostscript ConversionTool = iota
	ToolImageMagick
)

// PdfConvertCmd represents the pdf-convert command
var PdfConvertCmd = &cobra.Command{
	Use:   "pdf-convert [PDF_FILENAME]",
	Short: "Convert PDF with CMYK images to RGB format",
	Long: `Convert a PDF file with CMYK images to RGB format using Ghostscript or ImageMagick.

This command is useful when you encounter PDFs with CMYK JPEG images that fail
during image extraction. The conversion process creates a new PDF with RGB images
that are compatible with standard image processing libraries.

Conversion Methods:
- Ghostscript (default, recommended) - High-quality RGB conversion
- ImageMagick - Alternative conversion method

The converted PDF will have the same text content and structure, but with RGB images
that can be successfully extracted and processed.

Examples:
  # Convert a single file using Ghostscript (default)
  weave docs pdf-convert document.pdf --ghostscript

  # Convert all PDFs in a directory (non-recursive)
  weave docs pdf-convert --directory /path/to/pdfs

  # Convert all PDFs in a directory and subdirectories
  weave docs pdf-convert --directory /path/to/pdfs --recurse

  # Convert using ImageMagick
  weave docs pdf-convert document.pdf --imagemagick

  # Specify custom output filename
  weave docs pdf-convert document.pdf --ghostscript --converted-filename output.pdf

  # Convert to RGB (auto-detects tool)
  weave docs pdf-convert document.pdf --rgb`,
	Args: cobra.MaximumNArgs(1),
	Run:  runPdfConvert,
}

func init() {
	DocumentCmd.AddCommand(PdfConvertCmd)

	PdfConvertCmd.Flags().Bool("ghostscript", false, "Use Ghostscript for conversion (recommended)")
	PdfConvertCmd.Flags().Bool("imagemagick", false, "Use ImageMagick for conversion")
	PdfConvertCmd.Flags().Bool("rgb", false, "Convert to RGB format (auto-detects available tool)")
	PdfConvertCmd.Flags().String("converted-filename", "", "Output filename for converted PDF (default: <filename>-rgb-<tool>.pdf)")
	PdfConvertCmd.Flags().StringP("directory", "d", "", "Convert all PDF files in the specified directory")
	PdfConvertCmd.Flags().Bool("recurse", false, "Recursively process subdirectories when using --directory")
}

func runPdfConvert(cmd *cobra.Command, args []string) {
	useGhostscript, _ := cmd.Flags().GetBool("ghostscript")
	useImageMagick, _ := cmd.Flags().GetBool("imagemagick")
	useRGB, _ := cmd.Flags().GetBool("rgb")
	convertedFilename, _ := cmd.Flags().GetString("converted-filename")
	directory, _ := cmd.Flags().GetString("directory")
	recurse, _ := cmd.Flags().GetBool("recurse")

	// Check if directory mode or single file mode
	if directory != "" {
		// Directory mode
		if len(args) > 0 {
			utils.PrintError("Cannot specify both a file and --directory flag")
			os.Exit(1)
		}
		if convertedFilename != "" {
			utils.PrintError("Cannot use --converted-filename with --directory flag")
			os.Exit(1)
		}
		runDirectoryConvert(cmd, directory, useGhostscript, useImageMagick, useRGB, recurse)
		return
	}

	// Check if --recurse is used without --directory
	if recurse {
		utils.PrintError("--recurse flag can only be used with --directory flag")
		os.Exit(1)
	}

	// Single file mode
	if len(args) == 0 {
		utils.PrintError("Either specify a PDF file or use --directory flag")
		os.Exit(1)
	}

	inputFile := args[0]

	// Validate input file exists
	if _, err := os.Stat(inputFile); os.IsNotExist(err) {
		utils.PrintError(fmt.Sprintf("File not found: %s", inputFile))
		os.Exit(1)
	}

	// Ensure input is a PDF
	if !strings.HasSuffix(strings.ToLower(inputFile), ".pdf") {
		utils.PrintError("Input file must be a PDF (.pdf extension)")
		os.Exit(1)
	}

	// Determine which tool to use
	var tool ConversionTool
	if useGhostscript {
		tool = ToolGhostscript
	} else if useImageMagick {
		tool = ToolImageMagick
	} else if useRGB {
		// Auto-detect available tool
		if isGhostscriptInstalled() {
			tool = ToolGhostscript
		} else if isImageMagickInstalled() {
			tool = ToolImageMagick
		} else {
			utils.PrintError("Neither Ghostscript nor ImageMagick is installed")
			printInstallationTips()
			os.Exit(1)
		}
	} else {
		// Default to Ghostscript if available
		if isGhostscriptInstalled() {
			tool = ToolGhostscript
		} else if isImageMagickInstalled() {
			tool = ToolImageMagick
		} else {
			utils.PrintError("No conversion tool specified and none found")
			fmt.Println("\n💡 Please specify --ghostscript or --imagemagick, or use --rgb to auto-detect")
			printInstallationTips()
			os.Exit(1)
		}
	}

	// Check if the selected tool is installed
	if tool == ToolGhostscript && !isGhostscriptInstalled() {
		utils.PrintError("Ghostscript is not installed")
		printGhostscriptInstallationTip()
		os.Exit(1)
	}
	if tool == ToolImageMagick && !isImageMagickInstalled() {
		utils.PrintError("ImageMagick is not installed")
		printImageMagickInstallationTip()
		os.Exit(1)
	}

	// Generate output filename if not specified
	outputFile := convertedFilename
	if outputFile == "" {
		outputFile = generateOutputFilename(inputFile, tool)
	}

	// Validate output directory exists and is writable
	outputDir := filepath.Dir(outputFile)
	if outputDir == "" {
		outputDir = "."
	}
	if err := validateOutputDirectory(outputDir); err != nil {
		utils.PrintError(fmt.Sprintf("Cannot write to output directory: %v", err))
		os.Exit(1)
	}

	// Perform conversion
	fmt.Printf("🔄 Converting PDF using %s...\n", getToolName(tool))
	fmt.Printf("   Input:  %s\n", inputFile)
	fmt.Printf("   Output: %s\n", outputFile)

	if err := convertPDF(inputFile, outputFile, tool); err != nil {
		utils.PrintError(fmt.Sprintf("Conversion failed: %v", err))
		printConversionTroubleshootingTips(tool)
		os.Exit(1)
	}

	// Verify output file was created
	if _, err := os.Stat(outputFile); os.IsNotExist(err) {
		utils.PrintError("Conversion completed but output file was not created")
		os.Exit(1)
	}

	utils.PrintSuccess(fmt.Sprintf("PDF converted successfully: %s", outputFile))
	fmt.Printf("\n💡 You can now process this PDF with:\n")
	fmt.Printf("   weave docs create <collection> %s --image-collection <image-collection>\n", outputFile)
}

// runDirectoryConvert processes all PDF files in a directory
func runDirectoryConvert(cmd *cobra.Command, directory string, useGhostscript, useImageMagick, useRGB, recurse bool) {
	// Validate directory exists
	info, err := os.Stat(directory)
	if os.IsNotExist(err) {
		utils.PrintError(fmt.Sprintf("Directory not found: %s", directory))
		os.Exit(1)
	}
	if err != nil {
		utils.PrintError(fmt.Sprintf("Error accessing directory: %v", err))
		os.Exit(1)
	}
	if !info.IsDir() {
		utils.PrintError(fmt.Sprintf("Not a directory: %s", directory))
		os.Exit(1)
	}

	// Determine which tool to use
	var tool ConversionTool
	if useGhostscript {
		tool = ToolGhostscript
	} else if useImageMagick {
		tool = ToolImageMagick
	} else if useRGB {
		// Auto-detect available tool
		if isGhostscriptInstalled() {
			tool = ToolGhostscript
		} else if isImageMagickInstalled() {
			tool = ToolImageMagick
		} else {
			utils.PrintError("Neither Ghostscript nor ImageMagick is installed")
			printInstallationTips()
			os.Exit(1)
		}
	} else {
		// Default to Ghostscript if available
		if isGhostscriptInstalled() {
			tool = ToolGhostscript
		} else if isImageMagickInstalled() {
			tool = ToolImageMagick
		} else {
			utils.PrintError("No conversion tool specified and none found")
			fmt.Println("\n💡 Please specify --ghostscript or --imagemagick, or use --rgb to auto-detect")
			printInstallationTips()
			os.Exit(1)
		}
	}

	// Check if the selected tool is installed
	if tool == ToolGhostscript && !isGhostscriptInstalled() {
		utils.PrintError("Ghostscript is not installed")
		printGhostscriptInstallationTip()
		os.Exit(1)
	}
	if tool == ToolImageMagick && !isImageMagickInstalled() {
		utils.PrintError("ImageMagick is not installed")
		printImageMagickInstallationTip()
		os.Exit(1)
	}

	// Find all PDF files in the directory
	var pdfFiles []string
	if recurse {
		// Recursively scan subdirectories
		err = filepath.WalkDir(directory, func(path string, d fs.DirEntry, err error) error {
			if err != nil {
				return err
			}
			if !d.IsDir() && strings.HasSuffix(strings.ToLower(d.Name()), ".pdf") {
				pdfFiles = append(pdfFiles, path)
			}
			return nil
		})
	} else {
		// Only scan the specified directory (non-recursive)
		entries, err := os.ReadDir(directory)
		if err != nil {
			utils.PrintError(fmt.Sprintf("Error reading directory: %v", err))
			os.Exit(1)
		}
		for _, entry := range entries {
			if !entry.IsDir() && strings.HasSuffix(strings.ToLower(entry.Name()), ".pdf") {
				pdfFiles = append(pdfFiles, filepath.Join(directory, entry.Name()))
			}
		}
		err = nil
	}
	if err != nil {
		utils.PrintError(fmt.Sprintf("Error scanning directory: %v", err))
		os.Exit(1)
	}

	if len(pdfFiles) == 0 {
		if recurse {
			utils.PrintError(fmt.Sprintf("No PDF files found in directory (including subdirectories): %s", directory))
		} else {
			utils.PrintError(fmt.Sprintf("No PDF files found in directory: %s", directory))
		}
		os.Exit(1)
	}

	if recurse {
		fmt.Printf("🔄 Found %d PDF file(s) in directory (including subdirectories): %s\n", len(pdfFiles), directory)
	} else {
		fmt.Printf("🔄 Found %d PDF file(s) in directory: %s\n", len(pdfFiles), directory)
	}
	fmt.Printf("   Using %s for conversion\n\n", getToolName(tool))

	// Convert each PDF
	successCount := 0
	failCount := 0
	for i, inputFile := range pdfFiles {
		outputFile := generateOutputFilename(inputFile, tool)

		fmt.Printf("[%d/%d] Converting: %s\n", i+1, len(pdfFiles), filepath.Base(inputFile))
		fmt.Printf("        Output: %s\n", filepath.Base(outputFile))

		if err := convertPDF(inputFile, outputFile, tool); err != nil {
			utils.PrintError(fmt.Sprintf("Failed: %v", err))
			failCount++
		} else {
			// Verify output file was created
			if _, err := os.Stat(outputFile); os.IsNotExist(err) {
				utils.PrintError("Conversion completed but output file was not created")
				failCount++
			} else {
				utils.PrintSuccess("Converted successfully")
				successCount++
			}
		}
		fmt.Println()
	}

	// Print summary
	fmt.Printf("📊 Conversion Summary:\n")
	fmt.Printf("   Total:   %d files\n", len(pdfFiles))
	fmt.Printf("   Success: %d files\n", successCount)
	if failCount > 0 {
		fmt.Printf("   Failed:  %d files\n", failCount)
	}
}

// isGhostscriptInstalled checks if Ghostscript is installed
func isGhostscriptInstalled() bool {
	_, err := exec.LookPath("gs")
	return err == nil
}

// isImageMagickInstalled checks if ImageMagick is installed
func isImageMagickInstalled() bool {
	_, err := exec.LookPath("convert")
	return err == nil
}

// generateOutputFilename generates the output filename based on the input and tool
func generateOutputFilename(inputFile string, tool ConversionTool) string {
	dir := filepath.Dir(inputFile)
	base := filepath.Base(inputFile)
	ext := filepath.Ext(base)
	name := strings.TrimSuffix(base, ext)

	var suffix string
	if tool == ToolGhostscript {
		suffix = "-rgb-ghostscript"
	} else {
		suffix = "-rgb-imagemagick"
	}

	outputBase := name + suffix + ext
	if dir == "." || dir == "" {
		return outputBase
	}
	return filepath.Join(dir, outputBase)
}

// validateOutputDirectory checks if the output directory exists and is writable
func validateOutputDirectory(dir string) error {
	info, err := os.Stat(dir)
	if os.IsNotExist(err) {
		return fmt.Errorf("directory does not exist: %s", dir)
	}
	if err != nil {
		return err
	}
	if !info.IsDir() {
		return fmt.Errorf("not a directory: %s", dir)
	}

	// Test write permission by creating a temporary file
	tempFile := filepath.Join(dir, ".weave-write-test")
	f, err := os.Create(tempFile)
	if err != nil {
		return fmt.Errorf("directory is not writable: %s", dir)
	}
	f.Close()
	os.Remove(tempFile)

	return nil
}

// convertPDF performs the actual PDF conversion
func convertPDF(inputFile, outputFile string, tool ConversionTool) error {
	var cmd *exec.Cmd

	if tool == ToolGhostscript {
		// Use the same Ghostscript command from the tips
		cmd = exec.Command("gs",
			"-sDEVICE=pdfwrite",
			"-dProcessColorModel=/DeviceRGB",
			"-dColorConversionStrategy=/RGB",
			"-dNOPAUSE",
			"-dBATCH",
			"-sOutputFile="+outputFile,
			inputFile,
		)
	} else {
		// Use ImageMagick convert command
		cmd = exec.Command("convert",
			"-density", "300",
			"-colorspace", "RGB",
			inputFile,
			outputFile,
		)
	}

	// Capture output for error reporting
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("%v\nOutput: %s", err, string(output))
	}

	return nil
}

// getToolName returns the human-readable name of the tool
func getToolName(tool ConversionTool) string {
	if tool == ToolGhostscript {
		return "Ghostscript"
	}
	return "ImageMagick"
}

// printInstallationTips prints installation tips for both tools
func printInstallationTips() {
	fmt.Println("\n💡 Installation Tips:")
	printGhostscriptInstallationTip()
	printImageMagickInstallationTip()
}

// printGhostscriptInstallationTip prints Ghostscript installation instructions
func printGhostscriptInstallationTip() {
	fmt.Println("\n   Ghostscript (recommended):")
	fmt.Println("   • macOS:   brew install ghostscript")
	fmt.Println("   • Ubuntu:  sudo apt-get install ghostscript")
	fmt.Println("   • Fedora:  sudo dnf install ghostscript")
	fmt.Println("   • Windows: Download from https://www.ghostscript.com/download/gsdnld.html")
}

// printImageMagickInstallationTip prints ImageMagick installation instructions
func printImageMagickInstallationTip() {
	fmt.Println("\n   ImageMagick:")
	fmt.Println("   • macOS:   brew install imagemagick")
	fmt.Println("   • Ubuntu:  sudo apt-get install imagemagick")
	fmt.Println("   • Fedora:  sudo dnf install ImageMagick")
	fmt.Println("   • Windows: Download from https://imagemagick.org/script/download.php")
}

// printConversionTroubleshootingTips prints troubleshooting tips when conversion fails
func printConversionTroubleshootingTips(tool ConversionTool) {
	fmt.Println("\n💡 Troubleshooting Tips:")

	if tool == ToolGhostscript {
		fmt.Println("\n   If Ghostscript conversion failed:")
		fmt.Println("   • Check that the input PDF is valid")
		fmt.Println("   • Try using ImageMagick instead: --imagemagick")
		fmt.Println("   • Ensure you have write permissions in the output directory")
		fmt.Println("   • Check Ghostscript version: gs --version")
	} else {
		fmt.Println("\n   If ImageMagick conversion failed:")
		fmt.Println("   • Check that the input PDF is valid")
		fmt.Println("   • Try using Ghostscript instead: --ghostscript")
		fmt.Println("   • Ensure you have write permissions in the output directory")
		fmt.Println("   • Check ImageMagick version: convert --version")
		fmt.Println("   • Note: Some ImageMagick installations have PDF security policies")
		fmt.Println("     that may need to be adjusted in /etc/ImageMagick-*/policy.xml")
	}
}
