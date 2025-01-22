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
			return err
		}

		newPath := filepath.Join(monthDayDir, filepath.Base(file.Path))
		if err := os.Rename(file.Path, newPath); err != nil {
			return err
		}
	}

	return nil
}
