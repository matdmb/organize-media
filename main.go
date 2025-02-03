package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/matdmb/organize-media/internal"
)

type Params struct {
	Source        string
	Destination   string
	Compression   int
	SkipUserInput bool // New flag to bypass user input
	DeleteSource  bool // New flag to delete source files after processing
}

func main() {
	// Define flags
	source := flag.String("source", "", "Path to the source directory containing pictures")
	dest := flag.String("dest", "", "Path to the destination directory for organized pictures")
	compression := flag.Int("compression", -1, "Compression level for JPG files (0-100, optionalle)")
	delete := flag.Bool("delete", false, "Delete source files after processing")

	// Parse the flags
	flag.Parse()

	// Validate required flags
	if *source == "" || *dest == "" {
		fmt.Println("Usage:")
		flag.PrintDefaults()
		os.Exit(1)
	}

	// Initialize Params struct
	params := &Params{
		Source:       *source,
		Destination:  *dest,
		Compression:  *compression,
		DeleteSource: *delete,
	}

	// Run the main logic
	if err := run(params); err != nil {
		log.Fatalf("Error: %v", err)
	}
}

func run(params *Params) error {
	// Validate source directory existence
	if _, err := os.Stat(params.Source); os.IsNotExist(err) {
		return fmt.Errorf("source directory does not exist: %s", params.Source)
	}

	// Validate destination directory existence
	if _, err := os.Stat(params.Destination); os.IsNotExist(err) {
		return fmt.Errorf("destination directory does not exist: %s", params.Destination)
	}

	// Validate compression range
	if params.Compression < -1 || params.Compression > 100 {
		return fmt.Errorf("compression level must be an integer between 0 and 100")
	}

	log.Printf("Source directory: %s", params.Source)
	log.Printf("Destination directory: %s", params.Destination)

	if params.Compression >= 0 {
		log.Printf("Compression level: %d", params.Compression)
	} else {
		log.Printf("Compression: not applied")
	}

	log.Printf("Delete source files: %t", params.DeleteSource)

	// Count files in the source directory
	totalFiles, err := internal.CountFiles(params.Source)
	if err != nil {
		return fmt.Errorf("error counting files: %v", err)
	}

	if totalFiles == 0 {
		return fmt.Errorf("no files to process in source directory")
	}

	fmt.Printf("Total files to process: %d\n", totalFiles)

	if !params.SkipUserInput {
		// Ask for user confirmation
		fmt.Printf("Do you want to proceed with processing %d files? (y/n): ", totalFiles)
		var response string
		if _, err := fmt.Fscanln(os.Stdin, &response); err != nil {
			return fmt.Errorf("error reading input: %v", err)
		}
		if strings.ToLower(response) != "y" {
			fmt.Println("Operation cancelled.")
			return fmt.Errorf("operation cancelled by user")
		}
	} else {
		log.Println("Skipping user input confirmation (test mode).")
	}

	// List files in the source directory
	files, err := internal.ListFiles(params.Source)
	if err != nil {
		return fmt.Errorf("error listing files: %v", err)
	}

	// Ensure destination directory is writable
	testFile := filepath.Join(params.Destination, "test_write.tmp")
	if err := os.WriteFile(testFile, []byte("test"), 0644); err != nil {
		return fmt.Errorf("destination directory is not writable: %v", err)
	}
	// Remove the test file after the check
	defer os.Remove(testFile)

	// Move and optionally compress files
	summary, err := internal.ProcessFiles(files, params.Destination, params.Compression, params.DeleteSource)
	if err != nil {
		return fmt.Errorf("error moving files: %v", err)
	}

	// Print processing summary
	fmt.Printf("\nProcessing Summary:\n")
	fmt.Printf("%d files have been successfully processed.\n", summary.Moved+summary.Compressed)
	fmt.Printf("Files processed without compression: %d\n", summary.Moved)
	fmt.Printf("Files processed and compressed: %d\n", summary.Compressed)
	fmt.Printf("Files skipped: %d\n", summary.Skipped)

	return nil
}
