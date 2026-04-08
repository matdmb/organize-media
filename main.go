package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/matdmb/organize-media/pkg/models"
	"github.com/matdmb/organize-media/pkg/organizemedia"
)

// For testing purposes
var osExit = os.Exit

func main() {
	// Define flags
	source := flag.String("source", "", "Path to the source directory containing pictures")
	dest := flag.String("dest", "", "Path to the destination directory for organized pictures")
	compression := flag.Int("compression", -1, "Compression level for JPG files (0-100, optional)")
	workers := flag.Int("workers", 0, "Number of workers to use for parallel processing (default: number of CPUs)")
	delete := flag.Bool("delete", false, "Delete source files after processing")
	logFile := flag.Bool("enable-log", false, "Enable logging to a file")

	// Parse the flags
	flag.Parse()

	// Validate required flags
	if err := validateFlags(*source, *dest); err != nil {
		handleValidationError()
	}

	// Run with validated params
	runOrganize(*source, *dest, *compression, *workers, *delete, *logFile)
}

// validateFlags checks if required flags are provided
func validateFlags(source, dest string) error {
	if source == "" || dest == "" {
		return fmt.Errorf("source and destination directories are required")
	}
	return nil
}

// handleValidationError prints usage info and exits
func handleValidationError() {
	fmt.Println("Usage:")
	fmt.Println("  -source    Source directory containing media files")
	fmt.Println("  -dest      Destination directory for organized files")
	fmt.Println("  -compression  JPEG compression level (0-100, default: 90, -1 to disable)")
	fmt.Println("  -workers    Number of workers to use for parallel processing (default: 1)")
	fmt.Println("  -delete    Delete source files after successful processing (default: false)")
	fmt.Println("  -enable-log  Enable logging to file (default: false)")
	fmt.Println("\nExample:")
	fmt.Println("  ./organize-media -source /path/to/photos -dest /path/to/organized")
	osExit(1)
}

// runOrganize runs the organize logic with the given parameters
func runOrganize(source, dest string, compression int, workers int, delete, logFile bool) {
	// Initialize Params struct
	params := &models.Params{
		Source:       source,
		Destination:  dest,
		Compression:  compression,
		Workers:      workers,
		DeleteSource: delete,
		EnableLog:    logFile,
	}

	// Run the main logic
	if err := organizemedia.Organize(params); err != nil {
		log.Fatalf("Error: %v", err)
	}
}
