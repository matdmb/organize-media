package internal

import (
	"fmt"
	"image"
	"image/jpeg"
	"io"
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

type ProcessingSummary struct {
	Moved      int
	Compressed int
	Skipped    int
}

var allowedExtensions = []string{".jpg", ".nef", ".cr2", "cr3", ".dng", ".arw", "raw"}

// ListFiles traverses a directory and returns a slice of ImageFile structs for supported image formats.
func ListFiles(directory string) ([]ImageFile, error) {
	var files []ImageFile

	err := filepath.Walk(directory, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

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
	ext = strings.ToLower(ext) // Normalize to lowercase
	for _, allowed := range allowedExtensions {
		if ext == allowed {
			return true
		}
	}
	return false
}

// CountFiles counts the number of files with allowed extensions in a directory.
func CountFiles(dir string) (int, int64, error) {
	var count int
	var totalSize int64

	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Increment count for files with allowed extensions
		if !info.IsDir() && isAllowedExtension(filepath.Ext(info.Name())) {
			count++
			totalSize += info.Size()
		}
		return nil
	})

	return count, totalSize, err
}

// ProcessFiles moves or copy image files to a destination directory, creating year/month-day subdirectories.
// If a compression level is specified (0-100), JPG files are compressed before being processed.
func ProcessFiles(files []ImageFile, dest string, compression int, deleteFile bool) (ProcessingSummary, error) {
	var summary ProcessingSummary

	for _, file := range files {
		fmt.Printf("Processing file: %s (Extension: %s)\n", file.Path, filepath.Ext(file.Path))

		targetPath, err := createDestinationPath(dest, file)
		if err != nil {
			return summary, err
		}

		if exists, err := fileExists(targetPath); err != nil {
			return summary, err
		} else if exists {
			fmt.Printf("File already exists: %s, skipping...\n", targetPath)
			summary.Skipped++
			continue
		}

		if err := processFile(file, targetPath, compression, deleteFile, &summary); err != nil {
			return summary, err
		}
	}
	return summary, nil
}

func createDestinationPath(dest string, file ImageFile) (string, error) {
	yearDir := filepath.Join(dest, fmt.Sprintf("%d", file.Date.Year()))
	monthDayDir := filepath.Join(yearDir, fmt.Sprintf("%02d-%02d", file.Date.Month(), file.Date.Day()))

	if err := os.MkdirAll(monthDayDir, os.ModePerm); err != nil {
		return "", fmt.Errorf("failed to create directory %s: %w", monthDayDir, err)
	}

	return filepath.Join(monthDayDir, filepath.Base(file.Path)), nil
}

func fileExists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}

func processFile(file ImageFile, targetPath string, compression int, deleteFile bool, summary *ProcessingSummary) error {
	isJPG := strings.ToLower(filepath.Ext(file.Path)) == ".jpg"
	if isJPG && compression >= 0 {
		return handleJPGFile(file.Path, targetPath, compression, deleteFile, summary)
	}
	return handleNonJPGFile(file.Path, targetPath, deleteFile, summary)
}

func handleJPGFile(srcPath, targetPath string, compression int, deleteFile bool, summary *ProcessingSummary) error {

	if err := compressImage(srcPath, targetPath, compression); err != nil {
		return err
	}

	if deleteFile {
		if err := os.Remove(srcPath); err != nil {
			return fmt.Errorf("failed to delete original file %s: %v", srcPath, err)
		}
		fmt.Printf("Compressed and moved file: %s\n", targetPath)
	} else {
		fmt.Printf("Compressed file: %s\n", targetPath)
	}
	summary.Compressed++
	return nil
}

func handleNonJPGFile(srcPath, targetPath string, deleteFile bool, summary *ProcessingSummary) error {

	if deleteFile {
		if err := os.Rename(srcPath, targetPath); err != nil {
			return fmt.Errorf("failed to move file %s to %s: %w", srcPath, targetPath, err)
		}
		fmt.Printf("Moved file: %s\n", targetPath)
	} else {
		if err := copyFile(srcPath, targetPath); err != nil {
			return err
		}
		fmt.Printf("Copied file: %s\n", targetPath)
	}
	summary.Moved++
	return nil
}

func copyFile(src, dest string) error {
	srcFile, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("failed to open source file %s: %w", src, err)
	}
	defer srcFile.Close()

	destFile, err := os.Create(dest)
	if err != nil {
		return fmt.Errorf("failed to create destination file %s: %w", dest, err)
	}
	defer destFile.Close()

	if _, err := io.Copy(destFile, srcFile); err != nil {
		return fmt.Errorf("failed to copy file from %s to %s: %w", src, dest, err)
	}
	return nil
}

// compressImage compresses a JPG image to the specified quality level and moves it to the destination.
func compressImage(src, dest string, quality int) error {
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
