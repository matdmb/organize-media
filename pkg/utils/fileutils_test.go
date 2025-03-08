package utils

import (
	"bytes"
	"fmt"
	"image"
	"image/jpeg"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/matdmb/organize-media/pkg/models"
)

// Helper function to handle platform-specific test expectations
func getSkippedCount() int {
	// On Windows, permission tests behave differently
	if runtime.GOOS == "windows" {
		return 0
	}
	return 1
}

func TestIsAllowedExtension(t *testing.T) {
	tests := []struct {
		ext      string
		expected bool
	}{
		{".jpg", true},
		{".JPG", true}, // Test case-insensitivity
		{".nef", true},
		{".png", false},
		{".txt", false},
	}

	for _, test := range tests {
		result := isAllowedExtension(test.ext)
		if result != test.expected {
			t.Errorf("isAllowedExtension(%s) = %v; want %v", test.ext, result, test.expected)
		}
	}
}

func TestCountFiles(t *testing.T) {
	tempDir := t.TempDir()
	allowedFile := filepath.Join(tempDir, "test.jpg")
	disallowedFile := filepath.Join(tempDir, "test.txt")

	// Create temporary files
	os.WriteFile(allowedFile, []byte{}, 0644)
	os.WriteFile(disallowedFile, []byte{}, 0644)

	count, _, err := CountFiles(tempDir)
	if err != nil {
		t.Fatalf("CountFiles returned an error: %v", err)
	}

	if count != 1 {
		t.Errorf("CountFiles returned %d; want 1", count)
	}
}

func TestFileExists(t *testing.T) {
	// Create temporary directory for tests
	tempDir := t.TempDir()

	// Create a test file
	testFile := filepath.Join(tempDir, "exists.txt")
	if err := os.WriteFile(testFile, []byte("test"), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	t.Run("existing file", func(t *testing.T) {
		exists, err := fileExists(testFile)
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}
		if !exists {
			t.Error("Expected file to exist, but got false")
		}
	})

	t.Run("non-existent file", func(t *testing.T) {
		exists, err := fileExists(filepath.Join(tempDir, "nonexistent.txt"))
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}
		if exists {
			t.Error("Expected file to not exist, but got true")
		}
	})

	t.Run("no permission", func(t *testing.T) {
		// Skip on Windows as permission tests behave differently
		if runtime.GOOS == "windows" {
			t.Skip("Skipping permission test on Windows")
		}

		noPermDir := filepath.Join(tempDir, "noperm")
		if err := os.Mkdir(noPermDir, 0000); err != nil {
			t.Fatalf("Failed to create no-permission directory: %v", err)
		}
		defer os.Chmod(noPermDir, 0700) // Restore permissions for cleanup

		exists, err := fileExists(filepath.Join(noPermDir, "test.txt"))
		if err == nil {
			t.Error("Expected permission error, got nil")
		}
		if exists {
			t.Error("Expected false for exists with permission error")
		}
	})
}

