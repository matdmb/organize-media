package organizemedia

import (
	"os"
	"path/filepath"
	"testing"
)

func TestFormatSize(t *testing.T) {
	tests := []struct {
		name     string
		size     int64
		expected string
	}{
		{
			name:     "Bytes",
			size:     500,
			expected: "500 bytes",
		},
		{
			name:     "Kilobytes",
			size:     1500,
			expected: "1.46 KB",
		},
		{
			name:     "Megabytes",
			size:     1500000,
			expected: "1.43 MB",
		},
		{
			name:     "Gigabytes",
			size:     1500000000,
			expected: "1.40 GB",
		},
		{
			name:     "Zero",
			size:     0,
			expected: "0 bytes",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatSize(tt.size)
			if result != tt.expected {
				t.Errorf("formatSize(%d) = %s; want %s", tt.size, result, tt.expected)
			}
		})
	}
}

func TestSetupLogger(t *testing.T) {
	// Clean up after tests
	defer os.RemoveAll("./logs")

	tests := []struct {
		name      string
		enableLog bool
		wantErr   bool
	}{
		{
			name:      "logging disabled",
			enableLog: false,
			wantErr:   false,
		},
		{
			name:      "logging enabled",
			enableLog: true,
			wantErr:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			writer, err := setupLogger(tt.enableLog)

			// Check error cases
			if (err != nil) != tt.wantErr {
				t.Errorf("setupLogger() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			// Check writer is not nil
			if writer == nil {
				t.Error("setupLogger() returned nil writer")
				return
			}

			if tt.enableLog {
				// Verify logs directory was created
				if _, err := os.Stat("./logs"); os.IsNotExist(err) {
					t.Error("logs directory was not created")
				}

				// Verify at least one log file exists
				files, err := os.ReadDir("./logs")
				if err != nil {
					t.Errorf("failed to read logs directory: %v", err)
				}
				if len(files) == 0 {
					t.Error("no log files were created")
				}

				// Verify log file is writable
				logFile := files[0]
				logPath := filepath.Join("./logs", logFile.Name())
				f, err := os.OpenFile(logPath, os.O_WRONLY|os.O_APPEND, 0666)
				if err != nil {
					t.Errorf("log file is not writable: %v", err)
				}
				f.Close()
			}
		})
	}
}

func TestOrganize(t *testing.T) {
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
				if err := os.Chmod(destDir, 0555); err != nil {
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
			err := Organize(params)

			// Validate the outcome
			if (err != nil) != tt.expectError {
				t.Errorf("run() test %q failed. Expected error: %v, got: %v", tt.name, tt.expectError, err)
			}
		})
	}
}
