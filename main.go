package main

import (
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/matdmb/organize_pictures/internal"
)

func main() {
	source, dest := readParameters()

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
	fmt.Scanln(&response)
	if strings.ToLower(response) != "y" {
		fmt.Println("Operation cancelled.")
		return
	}

	files, err := internal.ListFiles(source)
	if err != nil {
		log.Fatalf("Error listing files: %v", err)
	}

	err = internal.MoveFiles(files, dest)
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
