package main

import (
	"os"
	"testing"
)

func TestMainFunction(t *testing.T) {
	// Set up a temporary source and destination directory
	srcDir := t.TempDir()
	destDir := t.TempDir()

	// Create a mock file in the source directory
	mockFile := srcDir + "/test.jpg"
	if _, err := os.Create(mockFile); err != nil {
		t.Fatalf("Failed to create mock file: %v", err)
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
}

func mockInput(input string) func() {
	oldStdin := os.Stdin
	r, w, _ := os.Pipe()
	w.Write([]byte(input + "\n"))
	w.Close()
	os.Stdin = r
	return func() { os.Stdin = oldStdin }
}
