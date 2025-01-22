package internal

import (
	"fmt"
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

func ListFiles(dir string) ([]ImageFile, error) {
	var files []ImageFile

	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
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

func isAllowedExtension(ext string) bool {
	ext = strings.ToLower(ext)
	for _, allowed := range allowedExtensions {
		if ext == allowed {
			return true
		}
	}
	return false
}

func CountFiles(dir string) (int, error) {
	var count int

	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if !info.IsDir() && isAllowedExtension(filepath.Ext(info.Name())) {
			count++
		}
		return nil
	})

	return count, err
}

func MoveFiles(files []ImageFile, dest string) error {
	for _, file := range files {
		yearDir := filepath.Join(dest, fmt.Sprintf("%d", file.Date.Year()))
		monthDayDir := filepath.Join(yearDir, fmt.Sprintf("%02d-%02d", file.Date.Month(), file.Date.Day()))

		if err := os.MkdirAll(monthDayDir, os.ModePerm); err != nil {
			return fmt.Errorf("failed to create directory %s: %w", monthDayDir, err)
		}

		newPath := filepath.Join(monthDayDir, filepath.Base(file.Path))

		// Checks if file already exists in destination
		if _, err := os.Stat(newPath); err == nil {
			// File exists, skipping
			fmt.Printf("File already exists: %s, skipping...\n", newPath)
			continue
		} else if !os.IsNotExist(err) {
			// Another error occurred
			return fmt.Errorf("error checking destination file %s: %w", newPath, err)
		}

		if err := os.Rename(file.Path, newPath); err != nil {
			return fmt.Errorf("failed to move file %s to %s: %w", file.Path, newPath, err)
		}
		fmt.Printf("Moved file %s to %s\n", file.Path, newPath)
	}

	return nil
}