func TestCopyOrCompressImage(t *testing.T) {
	// Create temp dirs for source and destination
	srcDir := t.TempDir()
	destDir := t.TempDir()

	// Create test image
	img := image.NewRGBA(image.Rect(0, 0, 100, 100))
	var imgBuffer bytes.Buffer
	if err := jpeg.Encode(&imgBuffer, img, nil); err != nil {
		t.Fatalf("Failed to create test image: %v", err)
	}
	imageData := imgBuffer.Bytes()

	// Test cases
	tests := []struct {
		name         string
		sourceFile   string
		isJPG        bool
		compression  int
		deleteSource bool
		wantSkipped  bool
		wantError    bool
	}{
		{
			name:         "Compress JPG",
			sourceFile:   filepath.Join(srcDir, "test.jpg"),
			isJPG:        true,
			compression:  50,
			deleteSource: false,
			wantSkipped:  false,
			wantError:    false,
		},
		{
			name:         "Copy non-JPG",
			sourceFile:   filepath.Join(srcDir, "test.nef"),
			isJPG:        false,
			compression:  50,
			deleteSource: false,
			wantSkipped:  false,
			wantError:    false,
		},
		{
			name:         "Skip existing file",
			sourceFile:   filepath.Join(srcDir, "existing.jpg"),
			isJPG:        true,
			compression:  50,
			deleteSource: false,
			wantSkipped:  true,
			wantError:    false,
		},
		{
			name:         "Delete source after copy",
			sourceFile:   filepath.Join(srcDir, "delete.jpg"),
			isJPG:        true,
			compression:  50,
			deleteSource: true,
			wantSkipped:  false,
			wantError:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create source file
			if err := os.WriteFile(tt.sourceFile, imageData, 0644); err != nil {
				t.Fatalf("Failed to create source file: %v", err)
			}

			destPath := filepath.Join(destDir, filepath.Base(tt.sourceFile))

			// For "skip existing" test, create destination file first
			if tt.wantSkipped {
				if err := os.MkdirAll(filepath.Dir(destPath), os.ModePerm); err != nil {
					t.Fatalf("Failed to create destination dir: %v", err)
				}
				if err := os.WriteFile(destPath, []byte("existing"), 0644); err != nil {
					t.Fatalf("Failed to create existing file: %v", err)
				}
			}

			params := &models.Params{
				Compression:  tt.compression,
				DeleteSource: tt.deleteSource,
			}

			var summary ProcessingSummary
			err := copyOrCompressImage(destPath, tt.sourceFile, imageData, tt.isJPG, params, &summary)

			if (err != nil) != tt.wantError {
				t.Errorf("copyOrCompressImage() error = %v, wantError %v", err, tt.wantError)
				return
			}

			if tt.wantSkipped {
				if summary.Skipped != 1 {
					t.Errorf("Expected file to be skipped")
				}
				return
			}

			// Verify file exists at destination
			if _, err := os.Stat(destPath); os.IsNotExist(err) {
				t.Error("Destination file was not created")
			}

			// Verify source file deletion
			if tt.deleteSource {
				if _, err := os.Stat(tt.sourceFile); !os.IsNotExist(err) {
					t.Error("Source file was not deleted")
				}
				if summary.Deleted != 1 {
					t.Error("Deleted count not incremented")
				}
			}

			// Verify compression/copy counters
			if tt.isJPG && tt.compression >= 0 {
				if summary.Compressed != 1 {
					t.Error("Compressed count not incremented for JPG")
				}
			} else {
				if summary.Copied != 1 {
					t.Error("Copied count not incremented for non-JPG")
				}
			}

			if summary.Processed != 1 {
				t.Error("Processed count not incremented")
			}
		})
	}
}

