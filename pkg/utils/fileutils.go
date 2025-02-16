package utils

import (
	"bytes"
	"fmt"
	"image"
	"image/jpeg"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/matdmb/organize-media/pkg/models"
	"github.com/rwcarlsen/goexif/exif"
)

type ImageFile struct {
	Path string
	Date time.Time
}

type ProcessingSummary struct {
	Processed  int
	Compressed int
	Copied     int
	Skipped    int
	Deleted    int
}

var allowedExtensions = []string{".jpg", ".nef", ".cr2", "cr3", ".dng", ".arw", "raw"}

// copyOrCompressImage processes the buffer, compressing if it's a JPG, and writes to disk.
func copyOrCompressImage(destPath string, sourceFile string, buffer []byte, isJPG bool, p *models.Params, summary *ProcessingSummary) error {

	// Check if file already exists
	if exists, err := fileExists(destPath); err != nil {
		return fmt.Errorf("failed to check destination file: %w", err)
	} else if exists {
		log.Printf("[SKIPPED] Destination file already exists: %s", destPath)
		summary.Skipped++
		return nil
	}

	// Ensure the destination directory exists
	if err := os.MkdirAll(filepath.Dir(destPath), os.ModePerm); err != nil {
		return err
	}

	var outputBuffer []byte
	var msg string
	if isJPG && p.Compression >= 0 {
		// Decode and re-encode with compression
		img, _, err := image.Decode(bytes.NewReader(buffer))
		if err != nil {
			return err
		}

		var compressedBuffer bytes.Buffer
		err = jpeg.Encode(&compressedBuffer, img, &jpeg.Options{Quality: p.Compression})
		if err != nil {
			return err
		}
		outputBuffer = compressedBuffer.Bytes()
		summary.Compressed++
		msg = "[COMPRESSED]"
	} else {
		// Use the original buffer if not JPG or compression is disabled
		outputBuffer = buffer
		summary.Copied++
		msg = "[COPIED]"
	}

	// Create the destination file
	destFile, err := os.Create(destPath)
	if err != nil {
		return err
	}
	defer destFile.Close()

	// Write the processed buffer
	_, err = destFile.Write(outputBuffer)
	log.Printf("%s Processed file to: %s", msg, destPath)
	summary.Processed++

	if p.DeleteSource {
		if err := os.Remove(sourceFile); err != nil {
			return fmt.Errorf("failed to delete source file: %w", err)
		}
		log.Printf("[DELETED] Deleted source file: %s", sourceFile)
		summary.Deleted++
	}

	return err
}

func ProcessMediaFiles(p *models.Params) (ProcessingSummary, error) {
	var summary ProcessingSummary

	log.Printf("Starting processing files...")

	err := filepath.Walk(p.Source, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return fmt.Errorf("failed to access path %q: %w", path, err)
		}

		if !info.IsDir() && isAllowedExtension(filepath.Ext(info.Name())) {

			fmt.Printf("Processing file: %s\n", path)

			// Open the file
			file, err := os.Open(path)
			if err != nil {
				return err
			}
			defer file.Close()

			// Read the entire file into memory
			buffer, err := io.ReadAll(file)
			if err != nil {
				return err
			}

			// Check if it's a JPG
			isJPG := strings.HasSuffix(strings.ToLower(path), ".jpg") || strings.HasSuffix(strings.ToLower(path), ".jpeg")

			// Decode EXIF data from the buffer
			exifData, err := exif.Decode(bytes.NewReader(buffer))
			if err != nil {
				return err
			}

			// Extract date from EXIF metadata
			date, err := exifData.DateTime()
			if err != nil {
				log.Printf("Warning: could not get EXIF date for file %s: %v", path, err)
				return err
			}

			// Format destination folder structure
			destDir := filepath.Join(p.Destination, fmt.Sprintf("%d", date.Year()), fmt.Sprintf("%02d-%02d", date.Month(), date.Day()))
			destPath := filepath.Join(destDir, filepath.Base(path))

			// Copy or compress before writing
			if err := copyOrCompressImage(destPath, path, buffer, isJPG, p, &summary); err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return summary, fmt.Errorf("failed to walk directory: %w", err)
	}

	return summary, nil
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

	log.Printf("CountFiles: %d files found in %s\n", count, dir)

	return count, totalSize, err
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
