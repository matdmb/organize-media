package internal

import (
	"image"
	"image/jpeg"
	"os"
	"path/filepath"
	"testing"
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
