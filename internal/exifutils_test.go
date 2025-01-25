package internal

import (
	"errors"
	"os"
	"testing"
	"time"

	"github.com/rwcarlsen/goexif/exif"
)

// Mock decode functions
func mockDecodeSuccess(f *os.File) (*exif.Exif, error) {
	return &exif.Exif{}, nil
}

func mockDecodeFailure(f *os.File) (*exif.Exif, error) {
	return nil, errors.New("no EXIF data")
}

// Mock dateTime functions
func mockDateTimeSuccess(exifData *exif.Exif) (time.Time, error) {
	return time.Date(2023, time.January, 1, 12, 0, 0, 0, time.UTC), nil
}

func mockDateTimeFailure(exifData *exif.Exif) (time.Time, error) {
	return time.Time{}, errors.New("invalid EXIF date")
}

func TestGetExifDate_ValidExifData(t *testing.T) {
	tempFile, err := os.CreateTemp("", "valid_exif.jpg")
	if err != nil {
		t.Fatalf("Failed to create temporary file: %v", err)
	}
	defer os.Remove(tempFile.Name())

	date, err := GetExifDate(tempFile.Name(), mockDecodeSuccess, mockDateTimeSuccess)
	if err != nil {
		t.Errorf("Expected no error, but got: %v", err)
	}

	expectedDate := time.Date(2023, time.January, 1, 12, 0, 0, 0, time.UTC)
	if !date.Equal(expectedDate) {
		t.Errorf("Expected date %v, but got: %v", expectedDate, date)
	}
}

func TestGetExifDate_NoExifData(t *testing.T) {
	tempFile, err := os.CreateTemp("", "no_exif.jpg")
	if err != nil {
		t.Fatalf("Failed to create temporary file: %v", err)
	}
	defer os.Remove(tempFile.Name())

	_, err = GetExifDate(tempFile.Name(), mockDecodeFailure, mockDateTimeSuccess)
	if err == nil {
		t.Error("Expected an error for no EXIF data, but got none")
	}
}

func TestGetExifDate_InvalidExifDate(t *testing.T) {
	tempFile, err := os.CreateTemp("", "invalid_exif.jpg")
	if err != nil {
		t.Fatalf("Failed to create temporary file: %v", err)
	}
	defer os.Remove(tempFile.Name())

	_, err = GetExifDate(tempFile.Name(), mockDecodeSuccess, mockDateTimeFailure)
	if err == nil {
		t.Error("Expected an error for invalid EXIF date, but got none")
	}
}

func TestGetExifDate_FileDoesNotExist(t *testing.T) {
	_, err := GetExifDate("nonexistent.jpg", mockDecodeSuccess, mockDateTimeSuccess)
	if err == nil {
		t.Error("Expected an error for non-existent file, but got none")
	}
}

func TestGetExifDate_WithDecodeExifFromFile(t *testing.T) {
	// Create a temporary file to simulate an image with EXIF data
	tempFile, err := os.CreateTemp("", "test_exif.jpg")
	if err != nil {
		t.Fatalf("Failed to create temporary file: %v", err)
	}
	defer os.Remove(tempFile.Name())

	// Call GetExifDate with decodeExifFromFile
	_, err = GetExifDate(tempFile.Name(), decodeExifFromFile, func(exifData *exif.Exif) (time.Time, error) {
		return exifData.DateTime()
	})

	if err == nil {
		t.Error("Expected an error for missing EXIF data, but got none")
	}
}

func TestDefaultGetExifDate(t *testing.T) {
	// Path to a sample image with EXIF metadata
	sampleImagePath := "testdata/sample_with_exif.jpg"

	// Ensure the sample image exists
	if _, err := os.Stat(sampleImagePath); os.IsNotExist(err) {
		t.Fatalf("Sample image with EXIF metadata not found: %v", err)
	}

	// Call DefaultGetExifDate to extract the date
	extractedDate, err := DefaultGetExifDate(sampleImagePath)
	if err != nil {
		t.Fatalf("DefaultGetExifDate returned an error: %v", err)
	}

	// Define the expected date (match this with the actual EXIF date of your sample image)
	expectedDate := time.Date(2022, 12, 25, 9, 30, 0, 0, time.UTC)
	if !extractedDate.UTC().Equal(expectedDate) {
		t.Errorf("Expected date %v, got %v", expectedDate, extractedDate.UTC())
	}
}
