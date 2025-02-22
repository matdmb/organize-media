package utils

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/matdmb/organize-media/pkg/models"
)

func BenchmarkProcessSpecificFiles(b *testing.B) {
	// Stop timer during setup
	b.StopTimer()

	// Create temp directories
	destDir := b.TempDir()
	sourceDir := b.TempDir()

	// Copy test files to source directory
	testFiles := []string{"DSC00001.ARW", "DSC00001.JPG"}
	for _, file := range testFiles {
		srcPath := filepath.Join("../testdata/sony", file)
		destPath := filepath.Join(sourceDir, file)

		if err := copyFile(srcPath, destPath); err != nil {
			b.Fatalf("Failed to copy test file %s: %v", file, err)
		}
	}

	// Run benchmark
	for i := 0; i < b.N; i++ {
		// Create unique destination for this iteration
		iterDir := filepath.Join(destDir, fmt.Sprintf("iter_%d", i))

		params := &models.Params{
			Source:      sourceDir,
			Destination: iterDir,
			Compression: 80,
		}

		// Start timing the actual processing
		b.StartTimer()

		_, err := ProcessMediaFiles(params)

		// Stop timing to exclude cleanup
		b.StopTimer()

		if err != nil {
			b.Fatalf("ProcessMediaFiles failed: %v", err)
		}

		// Optional: cleanup iteration directory
		os.RemoveAll(iterDir)
	}
}

func copyFile(src, dst string) error {
	data, err := os.ReadFile(src)
	if err != nil {
		return err
	}
	return os.WriteFile(dst, data, 0644)
}
