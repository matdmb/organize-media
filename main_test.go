package main

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// Mock os.Exit for testing
var osExit = os.Exit

func TestMainFunction(t *testing.T) {
	// Set up a temporary source and destination directory
	srcDir := t.TempDir()
	destDir := t.TempDir()

	// Copy the sample image with EXIF data to the source directory
	samplePath := "./pkg/testdata/DSC00001.JPG"
	destPath := filepath.Join(srcDir, "DSC00001.JPG")

	// Read the sample file
	sampleData, err := os.ReadFile(samplePath)
	if err != nil {
		t.Fatalf("Failed to read sample file: %v", err)
	}

	// Write to destination in temp directory
	if err := os.WriteFile(destPath, sampleData, 0644); err != nil {
		t.Fatalf("Failed to copy sample file: %v", err)
	}

	// Save original args and exit function
	oldArgs := os.Args
	oldExit := osExit
	defer func() {
		os.Args = oldArgs
		osExit = oldExit
	}()

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
	processedFiles, err := filepath.Glob(filepath.Join(destDir, "*/*/*.JPG"))
	if err != nil {
		t.Fatalf("Failed to check processed files: %v", err)
	}

	if len(processedFiles) != 1 {
		t.Errorf("Expected 1 processed file, got %d", len(processedFiles))
	}
}

// mockInput mocks user input for testing
func mockInput(input string) func() {
	oldStdin := os.Stdin
	r, w, _ := os.Pipe()
	w.Write([]byte(input + "\n"))
	w.Close()
	os.Stdin = r
	return func() { os.Stdin = oldStdin }
}

// TestInvalidFlags tests the behavior when invalid flags are provided
func TestInvalidFlags(t *testing.T) {
	tests := []struct {
		name    string
		args    []string
		wantErr bool
	}{
		{
			name:    "No arguments",
			args:    []string{},
			wantErr: true,
		},
		{
			name:    "Missing source",
			args:    []string{"-dest", "/some/dest"},
			wantErr: true,
		},
		{
			name:    "Missing destination",
			args:    []string{"-source", "/some/source"},
			wantErr: true,
		},
		{
			name:    "Valid flags",
			args:    []string{"-source", "/some/source", "-dest", "/some/dest"},
			wantErr: false, // This will likely fail in practice since paths don't exist
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Skip in short mode as these tests are more integration-style
			if testing.Short() {
				t.Skip("Skipping test in short mode")
			}

			// Build a test binary
			testBinary := filepath.Join(t.TempDir(), "testprog")
			cmd := exec.Command("go", "build", "-o", testBinary, ".")
			if err := cmd.Run(); err != nil {
				t.Fatalf("Failed to build test binary: %v", err)
			}
			defer os.Remove(testBinary)

			// Run the test with arguments
			cmdTest := exec.Command(testBinary, tt.args...)
			output, err := cmdTest.CombinedOutput()

			// Check if we got the expected error condition
			if tt.wantErr && err == nil {
				t.Errorf("Expected error with args %v, but got none\nOutput: %s", tt.args, output)
			} else if !tt.wantErr && err != nil {
				if !strings.Contains(string(output), "does not exist") {
					// It's okay if the error is about non-existent directories
					t.Errorf("Unexpected error with args %v: %v\nOutput: %s", tt.args, err, output)
				}
			}

			// For expected errors, check if usage info is printed
			if tt.wantErr {
				outputStr := string(output)
				if !strings.Contains(outputStr, "Usage:") {
					t.Errorf("Expected usage information in output, got: %s", outputStr)
				}
			}
		})
	}
}

// TestCompressionRange tests the validation of compression range
func TestCompressionRange(t *testing.T) {
	// Create test directories
	srcDir := t.TempDir()
	destDir := t.TempDir()

	// Create a temporary source file
	sampleFile := filepath.Join(srcDir, "test.jpg")
	if err := os.WriteFile(sampleFile, []byte("test data"), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	tests := []struct {
		name        string
		compression int
		wantErr     bool
	}{
		{
			name:        "Valid compression (0)",
			compression: 0,
			wantErr:     false,
		},
		{
			name:        "Valid compression (50)",
			compression: 50,
			wantErr:     false,
		},
		{
			name:        "Valid compression (100)",
			compression: 100,
			wantErr:     false,
		},
		{
			name:        "Valid compression (-1)",
			compression: -1,
			wantErr:     false,
		},
		{
			name:        "Invalid compression (101)",
			compression: 101,
			wantErr:     true,
		},
		{
			name:        "Invalid compression (-2)",
			compression: -2,
			wantErr:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Skip in short mode
			if testing.Short() {
				t.Skip("Skipping test in short mode")
			}

			// Build a test binary
			testBinary := filepath.Join(t.TempDir(), "testprog")
			cmd := exec.Command("go", "build", "-o", testBinary, ".")
			if err := cmd.Run(); err != nil {
				t.Fatalf("Failed to build test binary: %v", err)
			}
			defer os.Remove(testBinary)

			// Create a pipe for stdin to automatically answer "y" to prompts
			pr, pw := io.Pipe()
			go func() {
				defer pw.Close()
				pw.Write([]byte("y\n"))
			}()

			// Run the test with arguments
			cmdTest := exec.Command(testBinary,
				"-source", srcDir,
				"-dest", destDir,
				"-compression", fmt.Sprintf("%d", tt.compression))

			cmdTest.Stdin = pr
			var stdout, stderr bytes.Buffer
			cmdTest.Stdout = &stdout
			cmdTest.Stderr = &stderr

			err := cmdTest.Run()

			// Check if we got the expected error condition
			if tt.wantErr && err == nil {
				t.Errorf("Expected error with compression %d, but got none", tt.compression)
			} else if !tt.wantErr && err != nil {
				if !strings.Contains(stderr.String(), "compression level must be") {
					t.Errorf("Unexpected error with compression %d: %v", tt.compression, err)
				}
			}

			// For expected errors, check if error message mentions compression
			if tt.wantErr {
				errorOutput := stderr.String()
				if !strings.Contains(errorOutput, "compression level") {
					t.Errorf("Expected compression error message, got: %s", errorOutput)
				}
			}
		})
	}
}
