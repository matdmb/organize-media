package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/rwcarlsen/goexif/exif"
)

type ImageFile struct {
	Path string
	Date time.Time
}

var allowedExtensions = []string{".jpg", ".nef"}

func main() {
	source, dest := readParameters()

	totalFiles, err := countFiles(source)
	if err != nil {
		log.Fatalf("Error counting files: %v", err)
	}

	fmt.Printf("Total files to move: %d\n", totalFiles)

	if totalFiles == 0 {
		fmt.Println("No files to move. Exiting.")
		return
	}
	fmt.Printf("Do you want to proceed with moving %d files? (y/n): ", totalFiles)
	var response string
	fmt.Scanln(&response)
	if strings.ToLower(response) != "y" {
		fmt.Println("Operation cancelled.")
		return
	}

	files, err := listFiles(source)
	if err != nil {
		log.Fatalf("Error listing files: %v", err)
	}

	err = moveFiles(files, dest)
	if err != nil {
		log.Fatalf("Error moving files: %v", err)
	}

	fmt.Printf("%d files have been successfully moved.\n", len(files))
}

func readParameters() (string, string) {
	if len(os.Args) != 3 {
		log.Fatalf("Usage: %s <source_directory> <destination_directory>", os.Args[0])
	}

	source := os.Args[1]
	dest := os.Args[2]

	if _, err := os.Stat(source); os.IsNotExist(err) {
		log.Fatalf("Source directory does not exist: %s", source)
	}

	if _, err := os.Stat(dest); os.IsNotExist(err) {
		log.Fatalf("Destination directory does not exist: %s", dest)
	}

	return source, dest
}

func listFiles(dir string) ([]ImageFile, error) {
	var files []ImageFile

	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if !info.IsDir() && isAllowedExtension(filepath.Ext(info.Name())) {
			fileDate, err := getExifDate(path)
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

func getExifDate(path string) (time.Time, error) {
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

func moveFiles(files []ImageFile, dest string) error {
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

func countFiles(dir string) (int, error) {
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