func TestProcessMediaFiles(t *testing.T) {
	// Create temp destination directory only
	destDir := t.TempDir()

	tests := []struct {
		name        string
		params      *models.Params
		setupFunc   func(string) string // Function to set up test environment
		wantErr     bool
		wantSummary ProcessingSummary
	}{
		{
			name: "Process files with compression",
			params: &models.Params{
				Source:       "../testdata/DSC00001.JPG",
				Destination:  destDir,
				Compression:  50,
				DeleteSource: false,
			},
			wantErr: false,
			wantSummary: ProcessingSummary{
				Processed:  1,
				Compressed: 1,
				Copied:     0,
				Skipped:    0,
				Deleted:    0,
			},
		},
		{
			name: "Process RAW file",
			params: &models.Params{
				Source:       "../testdata/DSC00001.ARW",
				Destination:  destDir,
				Compression:  -1,
				DeleteSource: false,
			},
			wantErr: false,
			wantSummary: ProcessingSummary{
				Processed:  1,
				Compressed: 0,
				Copied:     1,
				Skipped:    0,
				Deleted:    0,
			},
		},
		{
			name: "Invalid source directory",
			params: &models.Params{
				Source:       "/nonexistent",
				Destination:  destDir,
				Compression:  50,
				DeleteSource: false,
			},
			wantErr:     true,
			wantSummary: ProcessingSummary{},
		},
		// Add test for file with no read permissions (to test error opening file)
		{
			name: "File with no read permissions",
			params: &models.Params{
				Destination:  destDir,
				Compression:  50,
				DeleteSource: false,
				// Source set in setupFunc
			},
			setupFunc: func(destDir string) string {
				// Create a temporary file with no read permissions
				tempDir := t.TempDir()
				filePath := filepath.Join(tempDir, "noperm.jpg")

				// First create the file normally
				err := os.WriteFile(filePath, []byte("test data"), 0644)
				if err != nil {
					t.Fatalf("Failed to create test file: %v", err)
				}

				// Then remove read permissions
				if runtime.GOOS != "windows" {
					if err := os.Chmod(filePath, 0200); err != nil {
						t.Fatalf("Failed to change file permissions: %v", err)
					}
				}

				return tempDir
			},
			wantErr: false, // Not expecting error at top level, just skipping file
			wantSummary: ProcessingSummary{
				Processed:  0,
				Compressed: 0,
				Copied:     0,
				Skipped:    getSkippedCount(), // Use a function instead of ternary
				Deleted:    0,
			},
		},
		// Test multiple files in a directory
		{
			name: "Multiple files in directory",
			params: &models.Params{
				Destination:  destDir,
				Compression:  50,
				DeleteSource: false,
				// Source set in setupFunc
			},
			setupFunc: func(destDir string) string {
				// Create a temporary directory with multiple files
				tempDir := t.TempDir()

				// Copy a sample file multiple times
				samplePath := "../testdata/DSC00001.JPG"
				sampleData, err := os.ReadFile(samplePath)
				if err != nil {
					t.Fatalf("Failed to read sample file: %v", err)
				}

				// Create 3 copies with different names
				for i := 1; i <= 3; i++ {
					destPath := filepath.Join(tempDir, fmt.Sprintf("test%d.jpg", i))
					if err := os.WriteFile(destPath, sampleData, 0644); err != nil {
						t.Fatalf("Failed to create test file %d: %v", i, err)
					}
				}

				return tempDir
			},
			wantErr: false,
			wantSummary: ProcessingSummary{
				Processed:  3,
				Compressed: 3,
				Copied:     0,
				Skipped:    0,
				Deleted:    0,
			},
		},
		// Fix the destination already exists test
		{
			name: "Destination already exists",
			params: &models.Params{
				Source:       "../testdata/DSC00001.JPG",
				Destination:  destDir,
				Compression:  50,
				DeleteSource: false,
				// Setup will modify destination
			},
			setupFunc: func(destDir string) string {
				// Create the destination directory structure and file, identical to what the production code would create
				yearDir := filepath.Join(destDir, "2025")
				monthDayDir := filepath.Join(yearDir, "01-11") // Format from the date in the sample file

				if err := os.MkdirAll(monthDayDir, 0755); err != nil {
					t.Fatalf("Failed to create test directory structure: %v", err)
				}

				// Create a file at the destination path with the expected EXACT name
				destFilePath := filepath.Join(monthDayDir, "DSC00001.JPG")

				// Copy the original file to the destination to simulate a previously processed file
				srcData, err := os.ReadFile("../testdata/DSC00001.JPG")
				if err != nil {
					t.Fatalf("Failed to read test source file: %v", err)
				}

				if err := os.WriteFile(destFilePath, srcData, 0644); err != nil {
					t.Fatalf("Failed to create test file: %v", err)
				}

				// Return the source file path
				return "../testdata/DSC00001.JPG"
			},
			wantErr: false,
			wantSummary: ProcessingSummary{
				Processed:  0,
				Compressed: 0,
				Copied:     0,
				Skipped:    1, // File exists, so it should be skipped
				Deleted:    0,
			},
		},
		// Test with deleteSource = true
		{
			name: "Delete source after processing",
			params: &models.Params{
				Destination:  destDir,
				Compression:  50,
				DeleteSource: true,
				// Source set in setupFunc
			},
			setupFunc: func(destDir string) string {
				// Create a temporary directory and copy the sample file there
				tempDir := t.TempDir()
				samplePath := "../testdata/DSC00001.JPG"
				sampleData, err := os.ReadFile(samplePath)
				if err != nil {
					t.Fatalf("Failed to read sample file: %v", err)
				}

				destPath := filepath.Join(tempDir, "to_delete.jpg")
				if err := os.WriteFile(destPath, sampleData, 0644); err != nil {
					t.Fatalf("Failed to create test file: %v", err)
				}

				return tempDir
			},
			wantErr: false,
			wantSummary: ProcessingSummary{
				Processed:  1,
				Compressed: 1,
				Copied:     0,
				Skipped:    0,
				Deleted:    1, // Source file should be deleted
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a new destination directory for each test
			testDestDir := filepath.Join(destDir, tt.name)
			if err := os.MkdirAll(testDestDir, 0755); err != nil {
				t.Fatalf("Failed to create test destination directory: %v", err)
			}

			tt.params.Destination = testDestDir

			// Run setup function if provided
			if tt.setupFunc != nil {
				sourcePath := tt.setupFunc(testDestDir)
				tt.params.Source = sourcePath
			}

			summary, err := ProcessMediaFiles(tt.params)

			if (err != nil) != tt.wantErr {
				t.Errorf("ProcessMediaFiles() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				// Compare everything except Duration
				gotSummary := ProcessingSummary{
					Processed:  summary.Processed,
					Compressed: summary.Compressed,
					Copied:     summary.Copied,
					Skipped:    summary.Skipped,
					Deleted:    summary.Deleted,
				}

				if gotSummary != tt.wantSummary {
					t.Errorf("ProcessMediaFiles() summary = %+v, want %+v", gotSummary, tt.wantSummary)
				}

				if tt.wantSummary.Processed > 0 {
					// Verify files were processed
					files, err := filepath.Glob(filepath.Join(testDestDir, "*/*/*.*"))
					if err != nil {
						t.Errorf("Failed to check processed files: %v", err)
					}
					expectedFiles := tt.wantSummary.Processed
					if len(files) != expectedFiles {
						t.Errorf("Expected %d processed files, got %d", expectedFiles, len(files))
					}
				}
			}
		})
	}
}
