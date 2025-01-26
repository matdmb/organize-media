package internal

import (
	"image"
	"image/color"
	"image/jpeg"
	"os"
	"path/filepath"
	"testing"
	"time"
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

	count, err := CountFiles(tempDir)
	if err != nil {
		t.Fatalf("CountFiles returned an error: %v", err)
	}

	if count != 1 {
		t.Errorf("CountFiles returned %d; want 1", count)
	}
}

func TestCompressAndMoveJPG(t *testing.T) {
	src := filepath.Join(t.TempDir(), "test.jpg")
	dest := filepath.Join(t.TempDir(), "compressed_test.jpg")

	// Create a temporary JPG file
	file, _ := os.Create(src)
	img := image.NewRGBA(image.Rect(0, 0, 100, 100))
	jpeg.Encode(file, img, nil)
	file.Close()

	err := compressAndMoveJPG(src, dest, 50)
	if err != nil {
		t.Fatalf("compressAndMoveJPG returned an error: %v", err)
	}

	// Check if destination file exists
	if _, err := os.Stat(dest); os.IsNotExist(err) {
		t.Errorf("Destination file %s was not created", dest)
	}
}

func TestMoveFiles(t *testing.T) {
	// Step 1: Create temporary source and destination directories
	srcDir := t.TempDir()
	destDir := t.TempDir()

	// Step 2: Prepare mock files with specific dates
	mockFiles := []ImageFile{
		{
			Path: filepath.Join(srcDir, "image1.jpg"),
			Date: time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC),
		},
		{
			Path: filepath.Join(srcDir, "image2.jpg"),
			Date: time.Date(2023, 1, 2, 12, 0, 0, 0, time.UTC),
		},
	}

	// Create valid JPEG mock files
	for _, file := range mockFiles {
		err := createMockJPEG(file.Path)
		if err != nil {
			t.Fatalf("Failed to create mock JPEG file: %v", err)
		}
	}

	// Step 3: Call the MoveFiles function
	summary, err := MoveFiles(mockFiles, destDir, 80) // Compression level set to 80
	if err != nil {
		t.Fatalf("MoveFiles failed: %v", err)
	}

	// Step 4: Verify the results
	if summary.Moved != 0 || summary.Compressed != 2 || summary.Skipped != 0 {
		t.Errorf("Unexpected summary: %+v", summary)
	}

	// Verify files are moved and compressed
	for _, file := range mockFiles {
		destPath := filepath.Join(destDir,
			file.Date.Format("2006"),
			file.Date.Format("01-02"),
			filepath.Base(file.Path),
		)

		if _, err := os.Stat(destPath); os.IsNotExist(err) {
			t.Errorf("File not found at destination: %s", destPath)
		}
	}
}

func TestMoveFilesWithoutCompression(t *testing.T) {
	// Step 1: Create temporary source and destination directories
	srcDir := t.TempDir()
	destDir := t.TempDir()

	// Step 2: Prepare mock files with specific dates
	mockFiles := []ImageFile{
		{
			Path: filepath.Join(srcDir, "image1.jpg"),
			Date: time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC),
		},
		{
			Path: filepath.Join(srcDir, "image2.jpg"),
			Date: time.Date(2023, 1, 2, 12, 0, 0, 0, time.UTC),
		},
	}

	// Create valid JPEG mock files
	for _, file := range mockFiles {
		err := createMockJPEG(file.Path)
		if err != nil {
			t.Fatalf("Failed to create mock JPEG file: %v", err)
		}
	}

	// Step 3: Call the MoveFiles function with compression set to -1 (no compression)
	summary, err := MoveFiles(mockFiles, destDir, -1)
	if err != nil {
		t.Fatalf("MoveFiles failed: %v", err)
	}

	// Step 4: Verify the results
	if summary.Moved != 2 || summary.Compressed != 0 || summary.Skipped != 0 {
		t.Errorf("Unexpected summary: %+v", summary)
	}

	// Verify files are moved without compression
	for _, file := range mockFiles {
		destPath := filepath.Join(destDir,
			file.Date.Format("2006"),
			file.Date.Format("01-02"),
			filepath.Base(file.Path),
		)

		// Check if the file exists at the destination
		if _, err := os.Stat(destPath); os.IsNotExist(err) {
			t.Errorf("File not found at destination: %s", destPath)
		}

		// Verify the file size is unchanged
		origInfo, _ := os.Stat(file.Path)
		destInfo, _ := os.Stat(destPath)
		if origInfo != nil && destInfo != nil && origInfo.Size() != destInfo.Size() {
			t.Errorf("File size mismatch: original=%d, moved=%d", origInfo.Size(), destInfo.Size())
		}
	}
}

// createMockJPEG creates a valid JPEG file at the specified path
func createMockJPEG(path string) error {
	img := image.NewRGBA(image.Rect(0, 0, 100, 100))
	for x := 0; x < 100; x++ {
		for y := 0; y < 100; y++ {
			img.Set(x, y, color.RGBA{uint8(x), uint8(y), 255, 255})
		}
	}

	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()

	return jpeg.Encode(file, img, nil)
}

