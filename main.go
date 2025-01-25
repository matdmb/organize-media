package main

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/matdmb/organize_pictures/internal"
)

func main() {
	source, dest, compression, err := readParameters()
	if err != nil {
		log.Fatalf("Error: %v", err)
	}

	log.Printf("Source directory: %s", source)
	log.Printf("Destination directory: %s", dest)
	if compression >= 0 {
		log.Printf("Compression level: %d", compression)
	} else {
		log.Printf("Compression: not applied")
	}

	totalFiles, err := internal.CountFiles(source)
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
	if _, err := fmt.Scanln(&response); err != nil {
		fmt.Printf("Error reading input: %v\n", err)
		return
	}
	if strings.ToLower(response) != "y" {
		fmt.Println("Operation cancelled.")
		return
	}

	// Define the real GetExifDate function
	realGetExifDate := func(path string, decoder func(string) ([]byte, error), parser func([]byte) (time.Time, error)) (time.Time, error) {
		// Use real EXIF reading logic here
		return internal.DefaultGetExifDate(path)
	}

	files, err := internal.ListFilesWithExif(source, realGetExifDate)
	if err != nil {
		log.Fatalf("Error listing files: %v", err)
	}

	err = internal.MoveFiles(files, dest, compression)
	if err != nil {
		log.Fatalf("Error moving files: %v", err)
	}

	fmt.Printf("%d files have been successfully moved.\n", len(files))
}

func readParameters() (string, string, int, error) {
	if len(os.Args) < 3 || len(os.Args) > 4 {
		return "", "", -1, fmt.Errorf("usage: %s <source_dir> <destination_dir> [compression (0-100)]", os.Args[0])
	}

	source := os.Args[1]
	dest := os.Args[2]

	if _, err := os.Stat(source); os.IsNotExist(err) {
		log.Fatalf("Source directory does not exist: %s", source)
	}

	if _, err := os.Stat(dest); os.IsNotExist(err) {
		log.Fatalf("Destination directory does not exist: %s", dest)
	}

	// Default compression level to -1 (no compression) if not provided
	compression := -1

	if len(os.Args) == 4 {
		var err error
		compression, err = strconv.Atoi(os.Args[3])
		if err != nil || compression < 0 || compression > 100 {
			return "", "", -1, fmt.Errorf("compression level must be an integer between 0 and 100")
		}
	}

	return source, dest, compression, nil
}
