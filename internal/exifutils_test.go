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
	// Test case 1: Valid EXIF metadata
	sampleImagePath := "testdata/sample_with_exif.jpg"
	if _, err := os.Stat(sampleImagePath); os.IsNotExist(err) {
		t.Fatalf("Sample image with EXIF metadata not found: %v", err)
	}

	extractedDate, err := DefaultGetExifDate(sampleImagePath)
	if err != nil {
		t.Fatalf("DefaultGetExifDate returned an error: %v", err)
	}

	// Log the extracted date for debugging
	t.Logf("Extracted date: %v", extractedDate)
	t.Logf("Extracted date in UTC: %v", extractedDate.UTC())

	// Define the expected date in the local time zone
	location := extractedDate.Location()
	expectedDate := time.Date(2022, 12, 25, 10, 30, 0, 0, location)

	// Compare the extracted date to the normalized expected date
	if !extractedDate.Equal(expectedDate) {
		t.Errorf("Expected date %v, got %v", expectedDate, extractedDate)
	}

	// Test case 2: File does not exist
	nonExistentPath := "testdata/non_existent_file.jpg"
	_, err = DefaultGetExifDate(nonExistentPath)
	if err == nil {
		t.Errorf("Expected an error for non-existent file, but got none")
	}

	// Test case 3: File without EXIF metadata
	fileWithoutExifPath := "testdata/sample_without_exif.jpg"
	/*if _, err := os.Create(fileWithoutExifPath); err != nil {
		t.Fatalf("Failed to create sample file without EXIF metadata: %v", err)
	}
	//defer os.Remove(fileWithoutExifPath) // Clean up after test
	*/
	_, err = DefaultGetExifDate(fileWithoutExifPath)
	if err == nil {
		t.Errorf("Expected an error for file without EXIF metadata, but got none")
	}

	// Test case 4: Corrupted or invalid EXIF file
	corruptedFilePath := "testdata/sample_corrupted_exif.jpg"
	/*	if _, err := os.Create(corruptedFilePath); err != nil {
			t.Fatalf("Failed to create corrupted EXIF file: %v", err)
		}
		defer os.Remove(corruptedFilePath) // Clean up after test
	*/
	_, err = DefaultGetExifDate(corruptedFilePath)
	if err == nil {
		t.Errorf("Expected an error for corrupted EXIF file, but got none")
	}
}
