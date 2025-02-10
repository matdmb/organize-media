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

	count, _, err := CountFiles(tempDir)
	if err != nil {
		t.Fatalf("CountFiles returned an error: %v", err)
	}

	if count != 1 {
		t.Errorf("CountFiles returned %d; want 1", count)
	}
}

func TestCompressImage(t *testing.T) {
	src := filepath.Join(t.TempDir(), "test.jpg")
	dest := filepath.Join(t.TempDir(), "compressed_test.jpg")

	// Create a temporary JPG file
	file, _ := os.Create(src)
	img := image.NewRGBA(image.Rect(0, 0, 100, 100))
	jpeg.Encode(file, img, nil)
	file.Close()

	err := compressImage(src, dest, 50)
	if err != nil {
		t.Fatalf("compressImage returned an error: %v", err)
	}

	// Check if destination file exists
	if _, err := os.Stat(dest); os.IsNotExist(err) {
		t.Errorf("Destination file %s was not created", dest)
	}
}

func TestProcessFiles(t *testing.T) {
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

	// Step 3: Call the ProcessFiles function
	summary, err := ProcessFiles(mockFiles, destDir, 80, true) // Compression level set to 80
	if err != nil {
		t.Fatalf("ProcessFiles failed: %v", err)
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

func TestProcessFilesWithoutCompression(t *testing.T) {
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

	// Step 3: Call the ProcessFiles function with compression set to -1 (no compression)
	summary, err := ProcessFiles(mockFiles, destDir, -1, true)
	if err != nil {
		t.Fatalf("ProcessFiles failed: %v", err)
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

func TestProcessFilesWithNoFiles(t *testing.T) {
	destDir := t.TempDir()

	// Call ProcessFiles with an empty slice
	summary, err := ProcessFiles([]ImageFile{}, destDir, -1, true)
	if err != nil {
		t.Fatalf("ProcessFiles failed: %v", err)
	}

	// Verify no files were processed
	if summary.Moved != 0 || summary.Compressed != 0 || summary.Skipped != 0 {
		t.Errorf("Unexpected summary: %+v", summary)
	}
}

func TestProcessFilesWithExistingFiles(t *testing.T) {
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

	// Call ProcessFiles
	summary, err := ProcessFiles([]ImageFile{mockFile}, destDir, -1, true)
	if err != nil {
		t.Fatalf("ProcessFiles failed: %v", err)
	}

	// Verify the file was skipped
	if summary.Moved != 0 || summary.Compressed != 0 || summary.Skipped != 1 {
		t.Errorf("Unexpected summary: %+v", summary)
	}
}

func TestProcessFilesWithSubdirectories(t *testing.T) {
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

	// Call ProcessFiles
	summary, err := ProcessFiles([]ImageFile{mockFile}, destDir, -1, true)
	if err != nil {
		t.Fatalf("ProcessFiles failed: %v", err)
	}

	// Verify the file was moved
	if summary.Moved != 1 || summary.Compressed != 0 || summary.Skipped != 0 {
		t.Errorf("Unexpected summary: %+v", summary)
	}
}

func TestProcessFilesWithCompressionError(t *testing.T) {
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

	// Call ProcessFiles
	_, err := ProcessFiles([]ImageFile{mockFile}, destDir, 80, true)
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

	// Step 3: Call ProcessFiles with compression enabled
	_, err := ProcessFiles(mockFiles, destDir, 80, true)
	if err != nil {
		t.Fatalf("ProcessFiles failed: %v", err)
	}

	// Step 4: Verify the files have been removed from the source directory
	for _, file := range mockFiles {
		if _, err := os.Stat(file.Path); !os.IsNotExist(err) {
			t.Errorf("Source file was not removed: %s", file.Path)
		}
	}
}

func TestListFilesWithTestdata(t *testing.T) {
	// Step 1: Point to the testdata folder
	testdataDir := "testdata/exif"

	// Step 2: Call ListFiles
	result, err := ListFiles(testdataDir)
	if err != nil {
		t.Fatalf("ListFiles failed: %v", err)
	}

	// Step 3: Verify the results
	expectedPaths := []string{
		filepath.Join(testdataDir, "sample_with_exif.jpg"), // File with valid EXIF
	}

	// Check that the number of files matches the expectation
	if len(result) != len(expectedPaths) {
		t.Errorf("Expected %d files, got %d", len(expectedPaths), len(result))
	}

	// Check that each expected file path exists in the result
	for _, expectedPath := range expectedPaths {
		found := false
		for _, file := range result {
			if file.Path == expectedPath {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Expected file not found in result: %s", expectedPath)
		}
	}
}

func TestListFilesHandlesExifErrors(t *testing.T) {
	testdataDir := "testdata/exif"

	// Call ListFiles
	result, err := ListFiles(testdataDir)
	if err != nil {
		t.Fatalf("ListFiles failed: %v", err)
	}

	// Verify that only valid files are included in the output
	for _, file := range result {
		if file.Path == filepath.Join(testdataDir, "sample_without_exif.jpg") ||
			file.Path == filepath.Join(testdataDir, "sample_corrupted_exif.jpg") {
			t.Errorf("File with invalid EXIF data should not be included: %s", file.Path)
		}
	}

	// Ensure sample_with_exif.jpg is present
	found := false
	for _, file := range result {
		if file.Path == filepath.Join(testdataDir, "sample_with_exif.jpg") {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("Valid file with EXIF data was not included")
	}
}
func TestCopyFile(t *testing.T) {
	// Create temp directories
	srcDir := t.TempDir()
	destDir := t.TempDir()

	// Test successful copy
	t.Run("successful copy", func(t *testing.T) {
		srcPath := filepath.Join(srcDir, "test.txt")
		destPath := filepath.Join(destDir, "test.txt")
		content := []byte("test content")

		// Create source file
		if err := os.WriteFile(srcPath, content, 0644); err != nil {
			t.Fatalf("Failed to create source file: %v", err)
		}

		// Copy file
		err := copyFile(srcPath, destPath)
		if err != nil {
			t.Fatalf("copyFile failed: %v", err)
		}

		// Verify content
		destContent, err := os.ReadFile(destPath)
		if err != nil {
			t.Fatalf("Failed to read destination file: %v", err)
		}
		if string(destContent) != string(content) {
			t.Errorf("Destination content does not match source. Got %s, want %s", destContent, content)
		}
	})

	// Test non-existent source file
	t.Run("non-existent source", func(t *testing.T) {
		err := copyFile(filepath.Join(srcDir, "nonexistent.txt"), filepath.Join(destDir, "test.txt"))
		if err == nil {
			t.Error("Expected error for non-existent source file, got nil")
		}
	})

	// Test invalid destination path
	t.Run("invalid destination", func(t *testing.T) {
		srcPath := filepath.Join(srcDir, "test.txt")
		if err := os.WriteFile(srcPath, []byte("test"), 0644); err != nil {
			t.Fatalf("Failed to create source file: %v", err)
		}

		err := copyFile(srcPath, filepath.Join(destDir, "invalid/path/test.txt"))
		if err == nil {
			t.Error("Expected error for invalid destination path, got nil")
		}
	})
}
