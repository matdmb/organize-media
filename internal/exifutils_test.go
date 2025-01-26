package internal

import (
	"testing"
	"time"
)

func TestGetExifDate(t *testing.T) {
	tests := []struct {
		name         string
		filePath     string
		expectedDate time.Time
		expectError  bool
	}{
		{
			name:         "Valid EXIF metadata",
			filePath:     "testdata/sample_with_exif.jpg",
			expectedDate: time.Date(2022, 12, 25, 10, 30, 0, 0, time.UTC),
			expectError:  false,
		},
		{
			name:         "File without EXIF metadata",
			filePath:     "testdata/sample_without_exif.jpg",
			expectedDate: time.Time{},
			expectError:  true,
		},
		{
			name:         "Nonexistent file",
			filePath:     "testdata/nonexistent.jpg",
			expectedDate: time.Time{},
			expectError:  true,
		},
		{
			name:         "Corrupted file",
			filePath:     "testdata/sample_corrupted_exif.jpg",
			expectedDate: time.Time{},
			expectError:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Call the function with the provided file path
			date, err := GetExifDate(tt.filePath)

			// Define the expected date in the local time zone
			//location := date.Location()
			//tt.expectedDate = tt.expectedDate.In(location)

			// Check for expected error
			if (err != nil) != tt.expectError {
				t.Fatalf("Test %q: Expected error: %v, got: %v", tt.name, tt.expectError, err)
			}

			// Check for expected date
			if !tt.expectError && !date.Equal(tt.expectedDate) {
				t.Errorf("Test %q: Expected date: %v, got: %v", tt.name, tt.expectedDate, date)
			}
		})
	}
}
