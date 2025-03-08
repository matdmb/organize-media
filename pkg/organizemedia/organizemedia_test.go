package organizemedia

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/matdmb/organize-media/pkg/models"
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

// TestOrganizeErrorHandling tests error cases in the Organize function
func TestOrganizeErrorHandling(t *testing.T) {
	// Create test directories
	sourceDir := t.TempDir()
	destDir := t.TempDir()

	// Test cases for directory error handling
	t.Run("Non-existent source directory", func(t *testing.T) {
		nonExistentSource := filepath.Join(sourceDir, "non-existent")

		params := &models.Params{
			Source:        nonExistentSource,
			Destination:   destDir,
			Compression:   -1,
			SkipUserInput: true,
		}

		err := Organize(params)
		if err == nil {
			t.Errorf("Expected error for non-existent source directory, got nil")
		}
	})

	t.Run("Non-existent destination directory", func(t *testing.T) {
		nonExistentDest := filepath.Join(destDir, "non-existent")

		params := &models.Params{
			Source:        sourceDir,
			Destination:   nonExistentDest,
			Compression:   -1,
			SkipUserInput: true,
		}

		err := Organize(params)
		if err == nil {
			t.Errorf("Expected error for non-existent destination directory, got nil")
		}
	})

	t.Run("Invalid compression level", func(t *testing.T) {
		params := &models.Params{
			Source:        sourceDir,
			Destination:   destDir,
			Compression:   101, // Invalid compression level
			SkipUserInput: true,
		}

		err := Organize(params)
		if err == nil {
			t.Errorf("Expected error for invalid compression level, got nil")
		}
	})

	t.Run("Permission denied for destination", func(t *testing.T) {
		// Skip on Windows as permission tests behave differently
		if os.Getenv("GOOS") == "windows" {
			t.Skip("Skipping permission test on Windows")
		}

		// Create a read-only destination directory
		readOnlyDir := filepath.Join(t.TempDir(), "read-only")
		if err := os.Mkdir(readOnlyDir, 0555); err != nil {
			t.Fatalf("Failed to create read-only directory: %v", err)
		}
		// Ensure we have a file to process
		sampleFile := filepath.Join(sourceDir, "test.jpg")
		if err := os.WriteFile(sampleFile, []byte("test data"), 0644); err != nil {
			t.Fatalf("Failed to create test file: %v", err)
		}

		params := &models.Params{
			Source:        sourceDir,
			Destination:   readOnlyDir,
			Compression:   -1,
			SkipUserInput: true,
		}

		err := Organize(params)
		if err == nil {
			t.Errorf("Expected permission error for read-only destination, got nil")
		}
	})

	t.Run("Skip user input", func(t *testing.T) {
		// Create a sample file
		sampleFile := filepath.Join(sourceDir, "test.jpg")
		if err := os.WriteFile(sampleFile, []byte("test data"), 0644); err != nil {
			t.Fatalf("Failed to create test file: %v", err)
		}

		params := &models.Params{
			Source:        sourceDir,
			Destination:   destDir,
			Compression:   -1,
			SkipUserInput: true, // Skip user confirmation
		}

		err := Organize(params)
		if err != nil {
			t.Errorf("Unexpected error with skip user input: %v", err)
		}
		// Note: We can't check the summary directly since it's not returned
	})

	t.Run("Empty source directory", func(t *testing.T) {
		// Use an empty directory
		emptyDir := t.TempDir()

		params := &models.Params{
			Source:        emptyDir,
			Destination:   destDir,
			Compression:   -1,
			SkipUserInput: true,
		}

		err := Organize(params)
		if err == nil {
			t.Errorf("Expected error for empty source directory, got nil")
		}

		// Check if the error message contains the expected message
		if !strings.Contains(err.Error(), "no files to process") {
			t.Errorf("Expected error message to contain 'no files to process', got: %v", err)
		}
	})
}

// createMockErrorFile creates a temporary file that will cause an error when read
func createMockErrorFile(dir string) (string, error) {
	// Create a directory to be mistaken for a file
	filePath := filepath.Join(dir, "error.jpg")
	err := os.Mkdir(filePath, 0755)
	if err != nil {
		return "", fmt.Errorf("Failed to create mock error file: %w", err)
	}
	// Make it read-only for additional error cases
	err = os.Chmod(filePath, 0400)
	if err != nil {
		return "", fmt.Errorf("Failed to set permissions: %w", err)
	}
	return filePath, nil
}

