package internal

import (
	"fmt"
	"path"
	"testing"
)

var result []ImageFile

func BenchmarkListFiles(b *testing.B) {
	sampleCounts := []int{0, 1, 2, 5, 10, 100}
	for _, sampleCount := range sampleCounts {
		b.Run(fmt.Sprintf("with %d samples", sampleCount), func(b *testing.B) {
			dir := makeBenchmarkMediaDirectory(b, sampleCount)

			var (
				files []ImageFile
				err   error
			)

			for i := 0; i < b.N; i++ {
				b.StartTimer()
				files, err = ListFiles(dir)
				b.StopTimer()

				if err != nil {
					b.Fatalf("failed to list files: %v", err)
				}
			}

			// avoid compiler optimisations eliminating the function under test
			// and artificially lowering the run time of the benchmark.
			result = files
		})
	}

}

func makeBenchmarkMediaDirectory(b *testing.B, sampleCount int) string {
	tmpDir := b.TempDir()

	for i := 0; i < sampleCount; i++ {
		err := copyFile("./testdata/sony/DSC00001.ARW", path.Join(tmpDir, fmt.Sprintf("%d.ARW", i)))
		if err != nil {
			b.Fatalf("failed to copy sample RAW file: %v", err)
		}

		err = copyFile("./testdata/sony/DSC00001.JPG", path.Join(tmpDir, fmt.Sprintf("%d.JPG", i)))
		if err != nil {
			b.Fatalf("failed to copy sample JPG file: %v", err)
		}
	}

	return tmpDir
}
