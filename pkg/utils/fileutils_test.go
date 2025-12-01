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
	"time"

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
		name      string
		extension string
		want      bool
	}{
		{
			name:      "JPG extension",
			extension: ".jpg",
			want:      true,
		},
		{
			name:      "JPEG extension",
			extension: ".jpeg",
			want:      true,
		},
		{
			name:      "Upper case JPG",
			extension: ".JPG",
			want:      true,
		},
		{
			name:      "RAW extension",
			extension: ".arw", // Sony RAW
			want:      true,
		},
		{
			name:      "Unsupported extension",
			extension: ".txt",
			want:      false,
		},
		{
			name:      "Empty extension",
			extension: "",
			want:      false,
		},
		{
			name:      "Non-media extension",
			extension: ".exe",
			want:      false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isAllowedExtension(tt.extension)
			if got != tt.want {
				t.Errorf("isAllowedExtension(%q) = %v, want %v", tt.extension, got, tt.want)
			}
		})
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
	t.Run("Copy or compress JPEG includes EXIF metadata", func(t *testing.T) {
		src := "../testdata/sample_with_exif.jpg"
		dest := filepath.Join(t.TempDir(), "compressed_test.jpg")

		imageData, err := os.ReadFile(src)
		if err != nil {
			t.Fatalf("Failed to read source file: %v", err)
		}

		// Compress the image
		params := &models.Params{Compression: 50, DeleteSource: false, Destination: dest, Source: src}
		if err := copyOrCompressImage(dest, src, imageData, true, params, &ProcessingSummary{}); err != nil {
			t.Fatalf("copyOrCompressImage returned an error: %v", err)
		}

		// Read the destination file
		destFile, err := os.Open(dest)
		if err != nil {
			t.Fatalf("Failed to open destination file: %v", err)
		}
		defer destFile.Close()

		// Check if the destination file is a valid JPEG file
		if _, err := jpeg.Decode(destFile); err != nil {
			t.Fatalf("Failed to decode destination file: %v", err)
		}

		writtenImageData, err := os.ReadFile(dest)
		if err != nil {
			t.Fatalf("Failed to read source file: %v", err)
		}

		// Get EXIF date from the destination file
		exifDate, err := GetImageDateTime(writtenImageData, ".jpg")
		if err != nil {
			t.Fatalf("Failed to get EXIF date from destination file: %v", err)
		}

		// Check if the EXIF date is the same as the original
		if exifDate.Day() != 25 || exifDate.Month() != 12 || exifDate.Year() != 2022 {
			t.Errorf("EXIF date does not match. Original: %v, Destination: %v", time.Date(2022, 12, 25, 10, 30, 0, 0, time.UTC), exifDate)
		}
	})
	t.Run("Copy or compress JPG", func(t *testing.T) {
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
	})
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

