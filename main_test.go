package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestRun(t *testing.T) {
	// Define table of test cases
	tests := []struct {
		name        string
		setup       func() *Params // Function to set up test parameters
		expectError bool           // Whether an error is expected
	}{
		{
			name: "ValidParams",
			setup: func() *Params {
				srcDir := t.TempDir()
				destDir := t.TempDir()

				// Create a mock file in the source directory
				mockFile := filepath.Join(srcDir, "test.jpg")
				if _, err := os.Create(mockFile); err != nil {
					t.Fatalf("Failed to create mock file: %v", err)
				}

				return &Params{
					Source:        srcDir,
					Destination:   destDir,
					Compression:   -1,   // No compression
					SkipUserInput: true, // Skip user input for automated testing
				}
			},
			expectError: false,
		},
		{
			name: "WithCompression",
			setup: func() *Params {
				srcDir := t.TempDir()
				destDir := t.TempDir()

				// Create a mock file in the source directory
				mockFile := filepath.Join(srcDir, "test.jpg")
				if _, err := os.Create(mockFile); err != nil {
					t.Fatalf("Failed to create mock file: %v", err)
				}

				return &Params{
					Source:        srcDir,
					Destination:   destDir,
					Compression:   50,   // Valid compression value
					SkipUserInput: true, // Skip user input for automated testing
				}
			},
			expectError: false,
		},
		{
			name: "MissingSource",
			setup: func() *Params {
				return &Params{
					Source:      "/non/existent/source",
					Destination: t.TempDir(),
					Compression: -1,
				}
			},
			expectError: true,
		},
		{
			name: "MissingDestination",
			setup: func() *Params {
				return &Params{
					Source:      t.TempDir(),
					Destination: "/non/existent/destination",
					Compression: -1,
				}
			},
			expectError: true,
		},
		{
			name: "InvalidCompressionValue",
			setup: func() *Params {
				srcDir := t.TempDir()
				destDir := t.TempDir()

				return &Params{
					Source:        srcDir,
					Destination:   destDir,
					Compression:   200, // Invalid compression value
					SkipUserInput: true,
				}
			},
			expectError: true,
		},
		{
			name: "ReadOnlyDestination",
			setup: func() *Params {
				srcDir := t.TempDir()
				destDir := t.TempDir()

				// Create a mock file in the source directory
				mockFile := filepath.Join(srcDir, "test.jpg")
				if _, err := os.Create(mockFile); err != nil {
					t.Fatalf("Failed to create mock file: %v", err)
				}

				// Make the destination directory read-only
				if err := os.Chmod(destDir, 0444); err != nil {
					t.Fatalf("Failed to make destination directory read-only: %v", err)
				}

				return &Params{
					Source:        srcDir,
					Destination:   destDir,
					Compression:   -1,
					SkipUserInput: true,
				}
			},
			expectError: true,
		},
		{
			name: "EmptySourceDirectory",
			setup: func() *Params {
				return &Params{
					Source:        t.TempDir(), // Empty source directory
					Destination:   t.TempDir(),
					Compression:   -1,
					SkipUserInput: true,
				}
			},
			expectError: true,
		},
	}

	// Iterate through each test case
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup test parameters
			params := tt.setup()

			// Execute the run function
			err := run(params)

			// Validate the outcome
			if (err != nil) != tt.expectError {
				t.Errorf("run() test %q failed. Expected error: %v, got: %v", tt.name, tt.expectError, err)
			}
		})
	}
}

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
