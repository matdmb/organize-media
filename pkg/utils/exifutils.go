package utils

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"strings"
	"time"
)

// EXIF tag constants
const (
	ExifIdentifier       = "Exif\x00\x00"
	TiffHeaderLength     = 8
	TagDateTime          = 0x0132 // standard date/time tag
	TagDateTimeOriginal  = 0x9003 // when photo was taken
	TagDateTimeDigitized = 0x9004 // when photo was digitized

	// TIFF header byte order markers
	BigEndianMarker    = "MM"
	LittleEndianMarker = "II"
)

// TiffHeader represents the header of a TIFF file
type TiffHeader struct {
	ByteOrder binary.ByteOrder
	Offset    int64
}

// EXIF datetime format according to the standard
const ExifTimeLayout = "2006:01:02 15:04:05"

// Define supported image file extensions
// SupportedExtensions contains all image formats that can be processed by the application
var SupportedExtensions = map[string]bool{
	".jpg":  true,
	".jpeg": true,
	".nef":  true, // Nikon RAW
	".cr2":  true, // Canon RAW
	".cr3":  true, // Canon RAW
	".arw":  true, // Sony RAW
	".heic": true, // Apple HEIC
	".heif": true, // Apple HEIF
	".raf":  true, // Fujifilm RAW
	".rw2":  true, // Panasonic RAW
	".dng":  true, // Adobe DNG
	".raw":  true, // Generic RAW
	// Add more formats here as needed
}

// GetImageDateTime extracts the date and time from an image buffer
func GetImageDateTime(buffer []byte, fileExt string) (time.Time, error) {
	// Create a reader from the buffer
	reader := bytes.NewReader(buffer)

	ext := strings.ToLower(fileExt)

	// Try different extraction strategies based on file format
	strategies := []func(io.ReadSeeker, string) (time.Time, error){
		ExtractExifFromJPEG,    // JPEG-specific structure
		ExtractExifFromTIFF,    // Standard TIFF structure (works for most RAW)
		ExtractExifWithOffsets, // Try different offsets (for CR2, etc.)
		ScanForDateTimeString,  // Last resort fallback
	}

	// For non-JPEG files, we can skip the JPEG-specific strategy
	if ext != ".jpg" && ext != ".jpeg" {
		strategies = strategies[1:]
	}

	// Try each strategy in order
	for _, strategy := range strategies {
		// Reset reader position before each attempt
		if _, err := reader.Seek(0, io.SeekStart); err != nil {
			return time.Time{}, err
		}

		t, err := strategy(reader, ext)
		if err == nil {
			return t, nil
		}
		// If this strategy failed, continue with the next one
	}

	return time.Time{}, fmt.Errorf("no date/time information found")
}

// ExtractExifFromJPEG extracts date/time from JPEG data in a buffer
func ExtractExifFromJPEG(reader io.ReadSeeker, _ string) (time.Time, error) {
	// JPEG starts with SOI marker FF D8
	buf := make([]byte, 2)
	if _, err := io.ReadFull(reader, buf); err != nil {
		return time.Time{}, err
	}

	if buf[0] != 0xFF || buf[1] != 0xD8 {
		return time.Time{}, fmt.Errorf("not a valid JPEG file")
	}

	// Search for the EXIF APP1 marker (FF E1)
	for {
		// Read marker - this may fail at EOF which is fine
		if _, err := io.ReadFull(reader, buf); err != nil {
			break
		}

		// Look for 0xFF marker
		if buf[0] != 0xFF {
			continue
		}

		// If we found the APP1 marker (0xE1)
		if buf[1] == 0xE1 {
			// Read length (2 bytes)
			lengthBuf := make([]byte, 2)
			if _, err := io.ReadFull(reader, lengthBuf); err != nil {
				break
			}

			// The length includes the length bytes themselves
			length := int(lengthBuf[0])<<8 | int(lengthBuf[1])

			// Check for EXIF header (6 bytes: "Exif\0\0")
			exifHeader := make([]byte, 6)
			if _, err := io.ReadFull(reader, exifHeader); err != nil {
				break
			}

			if string(exifHeader) == ExifIdentifier {
				// Parse the TIFF data that follows
				t, err := ParseTIFFHeader(reader)
				if err == nil {
					return t, nil
				}
			}

			// Skip the rest of this segment if we didn't find EXIF or couldn't parse it
			skipLength := length - 2 - 6 // Subtract length bytes and EXIF header
			if skipLength > 0 {
				if _, err := reader.Seek(int64(skipLength), io.SeekCurrent); err != nil {
					break
				}
			}
		} else if buf[1] == 0xDA {
			// Start of scan - no more metadata
			break
		} else {
			// It's some other marker, skip it
			lengthBuf := make([]byte, 2)
			if _, err := io.ReadFull(reader, lengthBuf); err != nil {
				break
			}

			length := int(lengthBuf[0])<<8 | int(lengthBuf[1])
			if length < 2 {
				// Invalid length
				break
			}

			// Skip the rest of this segment
			if _, err := reader.Seek(int64(length-2), io.SeekCurrent); err != nil {
				break
			}
		}
	}

	return time.Time{}, fmt.Errorf("no EXIF data found in JPEG structure")
}