// TestProcessMediaFilesErrors tests error handling in the ProcessMediaFiles function
func TestProcessMediaFilesErrors(t *testing.T) {
	sourceDir := t.TempDir()
	destDir := t.TempDir()

	t.Run("Error reading file", func(t *testing.T) {
		// Skip on Windows as permission tests behave differently
		if os.Getenv("GOOS") == "windows" {
			t.Skip("Skipping permission test on Windows")
		}

		// Create a file that will cause errors when read
		errorFile, err := createMockErrorFile(sourceDir)
		if err != nil {
			t.Fatalf("Failed to setup test: %v", err)
		}

		// Attempt to process the file
		params := &models.Params{
			Source:        sourceDir,
			Destination:   destDir,
			Compression:   -1,
			SkipUserInput: true,
			DeleteSource:  false,
		}

		err = Organize(params)
		// The function should complete but log errors
		if err != nil {
			// This might not actually error out the whole function
			// but we should see files skipped in the summary
			t.Logf("Got error: %v", err)
		}

		// Clean up
		os.Chmod(errorFile, 0700)
		os.RemoveAll(errorFile)
	})
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
	// Create temp destination directory
	sourDir := t.TempDir()
	destDir := t.TempDir()

	tests := []struct {
		name        string
		params      *models.Params
		wantErr     bool
		setupFiles  bool
		errorString string
	}{
		{
			name: "successful organization with sample file",
			params: &models.Params{
				Source:        "../testdata/DSC00001.JPG",
				Destination:   destDir,
				Compression:   80,
				DeleteSource:  false,
				EnableLog:     false,
				SkipUserInput: true,
			},
			wantErr: false,
		},
		{
			name: "successful organization with compression set to -1",
			params: &models.Params{
				Source:        "../testdata/DSC00001.JPG",
				Destination:   destDir,
				Compression:   -1,
				DeleteSource:  false,
				EnableLog:     false,
				SkipUserInput: true,
			},
			wantErr: false,
		},
		/*{
			name: "handle corrupted EXIF data",
			params: &models.Params{
				Source:        "../testdata/exif/sample_corrupted_exif.jpg",
				Destination:   destDir,
				Compression:   80,
				DeleteSource:  false,
				SkipUserInput: true,
			},
			wantErr:     true,
			errorString: "exif: decode failed",
		},
		{
			name: "handle image without EXIF data",
			params: &models.Params{
				Source:        "../testdata/exif/sample_without_exif.jpg",
				Destination:   destDir,
				Compression:   80,
				DeleteSource:  false,
				SkipUserInput: true,
			},
			wantErr:     true,
			errorString: "exif: failed to find exif",
		},
		*/{
			name: "non-existent source directory",
			params: &models.Params{
				Source:        "/non/existent/path",
				Destination:   destDir,
				Compression:   80,
				SkipUserInput: true,
			},
			wantErr:     true,
			errorString: "source directory does not exist",
		},
		{
			name: "invalid compression value",
			params: &models.Params{
				Source:        "../testdata",
				Destination:   destDir,
				Compression:   101,
				SkipUserInput: true,
			},
			wantErr:     true,
			errorString: "compression level must be an integer between 0 and 100",
		},
		{
			name: "empty source directory",
			params: &models.Params{
				Source:        sourDir,
				Destination:   destDir,
				Compression:   50,
				SkipUserInput: true,
			},
			wantErr:     true,
			errorString: "no files to process in source directory",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a unique destination directory for each test
			testDestDir := filepath.Join(destDir, tt.name)
			err := os.MkdirAll(testDestDir, 0755)
			if err != nil {
				t.Fatalf("Failed to create test destination directory: %v", err)
			}
			tt.params.Destination = testDestDir

			err = Organize(tt.params)

			// Check error cases
			if (err != nil) != tt.wantErr {
				t.Errorf("Organize() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr && !strings.Contains(err.Error(), tt.errorString) {
				t.Errorf("Organize() error = %v, want error containing %v", err, tt.errorString)
				return
			}

			// For successful cases, verify files were processed
			if !tt.wantErr {
				// Check if files exist in destination
				files, err := filepath.Glob(filepath.Join(testDestDir, "*/*/*.JPG"))
				if err != nil {
					t.Errorf("Failed to check processed files: %v", err)
				}
				if len(files) == 0 {
					t.Error("No files were processed")
				}

				// Verify directory structure (YYYY/MM-DD)
				for _, file := range files {
					relPath, err := filepath.Rel(testDestDir, file)
					if err != nil {
						t.Errorf("Failed to get relative path: %v", err)
					}
					parts := strings.Split(relPath, string(os.PathSeparator))
					if len(parts) != 3 {
						t.Errorf("Unexpected directory structure: %v", relPath)
					}
				}
			}
		})
	}
}
