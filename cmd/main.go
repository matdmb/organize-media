package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/matdmb/organize-media/pkg/organizemedia"
)

func main() {
	// Define flags
	source := flag.String("source", "", "Path to the source directory containing pictures")
	dest := flag.String("dest", "", "Path to the destination directory for organized pictures")
	compression := flag.Int("compression", -1, "Compression level for JPG files (0-100, optional)")
	delete := flag.Bool("delete", false, "Delete source files after processing")
	logFile := flag.Bool("enable-log", false, "Enable logging to a file")

	// Parse the flags
	flag.Parse()

	// Validate required flags
	if *source == "" || *dest == "" {
		fmt.Println("Usage:")
		flag.PrintDefaults()
		os.Exit(1)
	}

	// Initialize Params struct
	params := &organizemedia.Params{
		Source:       *source,
		Destination:  *dest,
		Compression:  *compression,
		DeleteSource: *delete,
		EnableLog:    *logFile,
	}

	// Run the main logic
	if err := organizemedia.Organize(params); err != nil {
		log.Fatalf("Error: %v", err)
	}
}