// ExtractExifFromTIFF tries to parse the buffer as a standard TIFF structure
func ExtractExifFromTIFF(reader io.ReadSeeker, _ string) (time.Time, error) {
	return ParseTIFFHeader(reader)
}

// ExtractExifWithOffsets tries to find EXIF data at various offsets in the buffer
// This is useful for some RAW formats that have different header structures
func ExtractExifWithOffsets(reader io.ReadSeeker, ext string) (time.Time, error) {
	// Different offsets to try based on file type
	var offsets []int64

	// For CR2 files (Canon), they often have different structures
	if ext == ".cr2" {
		offsets = []int64{0, 8, 16}
	} else if ext == ".arw" {
		offsets = []int64{0, 4, 8, 12}
	} else if ext == ".nef" {
		offsets = []int64{0, 4, 8}
	} else {
		// Default offsets to try for other formats
		offsets = []int64{0, 4, 8}
	}

	// Try each offset
	for _, offset := range offsets {
		if _, err := reader.Seek(offset, io.SeekStart); err != nil {
			continue
		}
		t, err := ParseTIFFHeader(reader)
		if err == nil {
			return t, nil
		}
	}

	return time.Time{}, fmt.Errorf("couldn't find EXIF data at known offsets")
}

// ScanForDateTimeString scans a buffer for strings that match the EXIF date/time format
func ScanForDateTimeString(reader io.ReadSeeker, _ string) (time.Time, error) {
	// Reset to beginning of buffer
	if _, err := reader.Seek(0, io.SeekStart); err != nil {
		return time.Time{}, err
	}

	// Create a buffer to read chunks
	buffer := make([]byte, 4096)
	offset := int64(0)

	// Read in chunks
	for {
		n, err := reader.Read(buffer)
		if err != nil && err != io.EOF {
			return time.Time{}, err
		}
		if n == 0 {
			break
		}

		// Convert buffer to string for matching
		content := string(buffer[:n])

		// Look for date patterns
		for i := 0; i < len(content)-19; i++ {
			potentialDate := content[i : i+19]
			// Check if it matches our pattern
			if len(potentialDate) == 19 &&
				potentialDate[4] == ':' && potentialDate[7] == ':' &&
				potentialDate[10] == ' ' && potentialDate[13] == ':' &&
				potentialDate[16] == ':' {

				// Try to parse it
				t, err := time.Parse(ExifTimeLayout, potentialDate)
				if err == nil && t.Year() >= 1990 && t.Year() <= 2100 {
					// Looks like a valid date
					return t, nil
				}
			}
		}

		offset += int64(n)

		// If we've scanned a reasonable amount without finding a date,
		// it's unlikely we'll find one
		if offset > 1024*1024 { // 1MB
			break
		}
	}

	return time.Time{}, fmt.Errorf("no date/time information found")
}

