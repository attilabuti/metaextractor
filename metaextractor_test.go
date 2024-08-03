package metaextractor

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMetaExtractor(t *testing.T) {
	// Setup MetaExtractor
	extractor := NewMetaExtractor(Options{
		//TridPath:     "/usr/bin/trid", // Adjust this path as needed
		//TridDefs:     "/usr/share/trid/triddefs.trd", // Adjust this path as needed
		//TridTimeout:  5 * time.Second,
		//TridMatches:  5,
		//ExifToolPath: "/usr/bin/exiftool", // Adjust this path as needed
	})

	// Test cases
	testCases := []struct {
		name     string
		fileName string
		checkFn  func(*testing.T, Metadata, error)
	}{
		{
			name:     "Empty File",
			fileName: "empty",
			checkFn: func(t *testing.T, metadata Metadata, err error) {
				require.Error(t, err)
				assert.Equal(t, "empty", metadata.Name)
				assert.Equal(t, int64(1044), metadata.Size)
				assert.Equal(t, "", metadata.Extension)
				assert.False(t, metadata.ExtMismatch)
				assert.Empty(t, metadata.Types)
				assert.Empty(t, metadata.Exif)
			},
		},
		{
			name:     "MP3 File",
			fileName: "sample.mp3",
			checkFn: func(t *testing.T, metadata Metadata, err error) {
				require.NoError(t, err)
				assert.Equal(t, "sample.mp3", metadata.Name)
				assert.Equal(t, ".mp3", metadata.Extension)
				assert.False(t, metadata.ExtMismatch)
				assert.NotEmpty(t, metadata.Types)
				assert.NotEmpty(t, metadata.Exif)
				assert.Contains(t, metadata.Exif, "FileType")
				assert.Equal(t, "MP3", metadata.Exif["FileType"])
			},
		},
		{
			name:     "PDF File",
			fileName: "sample.doc",
			checkFn: func(t *testing.T, metadata Metadata, err error) {
				require.NoError(t, err)
				assert.Equal(t, "sample.doc", metadata.Name)
				assert.Equal(t, ".doc", metadata.Extension)
				assert.True(t, metadata.ExtMismatch)
				assert.NotEmpty(t, metadata.Types)
				assert.NotEmpty(t, metadata.Exif)
				assert.Contains(t, metadata.Exif, "FileType")
				assert.Equal(t, "PDF", metadata.Exif["FileType"])
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			filePath := filepath.Join("testdata", tc.fileName)
			metadata, err := extractor.Extract(filePath)
			tc.checkFn(t, metadata, err)
		})
	}
}

func TestMetaExtractor_Errors(t *testing.T) {
	extractor := NewMetaExtractor(Options{})

	t.Run("No File Specified", func(t *testing.T) {
		_, err := extractor.Extract("")
		assert.ErrorIs(t, err, ErrNoFileSpecified)
	})

	t.Run("File Not Found", func(t *testing.T) {
		_, err := extractor.Extract("nonexistent_file")
		assert.ErrorIs(t, err, ErrFileNotFound)
	})
}

func TestGetFileTimes(t *testing.T) {
	tempFile, err := os.CreateTemp("", "test_file_times")
	require.NoError(t, err)
	defer os.Remove(tempFile.Name())

	fileTime, err := getFileTimes(tempFile.Name())
	require.NoError(t, err)

	assert.False(t, fileTime.ModTime.IsZero())
	assert.False(t, fileTime.AccessTime.IsZero())

	// ChangeTime and BirthTime might not be available on all systems
	if fileTime.ChangeTime.IsZero() {
		t.Log("ChangeTime is not available on this system")
	}
	if fileTime.BirthTime.IsZero() {
		t.Log("BirthTime is not available on this system")
	}
}
