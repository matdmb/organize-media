package organizemedia

import (
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/matdmb/organize-media/pkg/models"
	"github.com/matdmb/organize-media/pkg/utils"
)

func Organize(params *models.Params) error {
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

	var logOutput io.Writer
	// Setup logger
	logOutput, err := setupLogger(params.EnableLog)
	if err != nil {
		return err
	}
	log.SetOutput(logOutput)

	log.Println("Application started.")

	log.Printf("Source directory: %s", params.Source)
	log.Printf("Destination directory: %s", params.Destination)

	if params.Compression >= 0 {
		log.Printf("Compression level: %d", params.Compression)
	} else {
		log.Printf("Compression: not applied")
	}

	log.Printf("Delete source files: %t", params.DeleteSource)

	// Count files in the source directory
	totalFiles, size, err := utils.CountFiles(params.Source)
	if err != nil {
		return fmt.Errorf("error counting files: %v", err)
	}

	if totalFiles == 0 {
		return fmt.Errorf("no files to process in source directory")
	}

	fmt.Printf("Number of files to process: %d [%s]\n", totalFiles, formatSize(size))

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

	// Ensure destination directory is writable
	testFile := filepath.Join(params.Destination, "test_write.tmp")
	if err := os.WriteFile(testFile, []byte("test"), 0644); err != nil {
		return fmt.Errorf("destination directory is not writable: %v", err)
	}
	// Remove the test file after the check
	defer os.Remove(testFile)

	summary, err := utils.ProcessMediaFiles(params)
	if err != nil {
		return fmt.Errorf("error moving files: %v", err)
	}

	// Print processing summary
	log.Printf("Processing Summary:")
	log.Printf("%d files have been successfully processed", summary.Processed)
	log.Printf("Number of files copied: %d", summary.Copied)
	log.Printf("Number of files compressed: %d", summary.Compressed)
	log.Printf("Number of files deleted: %d", summary.Deleted)
	log.Printf("Number of files skipped: %d", summary.Skipped)
	log.Println("Process completed.")

	return nil
}

// formatSize formats the size in bytes to a human-readable string in GB, MB, or KB.
func formatSize(size int64) string {
	const (
		KB = 1 << 10
		MB = 1 << 20
		GB = 1 << 30
	)

	switch {
	case size >= GB:
		return fmt.Sprintf("%.2f GB", float64(size)/GB)
	case size >= MB:
		return fmt.Sprintf("%.2f MB", float64(size)/MB)
	case size >= KB:
		return fmt.Sprintf("%.2f KB", float64(size)/KB)
	default:
		return fmt.Sprintf("%d bytes", size)
	}
}

func setupLogger(enableLog bool) (io.Writer, error) {
	if enableLog {
		// Create logs directory if it doesn't exist
		destinationFolder := "./logs"
		if err := os.MkdirAll(destinationFolder, 0755); err != nil {
			return nil, fmt.Errorf("failed to create logs directory: %v", err)
		}

		// Create log file with timestamped name
		logFileName := time.Now().Format("2006-01-02_15-04-05") + ".log"
		logFilePath := filepath.Join(destinationFolder, logFileName)

		// Open the log file
		logFile, err := os.OpenFile(logFilePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
		if err != nil {
			return nil, fmt.Errorf("failed to open log file: %v", err)
		}

		// Set log output to the file
		log.SetFlags(log.LstdFlags | log.Lmicroseconds)
		log.SetOutput(logFile)
		log.Println("Log initialized at", time.Now().Format(time.RFC1123))

		// Return multi-writer to log to both terminal and log file
		return io.MultiWriter(os.Stdout, logFile), nil
	}

	// Default to logging only to the terminal
	return os.Stdout, nil
}