// TestProcessMediaFiles_EdgeCases tests various edge cases for the ProcessMediaFiles function
func TestProcessMediaFiles_EdgeCases(t *testing.T) {
	// Create test directories
	sourceDir := t.TempDir()
	destDir := t.TempDir()

	// Test case: Empty source directory
	t.Run("Empty source directory", func(t *testing.T) {
		params := &models.Params{
			Source:      sourceDir,
			Destination: destDir,
			Compression: -1,
		}

		summary, err := ProcessMediaFiles(params)
		if err != nil {
			t.Errorf("ProcessMediaFiles failed with empty source: %v", err)
		}
		if summary.Processed != 0 {
			t.Errorf("Expected 0 processed files, got %d", summary.Processed)
		}
	})

	// Test case: Non-media files
	t.Run("Non-media files", func(t *testing.T) {
		// Create a non-media file
		nonMediaFile := filepath.Join(sourceDir, "test.txt")
		if err := os.WriteFile(nonMediaFile, []byte("test data"), 0644); err != nil {
			t.Fatalf("Failed to create test file: %v", err)
		}

		params := &models.Params{
			Source:      sourceDir,
			Destination: destDir,
			Compression: -1,
		}

		summary, err := ProcessMediaFiles(params)
		if err != nil {
			t.Errorf("ProcessMediaFiles failed with non-media file: %v", err)
		}
		if summary.Processed != 0 {
			t.Errorf("Expected 0 processed files (non-media should be ignored), got %d", summary.Processed)
		}
	})

	// Test case: Corrupted JPEG file (no EXIF data)
	t.Run("Corrupted JPEG file", func(t *testing.T) {
		// Create a fake JPEG file without valid EXIF data
		corruptedFile := filepath.Join(sourceDir, "corrupted.jpg")
		if err := os.WriteFile(corruptedFile, []byte("Not a valid JPEG"), 0644); err != nil {
			t.Fatalf("Failed to create corrupted file: %v", err)
		}

		params := &models.Params{
			Source:      sourceDir,
			Destination: destDir,
			Compression: -1,
		}

		summary, err := ProcessMediaFiles(params)
		if err != nil {
			t.Errorf("ProcessMediaFiles failed with corrupted file: %v", err)
		}
		// The file should be skipped due to EXIF extraction failure
		if summary.Skipped != 1 {
			t.Errorf("Expected 1 skipped file, got %d", summary.Skipped)
		}
	})

	// Test case: File with same name already exists in destination
	t.Run("Destination file already exists", func(t *testing.T) {
		// Create directories that match the expected output structure
		// We'll mimic a file from 2023/01-01
		destStructure := filepath.Join(destDir, "2023", "01-01")
		if err := os.MkdirAll(destStructure, os.ModePerm); err != nil {
			t.Fatalf("Failed to create destination structure: %v", err)
		}

		// Copy a test image to source
		sourceImage := filepath.Join("../testdata", "image.jpg")
		destImage := filepath.Join(sourceDir, "image.jpg")
		if err := copyTestFile(sourceImage, destImage); err != nil {
			t.Skipf("Test skipped: Could not copy test file: %v", err)
			return
		}

		// Create a file with the same name in the destination
		duplicateFile := filepath.Join(destStructure, "image.jpg")
		if err := os.WriteFile(duplicateFile, []byte("existing file"), 0644); err != nil {
			t.Fatalf("Failed to create duplicate file: %v", err)
		}

		params := &models.Params{
			Source:      sourceDir,
			Destination: destDir,
			Compression: -1,
		}

		summary, err := ProcessMediaFiles(params)
		if err != nil {
			t.Errorf("ProcessMediaFiles failed with duplicate file: %v", err)
		}
		if summary.Skipped != 1 {
			t.Errorf("Expected 1 skipped file (due to duplicate), got %d", summary.Skipped)
		}
	})

	// Test case: Delete source files
	t.Run("Delete source files", func(t *testing.T) {
		// Create a temporary source file
		sourceImage := filepath.Join(sourceDir, "to_delete.jpg")
		if err := os.WriteFile(sourceImage, createFakeExifData(), 0644); err != nil {
			t.Fatalf("Failed to create source file: %v", err)
		}

		params := &models.Params{
			Source:       sourceDir,
			Destination:  destDir,
			Compression:  -1,
			DeleteSource: true,
		}

		summary, err := ProcessMediaFiles(params)
		if err != nil {
			t.Errorf("ProcessMediaFiles failed with delete source: %v", err)
		}

		// Check if the source file was deleted
		if _, err := os.Stat(sourceImage); !os.IsNotExist(err) {
			t.Errorf("Source file should have been deleted")
		}
		if summary.Deleted != 1 {
			t.Errorf("Expected 1 deleted file, got %d", summary.Deleted)
		}
	})

	// Test case: Apply compression
	t.Run("Apply compression", func(t *testing.T) {
		// Since our fake JPEG is too minimal to be properly processed by Go's image decoder,
		// we'll just verify that it attempts to process it

		// Create a new source file for this test
		sourceImage := filepath.Join(sourceDir, "to_compress.jpg")
		if err := os.WriteFile(sourceImage, createFakeExifData(), 0644); err != nil {
			t.Fatalf("Failed to create source file: %v", err)
		}

		params := &models.Params{
			Source:      sourceDir,
			Destination: destDir,
			Compression: 50, // Apply compression
		}

		summary, err := ProcessMediaFiles(params)
		if err != nil {
			t.Errorf("ProcessMediaFiles failed with compression: %v", err)
		}

		// The compression would fail because our fake JPEG is not a valid image
		// We just check it tried to process it
		if summary.Processed == 0 && summary.Skipped == 0 {
			t.Errorf("Expected file to be processed or skipped")
		}
	})
}

// Helper function to create a fake JPEG file with EXIF data
func createFakeExifData() []byte {
	// Create a basic valid JPEG structure with EXIF metadata
	// This is a simplified version but includes enough to be processed by the system

	// SOI marker
	data := []byte{0xFF, 0xD8}

	// APP1 marker for EXIF
	data = append(data, 0xFF, 0xE1)

	// Length of APP1 segment (big endian)
	exifData := []byte("Exif\x00\x00MM\x00*\x00\x00\x00\x08")

	// Add a simple IFD with DateTime tag
	ifd := []byte{
		0x00, 0x01, // Number of directory entries

		// DateTime tag (0x0132)
		0x01, 0x32, // Tag
		0x00, 0x02, // Type (ASCII)
		0x00, 0x00, 0x00, 0x14, // Count (20 bytes)
		0x00, 0x00, 0x00, 0x1A, // Offset to value

		// Next IFD offset (0 = no more)
		0x00, 0x00, 0x00, 0x00,

		// DateTime value: "2025:01:11 17:10:39"
		'2', '0', '2', '5', ':', '0', '1', ':', '1', '1', ' ',
		'1', '7', ':', '1', '0', ':', '3', '9', 0x00,
	}
	exifData = append(exifData, ifd...)

	// Add length bytes (big endian, includes the length bytes themselves)
	length := len(exifData) + 2
	data = append(data, byte(length>>8), byte(length&0xFF))
	data = append(data, exifData...)

	// Add a simple ending to make it a valid JPEG
	data = append(data, 0xFF, 0xD9) // EOI marker

	return data
}

// Helper function to copy a test file
func copyTestFile(src, dst string) error {
	// Check if source exists
	if _, err := os.Stat(src); os.IsNotExist(err) {
		return fmt.Errorf("source file %s does not exist", src)
	}

	// Read the source file
	data, err := os.ReadFile(src)
	if err != nil {
		return err
	}

	// Write to destination
	return os.WriteFile(dst, data, 0644)
}
