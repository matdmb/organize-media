package internal

import (
	"errors"
	"os"
	"time"

	"github.com/rwcarlsen/goexif/exif"
)

// GetExifDate retrieves the EXIF date from a file using the provided decode and dateTime functions.
func GetExifDate(
	filePath string,
	decodeFunc func(f *os.File) (*exif.Exif, error),
	dateTimeFunc func(exifData *exif.Exif) (time.Time, error),
) (time.Time, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return time.Time{}, err
	}
	defer file.Close()

	exifData, err := decodeFunc(file)
	if err != nil {
		return time.Time{}, errors.New("failed to decode EXIF data")
	}

	date, err := dateTimeFunc(exifData)
	if err != nil {
		return time.Time{}, errors.New("failed to retrieve EXIF date")
	}

	return date, nil
}

// DefaultGetExifDate reads the EXIF metadata and extracts the DateTime field.
func DefaultGetExifDate(path string) (time.Time, error) {
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

func decodeExifFromFile(f *os.File) (*exif.Exif, error) {
	return exif.Decode(f)
}