func TestMoveFilesWithNoFiles(t *testing.T) {
	destDir := t.TempDir()

	// Call MoveFiles with an empty slice
	summary, err := MoveFiles([]ImageFile{}, destDir, -1)
	if err != nil {
		t.Fatalf("MoveFiles failed: %v", err)
	}

	// Verify no files were processed
	if summary.Moved != 0 || summary.Compressed != 0 || summary.Skipped != 0 {
		t.Errorf("Unexpected summary: %+v", summary)
	}
}

func TestMoveFilesWithExistingFiles(t *testing.T) {
	srcDir := t.TempDir()
	destDir := t.TempDir()

	mockFile := ImageFile{
		Path: filepath.Join(srcDir, "image1.jpg"),
		Date: time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC),
	}

	// Create the source file
	err := createMockJPEG(mockFile.Path)
	if err != nil {
		t.Fatalf("Failed to create mock JPEG file: %v", err)
	}

	// Create the destination file to simulate an existing file
	destPath := filepath.Join(destDir,
		mockFile.Date.Format("2006"),
		mockFile.Date.Format("01-02"),
		filepath.Base(mockFile.Path),
	)
	if err := os.MkdirAll(filepath.Dir(destPath), os.ModePerm); err != nil {
		t.Fatalf("Failed to create destination directory: %v", err)
	}
	if err := createMockJPEG(destPath); err != nil {
		t.Fatalf("Failed to create existing destination file: %v", err)
	}

	// Call MoveFiles
	summary, err := MoveFiles([]ImageFile{mockFile}, destDir, -1)
	if err != nil {
		t.Fatalf("MoveFiles failed: %v", err)
	}

	// Verify the file was skipped
	if summary.Moved != 0 || summary.Compressed != 0 || summary.Skipped != 1 {
		t.Errorf("Unexpected summary: %+v", summary)
	}
}

func TestMoveFilesWithSubdirectories(t *testing.T) {
	srcDir := t.TempDir()
	destDir := t.TempDir()

	nestedDir := filepath.Join(srcDir, "subdir")
	if err := os.MkdirAll(nestedDir, os.ModePerm); err != nil {
		t.Fatalf("Failed to create nested directory: %v", err)
	}

	mockFile := ImageFile{
		Path: filepath.Join(nestedDir, "image1.jpg"),
		Date: time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC),
	}

	// Create the source file in the nested directory
	err := createMockJPEG(mockFile.Path)
	if err != nil {
		t.Fatalf("Failed to create mock JPEG file: %v", err)
	}

	// Call MoveFiles
	summary, err := MoveFiles([]ImageFile{mockFile}, destDir, -1)
	if err != nil {
		t.Fatalf("MoveFiles failed: %v", err)
	}

	// Verify the file was moved
	if summary.Moved != 1 || summary.Compressed != 0 || summary.Skipped != 0 {
		t.Errorf("Unexpected summary: %+v", summary)
	}
}

func TestMoveFilesWithCompressionError(t *testing.T) {
	srcDir := t.TempDir()
	destDir := t.TempDir()

	mockFile := ImageFile{
		Path: filepath.Join(srcDir, "corrupted.jpg"),
		Date: time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC),
	}

	// Create a corrupted file that cannot be decoded as a JPEG
	if err := os.WriteFile(mockFile.Path, []byte("invalid jpeg data"), 0644); err != nil {
		t.Fatalf("Failed to create corrupted file: %v", err)
	}

	// Call MoveFiles
	_, err := MoveFiles([]ImageFile{mockFile}, destDir, 80)
	if err == nil {
		t.Fatalf("Expected an error but got none")
	}
}

func TestFileRemovalAfterProcessing(t *testing.T) {
	// Step 1: Create temporary source and destination directories
	srcDir := t.TempDir()
	destDir := t.TempDir()

	// Step 2: Prepare mock files with specific dates
	mockFiles := []ImageFile{
		{
			Path: filepath.Join(srcDir, "image1.jpg"),
			Date: time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC),
		},
		{
			Path: filepath.Join(srcDir, "image2.jpg"),
			Date: time.Date(2023, 1, 2, 12, 0, 0, 0, time.UTC),
		},
	}

	// Create valid JPEG mock files in the source directory
	for _, file := range mockFiles {
		err := createMockJPEG(file.Path)
		if err != nil {
			t.Fatalf("Failed to create mock JPEG file: %v", err)
		}
	}

	// Step 3: Call MoveFiles with compression enabled
	_, err := MoveFiles(mockFiles, destDir, 80)
	if err != nil {
		t.Fatalf("MoveFiles failed: %v", err)
	}

	// Step 4: Verify the files have been removed from the source directory
	for _, file := range mockFiles {
		if _, err := os.Stat(file.Path); !os.IsNotExist(err) {
			t.Errorf("Source file was not removed: %s", file.Path)
		}
	}
}
