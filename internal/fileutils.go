package internal

import (
	"fmt"
	"image"
	"image/jpeg"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type ImageFile struct {
	Path string
	Date time.Time
}

var allowedExtensions = []string{".jpg", ".nef"}

// ListFiles traverses a directory and returns a slice of ImageFile structs for supported image formats.
func ListFiles(dir string) ([]ImageFile, error) {
	var files []ImageFile

	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		// Check if the file has an allowed extension
		if !info.IsDir() && isAllowedExtension(filepath.Ext(info.Name())) {
			fileDate, err := GetExifDate(path)
			if err != nil {
				log.Printf("Warning: could not get EXIF date for file %s: %v", path, err)
				return nil
			}

			files = append(files, ImageFile{Path: path, Date: fileDate})
		}
		return nil
	})

	return files, err
}

// isAllowedExtension checks if the file extension is in the list of allowed extensions.
func isAllowedExtension(ext string) bool {
	ext = strings.ToLower(ext)
	for _, allowed := range allowedExtensions {
		if ext == allowed {
			return true
		}
	}
	return false
}

// CountFiles counts the number of files with allowed extensions in a directory.
func CountFiles(dir string) (int, error) {
	var count int

	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Increment count for files with allowed extensions
		if !info.IsDir() && isAllowedExtension(filepath.Ext(info.Name())) {
			count++
		}
		return nil
	})

	return count, err
}

// MoveFiles moves image files to a destination directory, creating year/month-day subdirectories.
// If a compression level is specified (0-100), JPG files are compressed before being moved.
func MoveFiles(files []ImageFile, dest string, compression int) error {
	for _, file := range files {
		// Create year and month-day directories
		yearDir := filepath.Join(dest, fmt.Sprintf("%d", file.Date.Year()))
		monthDayDir := filepath.Join(yearDir, fmt.Sprintf("%02d-%02d", file.Date.Month(), file.Date.Day()))

		if err := os.MkdirAll(monthDayDir, os.ModePerm); err != nil {
			return err
		}

		newPath := filepath.Join(monthDayDir, filepath.Base(file.Path))

		// Check if the file already exists at the destination
		if _, err := os.Stat(newPath); err == nil {
			log.Printf("File %s already exists, skipping.", newPath)
			continue
		}

		// Compress and move JPG files if compression is enabled
		if strings.ToLower(filepath.Ext(file.Path)) == ".jpg" && compression >= 0 {
			if err := compressAndMoveJPG(file.Path, newPath, compression); err != nil {
				return err
			}
		} else {
			// Move the file without compression
			if err := os.Rename(file.Path, newPath); err != nil {
				return err
			}
		}

	}

	return nil
}

// compressAndMoveJPG compresses a JPG image to the specified quality level and moves it to the destination.
func compressAndMoveJPG(src, dest string, quality int) error {
	// Open the source image
	srcFile, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("failed to open source file %s: %v", src, err)
	}
	defer srcFile.Close()

	// Decode the image
	img, _, err := image.Decode(srcFile)
	if err != nil {
		return fmt.Errorf("failed to decode image %s: %v", src, err)
	}

	// Create the destination file
	destFile, err := os.Create(dest)
	if err != nil {
		return fmt.Errorf("failed to create destination file %s: %v", dest, err)
	}
	defer destFile.Close()

	// Encode the image with the specified quality
	options := &jpeg.Options{Quality: quality}
	if err := jpeg.Encode(destFile, img, options); err != nil {
		return fmt.Errorf("failed to encode image %s: %v", dest, err)
	}

	return nil
}