// ParseTIFFHeader parses TIFF header and IFD entries to find date/time
func ParseTIFFHeader(r io.Reader) (time.Time, error) {
	// Read byte order marker
	orderMarker := make([]byte, 2)
	if _, err := io.ReadFull(r, orderMarker); err != nil {
		return time.Time{}, err
	}

	var byteOrder binary.ByteOrder
	if string(orderMarker) == BigEndianMarker {
		byteOrder = binary.BigEndian
	} else if string(orderMarker) == LittleEndianMarker {
		byteOrder = binary.LittleEndian
	} else {
		return time.Time{}, fmt.Errorf("invalid TIFF byte order marker")
	}

	// Verify TIFF marker (should be 42)
	marker := make([]byte, 2)
	if _, err := io.ReadFull(r, marker); err != nil {
		return time.Time{}, err
	}

	if byteOrder.Uint16(marker) != 42 {
		return time.Time{}, fmt.Errorf("invalid TIFF marker")
	}

	// Get offset to first IFD
	offsetBytes := make([]byte, 4)
	if _, err := io.ReadFull(r, offsetBytes); err != nil {
		return time.Time{}, err
	}

	ifdOffset := byteOrder.Uint32(offsetBytes)

	// Seek to first IFD (relative to TIFF header start)
	if seeker, ok := r.(io.ReadSeeker); ok {
		// Calculate the current position (right after reading the offset)
		currentPos, err := seeker.Seek(0, io.SeekCurrent)
		if err != nil {
			return time.Time{}, err
		}

		// Calculate the start of TIFF header (current position - 8 bytes we've read)
		tiffHeaderStart := currentPos - 8

		// Seek to IFD from the start of TIFF header
		if _, err := seeker.Seek(tiffHeaderStart+int64(ifdOffset), io.SeekStart); err != nil {
			return time.Time{}, err
		}
	} else {
		// We can't seek in this reader, so we need to read and discard bytes
		toSkip := int(ifdOffset) - 8
		if toSkip < 0 {
			return time.Time{}, fmt.Errorf("invalid IFD offset")
		}
		skipBuf := make([]byte, 1024)
		for toSkip > 0 {
			n := toSkip
			if n > 1024 {
				n = 1024
			}
			read, err := r.Read(skipBuf[:n])
			if err != nil {
				return time.Time{}, err
			}
			toSkip -= read
		}
	}

	// Read IFD entry count
	entryCountBytes := make([]byte, 2)
	if _, err := io.ReadFull(r, entryCountBytes); err != nil {
		return time.Time{}, err
	}
	entryCount := byteOrder.Uint16(entryCountBytes)

	// Process each IFD entry
	for i := 0; i < int(entryCount); i++ {
		entryBytes := make([]byte, 12) // Each IFD entry is 12 bytes
		if _, err := io.ReadFull(r, entryBytes); err != nil {
			return time.Time{}, err
		}

		tag := byteOrder.Uint16(entryBytes[0:2])
		dataType := byteOrder.Uint16(entryBytes[2:4])
		count := byteOrder.Uint32(entryBytes[4:8])
		valueOffset := byteOrder.Uint32(entryBytes[8:12])

		// Check if it's one of the date/time tags
		if (tag == TagDateTimeOriginal || tag == TagDateTime || tag == TagDateTimeDigitized) && dataType == 2 /* ASCII */ {
			// For date strings within the IFD entry
			if count <= 4 {
				continue // Too short for a valid date
			}

			// For date strings that require seeking
			// Date/time strings are usually longer than 4 bytes
			// so they're stored elsewhere in the file
			if seeker, ok := r.(io.ReadSeeker); ok {
				currentPos, _ := seeker.Seek(0, io.SeekCurrent)

				// Calculate the start of TIFF header
				tiffHeaderStart := currentPos
				// Find where we are in the IFD to calculate TIFF header start
				tiffHeaderStart -= int64(12*(i+1) + 2) // 12 bytes per entry, 2 bytes for entry count

				// Seek to the string (relative to TIFF header)
				if _, err := seeker.Seek(tiffHeaderStart-8+int64(valueOffset), io.SeekStart); err != nil {
					return time.Time{}, err
				}

				// Date/time format is "YYYY:MM:DD HH:MM:SS\0"
				dateBytes := make([]byte, 20)
				if _, err := io.ReadFull(r, dateBytes); err != nil {
					return time.Time{}, err
				}

				// Go back to the IFD entries
				if _, err := seeker.Seek(currentPos, io.SeekStart); err != nil {
					return time.Time{}, err
				}

				dateStr := string(dateBytes[:19]) // Remove null terminator
				t, err := time.Parse(ExifTimeLayout, dateStr)
				if err != nil {
					continue // Try other date tags
				}
				return t, nil
			}
		}
	}

	return time.Time{}, fmt.Errorf("no date/time information found")
}
