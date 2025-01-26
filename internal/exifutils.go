package internal

import (
	"os"
	"time"

	"github.com/rwcarlsen/goexif/exif"
)

// GetExifDate retrieves the EXIF date from a file using the provided decode and dateTime functions.
func GetExifDate(path string) (time.Time, error) {
	file, err := os.Open(path)
	if err != nil {
		return time.Time{}, err
	}
	defer file.Close()

	// Parse EXIF metadata
	exifData, err := exif.Decode(file)
	if err != nil {
		return time.Time{}, err
	}

	// Extract DateTime field
	date, err := exifData.DateTime()
	if err != nil {
		return time.Time{}, err
	}

	return date, nil
}
