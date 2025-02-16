package utils

import (
	"bytes"
	"image"
	"image/jpeg"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/matdmb/organize-media/pkg/models"
)

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
		wantErr     bool
		wantSummary ProcessingSummary
	}{
		{
			name: "Process files with compression",
			params: &models.Params{
				Source:       "../testdata/exif/sample_with_exif.jpg",
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
				Source:       "../testdata/sony/DSC00001.ARW",
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
			name: "Process files with corrupted EXIF",
			params: &models.Params{
				Source:       "../testdata/exif/sample_corrupted_exif.jpg",
				Destination:  destDir,
				Compression:  50,
				DeleteSource: false,
			},
			wantErr:     true,
			wantSummary: ProcessingSummary{},
		},
		{
			name: "Process files without EXIF",
			params: &models.Params{
				Source:       "../testdata/exif/sample_without_exif.jpg",
				Destination:  destDir,
				Compression:  50,
				DeleteSource: false,
			},
			wantErr:     true,
			wantSummary: ProcessingSummary{},
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
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a new destination directory for each test
			testDestDir := filepath.Join(destDir, tt.name)
			tt.params.Destination = testDestDir

			summary, err := ProcessMediaFiles(tt.params)

			if (err != nil) != tt.wantErr {
				t.Errorf("ProcessMediaFiles() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				if summary != tt.wantSummary {
					t.Errorf("ProcessMediaFiles() summary = %+v, want %+v", summary, tt.wantSummary)
				}

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
		})
	}
}
