package internal

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestMoveFilesWithExistingFile(t *testing.T) {
	// Create temporary directories for source and destination
	srcDir := t.TempDir()
	destDir := t.TempDir()

	// Create a mock file in the source directory
	srcFilePath := filepath.Join(srcDir, "test.jpg")
	file, err := os.Create(srcFilePath)
	if err != nil {
		t.Fatalf("Failed to create source file: %v", err)
	}
	file.Close()

	// Create the same file in the destination directory to simulate a conflict
	destFilePath := filepath.Join(destDir, "2023", "01-01", "test.jpg")
	if err := os.MkdirAll(filepath.Dir(destFilePath), os.ModePerm); err != nil {
		t.Fatalf("Failed to create destination directory: %v", err)
	}
	destFile, err := os.Create(destFilePath)
	if err != nil {
		t.Fatalf("Failed to create destination file: %v", err)
	}
	destFile.Close()

	// Verify the destination file exists before calling MoveFiles
	destFileInfoBefore, err := os.Stat(destFilePath)
	if err != nil {
		t.Fatalf("Failed to stat destination file before MoveFiles: %v", err)
	}

	// Mock an ImageFile struct
	fileDate := time.Date(2023, time.January, 1, 0, 0, 0, 0, time.UTC)
	files := []ImageFile{
		{Path: srcFilePath, Date: fileDate},
	}

	// Call MoveFiles and ensure it does not overwrite the existing file
	err = MoveFiles(files, destDir, -1)
	if err != nil {
		t.Fatalf("MoveFiles returned an error: %v", err)
	}

	// Verify the source file still exists
	if _, err := os.Stat(srcFilePath); os.IsNotExist(err) {
		t.Errorf("Source file was removed even though destination file existed")
	}

	// Verify the destination file was not overwritten
	destFileInfoAfter, err := os.Stat(destFilePath)
	if err != nil {
		t.Fatalf("Failed to stat destination file after MoveFiles: %v", err)
	}
	if destFileInfoAfter.Size() != destFileInfoBefore.Size() {
		t.Errorf("Destination file was overwritten")
	}

	// Verify the debug output (optional if debug logs are added)
}

func TestCountFiles(t *testing.T) {
	dir := t.TempDir()

	// Create test files
	file1 := filepath.Join(dir, "file1.jpg")
	os.WriteFile(file1, []byte{}, 0644)
	file2 := filepath.Join(dir, "file2.jpg")
	os.WriteFile(file2, []byte{}, 0644)

	count, err := CountFiles(dir)
	if err != nil {
		t.Fatalf("CountFiles returned an error: %v", err)
	}

	if count != 2 {
		t.Errorf("Expected 2 files, got %d", count)
	}
}

func TestMoveFiles(t *testing.T) {
	sourceDir := t.TempDir()
	destDir := t.TempDir()

	// Create test files with mock dates
	file1 := filepath.Join(sourceDir, "file1.jpg")
	os.WriteFile(file1, []byte{}, 0644)
	file2 := filepath.Join(sourceDir, "file2.jpg")
	os.WriteFile(file2, []byte{}, 0644)

	mockFiles := []ImageFile{
		{Path: file1, Date: time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC)},
		{Path: file2, Date: time.Date(2023, 1, 2, 0, 0, 0, 0, time.UTC)},
	}

	err := MoveFiles(mockFiles, destDir, -1)
	if err != nil {
		t.Fatalf("MoveFiles returned an error: %v", err)
	}

	// Check if files moved correctly
	for _, file := range mockFiles {
		newPath := filepath.Join(destDir, file.Date.Format("2006"), file.Date.Format("01-02"), filepath.Base(file.Path))
		if _, err := os.Stat(newPath); os.IsNotExist(err) {
			t.Errorf("File %s was not moved to the correct location", file.Path)
		}
	}
}

func TestCompressAndMoveJPG(t *testing.T) {
	sourceDir := t.TempDir()
	destDir := t.TempDir()

	// Create a test JPG file
	file1 := filepath.Join(sourceDir, "file1.jpg")
	os.WriteFile(file1, []byte{}, 0644)

	err := compressAndMoveJPG(file1, destDir, 75)
	if err == nil {
		t.Error("Expected error for not implemented compressAndMoveJPG, but got none")
	}
}
