package internal

import (
	"os"
	"time"

	"github.com/rwcarlsen/goexif/exif"
)

func GetExifDate(path string) (time.Time, error) {
	file, err := os.Open(path)
	if err != nil {
		return time.Time{}, err
	}
	defer file.Close()

	exifData, err := exif.Decode(file)
	if err != nil {
		return time.Time{}, err
	}

	date, err := exifData.DateTime()
	if err != nil {
		return time.Time{}, err
	}

	return date, nil
}
