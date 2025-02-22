package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestMainFunction(t *testing.T) {
	// Set up a temporary source and destination directory
	srcDir := t.TempDir()
	destDir := t.TempDir()

	// Copy the sample image with EXIF data to the source directory
	samplePath := "./pkg/testdata/exif/sample_with_exif.jpg"
	destPath := filepath.Join(srcDir, "sample_with_exif.jpg")

	// Read the sample file
	sampleData, err := os.ReadFile(samplePath)
	if err != nil {
		t.Fatalf("Failed to read sample file: %v", err)
	}

	// Write to destination in temp directory
	if err := os.WriteFile(destPath, sampleData, 0644); err != nil {
		t.Fatalf("Failed to copy sample file: %v", err)
	}

	// Mock `os.Stdin` to automatically provide input
	defer mockInput("y")()

	// Mock command-line arguments
	os.Args = []string{"main", "-source", srcDir, "-dest", destDir, "-compression", "50"}

	// Run the main function
	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("main() panicked: %v", r)
		}
	}()

	main()

	// Verify the file was processed
	processedFiles, err := filepath.Glob(filepath.Join(destDir, "*/*/*.jpg"))
	if err != nil {
		t.Fatalf("Failed to check processed files: %v", err)
	}

	if len(processedFiles) != 1 {
		t.Errorf("Expected 1 processed file, got %d", len(processedFiles))
	}
}

func mockInput(input string) func() {
	oldStdin := os.Stdin
	r, w, _ := os.Pipe()
	w.Write([]byte(input + "\n"))
	w.Close()
	os.Stdin = r
	return func() { os.Stdin = oldStdin }
}
