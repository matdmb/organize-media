package main

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/matdmb/organize_pictures/internal"
)

func main() {
	if err := run(os.Args, os.Stdin); err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
}

func run(args []string, input *os.File) error {
	source, dest, compression, err := readParameters(args)
	if err != nil {
		return err
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
		return fmt.Errorf("error counting files: %v", err)
	}

	fmt.Printf("Total files to process: %d\n", totalFiles)
	if totalFiles == 0 {
		fmt.Println("No files to process. Exiting.")
		return nil
	}

	fmt.Printf("Do you want to proceed with processing %d files? (y/n): ", totalFiles)
	var response string
	if _, err := fmt.Fscanln(input, &response); err != nil {
		return fmt.Errorf("error reading input: %v", err)
	}
	if strings.ToLower(response) != "y" {
		fmt.Println("Operation cancelled.")
		return nil
	}

	files, err := internal.ListFiles(source)
	if err != nil {
		return fmt.Errorf("error listing files: %v", err)
	}

	summary, err := internal.MoveFiles(files, dest, compression)
	if err != nil {
		return fmt.Errorf("error moving files: %v", err)
	}

	fmt.Printf("\nProcessing Summary:\n")
	fmt.Printf("%d files have been successfully processed.\n", len(files))
	fmt.Printf("Files moved without compression: %d\n", summary.Moved)
	fmt.Printf("Files compressed and moved: %d\n", summary.Compressed)
	fmt.Printf("Files skipped: %d\n", summary.Skipped)

	return nil
}

func readParameters(args []string) (string, string, int, error) {
	if len(args) < 3 || len(args) > 4 {
		return "", "", -1, fmt.Errorf("usage: %s <source_dir> <destination_dir> [compression (0-100)]", args[0])
	}

	source := args[1]
	dest := args[2]

	if _, err := os.Stat(source); os.IsNotExist(err) {
		return "", "", -1, fmt.Errorf("source directory does not exist: %s", source)
	}

	if _, err := os.Stat(dest); os.IsNotExist(err) {
		return "", "", -1, fmt.Errorf("destination directory does not exist: %s", dest)
	}

	// Default compression level to -1 (no compression) if not provided
	compression := -1
	if len(args) == 4 {
		var err error
		compression, err = strconv.Atoi(args[3])
		if err != nil || compression < 0 || compression > 100 {
			return "", "", -1, fmt.Errorf("compression level must be an integer between 0 and 100")
		}
	}

	return source, dest, compression, nil
}
