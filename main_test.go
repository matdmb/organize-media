package main

import (
	"os"
	"testing"
)

func TestRunInvalidArgs(t *testing.T) {
	// Invalid args (missing source or destination)
	args := []string{"app"}
	err := run(args, os.Stdin)
	if err == nil {
		t.Fatal("Expected an error for missing arguments, but got none")
	}

	// Check the error message
	expectedErr := "usage: app <source_dir> <destination_dir> [compression (0-100)]"
	if err.Error() != expectedErr {
		t.Errorf("Expected error '%s', got '%s'", expectedErr, err.Error())
	}
}

func TestReadParametersValid(t *testing.T) {
	// Step 1: Create temporary source and destination directories
	srcDir := t.TempDir()
	destDir := t.TempDir()

	// Step 2: Call readParameters with valid arguments
	args := []string{"app", srcDir, destDir, "50"}
	source, dest, compression, err := readParameters(args)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	// Step 3: Verify the results
	if source != srcDir {
		t.Errorf("Expected source: %s, got: %s", srcDir, source)
	}
	if dest != destDir {
		t.Errorf("Expected destination: %s, got: %s", destDir, dest)
	}
	if compression != 50 {
		t.Errorf("Expected compression: 50, got: %d", compression)
	}
}

func TestReadParametersMissingArguments(t *testing.T) {
	// Step 1: Call readParameters with missing arguments
	args := []string{"app", "/some/source"}
	_, _, _, err := readParameters(args)
	if err == nil {
		t.Fatal("Expected an error for missing arguments, but got none")
	}

	// Step 2: Verify the error message
	expectedErr := "usage: app <source_dir> <destination_dir> [compression (0-100)]"
	if err.Error() != expectedErr {
		t.Errorf("Expected error: %s, got: %s", expectedErr, err.Error())
	}
}

func TestReadParametersInvalidCompression(t *testing.T) {
	// Step 1: Create temporary source and destination directories
	srcDir := t.TempDir()
	destDir := t.TempDir()

	// Step 2: Call readParameters with invalid compression values
	args := []string{"app", srcDir, destDir, "invalid"}
	_, _, _, err := readParameters(args)
	if err == nil {
		t.Fatal("Expected an error for invalid compression, but got none")
	}

	// Verify the error message
	expectedErr := "compression level must be an integer between 0 and 100"
	if err.Error() != expectedErr {
		t.Errorf("Expected error: %s, got: %s", expectedErr, err.Error())
	}

	// Step 3: Test out-of-range compression
	args = []string{"app", srcDir, destDir, "200"}
	_, _, _, err = readParameters(args)
	if err == nil {
		t.Fatal("Expected an error for out-of-range compression, but got none")
	}
}

func TestReadParametersNonexistentSource(t *testing.T) {
	// Step 1: Create a temporary destination directory
	destDir := t.TempDir()

	// Step 2: Call readParameters with a nonexistent source directory
	args := []string{"app", "/nonexistent/source", destDir}
	_, _, _, err := readParameters(args)
	if err == nil {
		t.Fatal("Expected an error for nonexistent source directory, but got none")
	}

	// Verify the error message
	expectedErr := "source directory does not exist: /nonexistent/source"
	if err.Error() != expectedErr {
		t.Errorf("Expected error: %s, got: %s", expectedErr, err.Error())
	}
}

func TestReadParametersNonexistentDestination(t *testing.T) {
	// Step 1: Create a temporary source directory
	srcDir := t.TempDir()

	// Step 2: Call readParameters with a nonexistent destination directory
	args := []string{"app", srcDir, "/nonexistent/dest"}
	_, _, _, err := readParameters(args)
	if err == nil {
		t.Fatal("Expected an error for nonexistent destination directory, but got none")
	}

	// Verify the error message
	expectedErr := "destination directory does not exist: /nonexistent/dest"
	if err.Error() != expectedErr {
		t.Errorf("Expected error: %s, got: %s", expectedErr, err.Error())
	}
}
