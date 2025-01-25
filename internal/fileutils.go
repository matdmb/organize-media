package internal

import (
	"fmt"
	"image"
	"image/jpeg"
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
func ListFilesWithExif(directory string, getExifDate func(string, func(string) ([]byte, error), func([]byte) (time.Time, error)) (time.Time, error)) ([]ImageFile, error) {
	var files []ImageFile

	err := filepath.Walk(directory, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}

		// Check for allowed extensions
		ext := filepath.Ext(info.Name())
		isAllowed := false
		for _, allowed := range allowedExtensions {
			if ext == allowed {
				isAllowed = true
				break
			}
		}
		if !isAllowed {
			return nil
		}

		// Extract EXIF date
		date, err := getExifDate(path, nil, func(data []byte) (time.Time, error) {
			// Return a mock or parsed date
			return time.Now(), nil
		})
		if err != nil {
			// Skip files without EXIF data
			return nil
		}

		files = append(files, ImageFile{
			Path: path,
			Date: date,
		})
		return nil
	})

	if err != nil {
		return nil, err
	}

	return files, nil
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
			return fmt.Errorf("failed to create directory %s: %w", monthDayDir, err)
		}

		newPath := filepath.Join(monthDayDir, filepath.Base(file.Path))

		// Check if the file already exists at the destination
		if _, err := os.Stat(newPath); err == nil {
			// File exists, skipping
			fmt.Printf("File already exists: %s, skipping...\n", newPath)
			continue
		} else if !os.IsNotExist(err) {
			// Another error occurred
			return fmt.Errorf("error checking destination file %s: %w", newPath, err)
		}

		// Compress and move JPG files if compression is enabled
		if strings.ToLower(filepath.Ext(file.Path)) == ".jpg" && compression >= 0 {
			if err := compressAndMoveJPG(file.Path, newPath, compression); err != nil {
				return err
			}

			// Delete the original file after successful compression
			if err := os.Remove(file.Path); err != nil {
				return fmt.Errorf("failed to delete original file %s: %v", file.Path, err)
			}
		} else {
			// Move the file without compression
			if err := os.Rename(file.Path, newPath); err != nil {
				return fmt.Errorf("failed to move file %s to %s: %w", file.Path, newPath, err)
			}
			fmt.Printf("Moved file %s to %s\n", file.Path, newPath)
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
	fmt.Printf("Compressed file %s created from %s\n", dest, src)
	return nil
}
