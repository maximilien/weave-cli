// SPDX-License-Identifier: MIT
// Copyright (c) 2026 dr.max

package document

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

func TestInspectTextAndImagePaths(t *testing.T) {
	root := t.TempDir()
	textPath := filepath.Join(root, "notes.txt")
	imagePath := filepath.Join(root, "photo.png")
	unknownPath := filepath.Join(root, "data.bin")
	if err := os.WriteFile(textPath, []byte(strings.Repeat("content ", 90)), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(imagePath, []byte("image fixture"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(unknownPath, []byte("unknown"), 0o600); err != nil {
		t.Fatal(err)
	}

	previousChunkSize, previousSummary := inspectChunkSize, inspectSummary
	inspectChunkSize, inspectSummary = 100, false
	t.Cleanup(func() {
		inspectChunkSize, inspectSummary = previousChunkSize, previousSummary
	})
	if err := runInspect(InspectCmd, []string{textPath}); err != nil {
		t.Fatalf("runInspect(text) error = %v", err)
	}
	inspectSummary = true
	if err := inspectText(textPath); err != nil {
		t.Fatalf("inspectText(summary) error = %v", err)
	}
	if err := runInspect(InspectCmd, []string{imagePath}); err != nil {
		t.Fatalf("runInspect(image) error = %v", err)
	}
	if err := inspectImage(filepath.Join(root, "missing.png")); err == nil {
		t.Fatal("inspectImage(missing) expected error")
	}
	if err := runInspect(InspectCmd, []string{unknownPath}); err == nil || !strings.Contains(err.Error(), "unsupported file type") {
		t.Fatalf("runInspect(unknown) error = %v", err)
	}
	if err := runInspect(InspectCmd, []string{filepath.Join(root, "missing.txt")}); err == nil || !strings.Contains(err.Error(), "does not exist") {
		t.Fatalf("runInspect(missing) error = %v", err)
	}
	invalidPDF := filepath.Join(root, "invalid.pdf")
	if err := os.WriteFile(invalidPDF, []byte("not a PDF"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := inspectPDF(invalidPDF); err == nil || !strings.Contains(err.Error(), "failed to extract PDF content") {
		t.Fatalf("inspectPDF(invalid) error = %v", err)
	}
}

func TestPDFConversionHelpersWithFakeTools(t *testing.T) {
	binDir := t.TempDir()
	writeExecutable(t, filepath.Join(binDir, "gs"), `#!/bin/sh
out=""
for arg in "$@"; do
  case "$arg" in -sOutputFile=*) out="${arg#*=}" ;; esac
done
printf 'converted by gs' > "$out"
`)
	writeExecutable(t, filepath.Join(binDir, "convert"), `#!/bin/sh
for arg in "$@"; do out="$arg"; done
printf 'converted by imagemagick' > "$out"
`)
	t.Setenv("PATH", binDir)

	if !isGhostscriptInstalled() || !isImageMagickInstalled() {
		t.Fatal("fake conversion tools were not detected")
	}
	root := t.TempDir()
	input := filepath.Join(root, "input.PDF")
	if err := os.WriteFile(input, []byte("PDF fixture"), 0o600); err != nil {
		t.Fatal(err)
	}

	ghostOutput := generateOutputFilename(input, ToolGhostscript)
	imageOutput := generateOutputFilename(input, ToolImageMagick)
	if !strings.HasSuffix(ghostOutput, "input-rgb-ghostscript.PDF") || !strings.HasSuffix(imageOutput, "input-rgb-imagemagick.PDF") {
		t.Fatalf("generated outputs = %q, %q", ghostOutput, imageOutput)
	}
	if got := generateOutputFilename("input.pdf", ToolGhostscript); got != "input-rgb-ghostscript.pdf" {
		t.Fatalf("relative output = %q", got)
	}
	if getToolName(ToolGhostscript) != "Ghostscript" || getToolName(ToolImageMagick) != "ImageMagick" {
		t.Fatal("unexpected conversion tool names")
	}
	if err := validateOutputDirectory(root); err != nil {
		t.Fatalf("validateOutputDirectory(valid) error = %v", err)
	}
	if err := validateOutputDirectory(filepath.Join(root, "missing")); err == nil {
		t.Fatal("validateOutputDirectory(missing) expected error")
	}
	if err := validateOutputDirectory(input); err == nil {
		t.Fatal("validateOutputDirectory(file) expected error")
	}
	if err := convertPDF(input, ghostOutput, ToolGhostscript); err != nil {
		t.Fatalf("convertPDF(ghostscript) error = %v", err)
	}
	if err := convertPDF(input, imageOutput, ToolImageMagick); err != nil {
		t.Fatalf("convertPDF(imagemagick) error = %v", err)
	}
	for _, output := range []string{ghostOutput, imageOutput} {
		if _, err := os.Stat(output); err != nil {
			t.Fatalf("converted output %s: %v", output, err)
		}
	}

	printInstallationTips()
	printConversionTroubleshootingTips(ToolGhostscript)
	printConversionTroubleshootingTips(ToolImageMagick)
}

func TestRunPDFConvertSingleAndDirectory(t *testing.T) {
	binDir := t.TempDir()
	writeExecutable(t, filepath.Join(binDir, "gs"), `#!/bin/sh
out=""
for arg in "$@"; do
  case "$arg" in -sOutputFile=*) out="${arg#*=}" ;; esac
done
printf 'converted' > "$out"
`)
	t.Setenv("PATH", binDir)
	root := t.TempDir()
	input := filepath.Join(root, "one.pdf")
	if err := os.WriteFile(input, []byte("PDF fixture"), 0o600); err != nil {
		t.Fatal(err)
	}

	cmd := newPDFConvertTestCommand()
	if err := cmd.Flags().Set("ghostscript", "true"); err != nil {
		t.Fatal(err)
	}
	runPdfConvert(cmd, []string{input})
	if _, err := os.Stat(filepath.Join(root, "one-rgb-ghostscript.pdf")); err != nil {
		t.Fatalf("single conversion output: %v", err)
	}

	nested := filepath.Join(root, "nested")
	if err := os.Mkdir(nested, 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "two.pdf"), []byte("PDF"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(nested, "three.PDF"), []byte("PDF"), 0o600); err != nil {
		t.Fatal(err)
	}
	runDirectoryConvert(cmd, root, true, false, false, false)
	runDirectoryConvert(cmd, root, true, false, false, true)
	for _, output := range []string{
		filepath.Join(root, "two-rgb-ghostscript.pdf"),
		filepath.Join(nested, "three-rgb-ghostscript.PDF"),
	} {
		if _, err := os.Stat(output); err != nil {
			t.Fatalf("directory conversion output %s: %v", output, err)
		}
	}
}

func TestConvertPDFReportsToolFailure(t *testing.T) {
	binDir := t.TempDir()
	writeExecutable(t, filepath.Join(binDir, "gs"), "#!/bin/sh\necho fixture failure >&2\nexit 2\n")
	t.Setenv("PATH", binDir)
	err := convertPDF("input.pdf", "output.pdf", ToolGhostscript)
	if err == nil || !strings.Contains(err.Error(), "fixture failure") {
		t.Fatalf("convertPDF(failure) error = %v", err)
	}
}

func newPDFConvertTestCommand() *cobra.Command {
	cmd := &cobra.Command{}
	cmd.Flags().Bool("ghostscript", false, "")
	cmd.Flags().Bool("imagemagick", false, "")
	cmd.Flags().Bool("rgb", false, "")
	cmd.Flags().String("converted-filename", "", "")
	cmd.Flags().String("directory", "", "")
	cmd.Flags().Bool("recurse", false, "")
	return cmd
}

func writeExecutable(t *testing.T, path, contents string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(contents), 0o700); err != nil {
		t.Fatal(err)
	}
}
