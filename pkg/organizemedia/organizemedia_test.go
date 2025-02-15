package organizemedia

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/matdmb/organize-media/pkg/models"
)

func TestFormatSize(t *testing.T) {
	tests := []struct {
		name     string
		size     int64
		expected string
	}{
		{
			name:     "Bytes",
			size:     500,
			expected: "500 bytes",
		},
		{
			name:     "Kilobytes",
			size:     1500,
			expected: "1.46 KB",
		},
		{
			name:     "Megabytes",
			size:     1500000,
			expected: "1.43 MB",
		},
		{
			name:     "Gigabytes",
			size:     1500000000,
			expected: "1.40 GB",
		},
		{
			name:     "Zero",
			size:     0,
			expected: "0 bytes",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatSize(tt.size)
			if result != tt.expected {
				t.Errorf("formatSize(%d) = %s; want %s", tt.size, result, tt.expected)
			}
		})
	}
}

func TestSetupLogger(t *testing.T) {
	// Clean up after tests
	defer os.RemoveAll("./logs")

	tests := []struct {
		name      string
		enableLog bool
		wantErr   bool
	}{
		{
			name:      "logging disabled",
			enableLog: false,
			wantErr:   false,
		},
		{
			name:      "logging enabled",
			enableLog: true,
			wantErr:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			writer, err := setupLogger(tt.enableLog)

			// Check error cases
			if (err != nil) != tt.wantErr {
				t.Errorf("setupLogger() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			// Check writer is not nil
			if writer == nil {
				t.Error("setupLogger() returned nil writer")
				return
			}

			if tt.enableLog {
				// Verify logs directory was created
				if _, err := os.Stat("./logs"); os.IsNotExist(err) {
					t.Error("logs directory was not created")
				}

				// Verify at least one log file exists
				files, err := os.ReadDir("./logs")
				if err != nil {
					t.Errorf("failed to read logs directory: %v", err)
				}
				if len(files) == 0 {
					t.Error("no log files were created")
				}

				// Verify log file is writable
				logFile := files[0]
				logPath := filepath.Join("./logs", logFile.Name())
				f, err := os.OpenFile(logPath, os.O_WRONLY|os.O_APPEND, 0666)
				if err != nil {
					t.Errorf("log file is not writable: %v", err)
				}
				f.Close()
			}
		})
	}
}

func TestOrganize(t *testing.T) {
	// Create temp destination directory
	destDir := t.TempDir()

	tests := []struct {
		name        string
		params      *models.Params
		wantErr     bool
		setupFiles  bool
		errorString string
	}{
		{
			name: "successful organization with sample file",
			params: &models.Params{
				Source:        "../testdata/exif/sample_with_exif.jpg",
				Destination:   destDir,
				Compression:   80,
				DeleteSource:  false,
				EnableLog:     false,
				SkipUserInput: true,
			},
			wantErr: false,
		},
		{
			name: "handle corrupted EXIF data",
			params: &models.Params{
				Source:        "../testdata/exif/sample_corrupted_exif.jpg",
				Destination:   destDir,
				Compression:   80,
				DeleteSource:  false,
				SkipUserInput: true,
			},
			wantErr:     true,
			errorString: "exif: decode failed",
		},
		{
			name: "handle image without EXIF data",
			params: &models.Params{
				Source:        "../testdata/exif/sample_without_exif.jpg",
				Destination:   destDir,
				Compression:   80,
				DeleteSource:  false,
				SkipUserInput: true,
			},
			wantErr:     true,
			errorString: "exif: failed to find exif",
		},
		{
			name: "non-existent source directory",
			params: &models.Params{
				Source:        "/non/existent/path",
				Destination:   destDir,
				Compression:   80,
				SkipUserInput: true,
			},
			wantErr:     true,
			errorString: "source directory does not exist",
		},
		{
			name: "invalid compression value",
			params: &models.Params{
				Source:        "../testdata/exif",
				Destination:   destDir,
				Compression:   101,
				SkipUserInput: true,
			},
			wantErr:     true,
			errorString: "compression level must be an integer between 0 and 100",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a unique destination directory for each test
			testDestDir := filepath.Join(destDir, tt.name)
			err := os.MkdirAll(testDestDir, 0755)
			if err != nil {
				t.Fatalf("Failed to create test destination directory: %v", err)
			}
			tt.params.Destination = testDestDir

			err = Organize(tt.params)

			// Check error cases
			if (err != nil) != tt.wantErr {
				t.Errorf("Organize() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr && !strings.Contains(err.Error(), tt.errorString) {
				t.Errorf("Organize() error = %v, want error containing %v", err, tt.errorString)
				return
			}

			// For successful cases, verify files were processed
			if !tt.wantErr {
				// Check if files exist in destination
				files, err := filepath.Glob(filepath.Join(testDestDir, "*/*/*.jpg"))
				if err != nil {
					t.Errorf("Failed to check processed files: %v", err)
				}
				if len(files) == 0 {
					t.Error("No files were processed")
				}

				// Verify directory structure (YYYY/MM-DD)
				for _, file := range files {
					relPath, err := filepath.Rel(testDestDir, file)
					if err != nil {
						t.Errorf("Failed to get relative path: %v", err)
					}
					parts := strings.Split(relPath, string(os.PathSeparator))
					if len(parts) != 3 {
						t.Errorf("Unexpected directory structure: %v", relPath)
					}
				}
			}
		})
	}
}
