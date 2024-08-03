// Package metaextractor provides functionality for extracting comprehensive metadata from files.
// It combines file system information, TrID file type analysis, and EXIF data extraction.
package metaextractor

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/attilabuti/trid"
	"github.com/barasher/go-exiftool"
	"github.com/djherbis/times"
)

// MetaExtractor represents a metadata extraction instance with specific configurations.
type MetaExtractor struct {
	trid         *trid.Trid
	tridMatches  int
	exifToolOpts []func(*exiftool.Exiftool) error
}

// Options configures the metadata extraction parameters.
type Options struct {
	// TridPath is the file system path to the TrID executable.
	TridPath string

	// TridDefs is the file system path to the TrID definitions package.
	TridDefs string

	// TridTimeout is the maximum duration allowed for TrID execution.
	TridTimeout time.Duration

	// TridMatches specifies the maximum number of file type matches to return from TrID.
	TridMatches int

	// ExifToolPath is the file system path to the ExifTool executable.
	ExifToolPath string
}

// Metadata contains comprehensive metadata extracted from a file.
type Metadata struct {
	// Name is the base name of the file, including the extension.
	Name string

	// Extension is the file extension (e.g., ".txt", ".pdf").
	Extension string

	// ExtMismatch indicates whether the file's extension differs from its detected type.
	ExtMismatch bool

	// Size is the file size in bytes.
	Size int64

	// Time contains various timestamps associated with the file.
	Time FileTime

	// Types is a slice of detected file types.
	// The first element (if present) is considered the most likely file type.
	Types []trid.FileType

	// Exif contains extracted EXIF metadata from the file.
	Exif ExifMetadata
}

// FileTime represents various timestamps associated with a file.
type FileTime struct {
	// ModTime is the last modification time of the file.
	ModTime time.Time

	// AccessTime is the last access time of the file.
	AccessTime time.Time

	// ChangeTime is the last status change time of the file.
	// This can differ from ModTime as it includes changes to permissions, ownership, etc.
	ChangeTime time.Time

	// BirthTime is the creation time of the file.
	// Note: This may not be available on all file systems.
	BirthTime time.Time
}

// ExifMetadata is a map of EXIF metadata extracted from a file.
type ExifMetadata map[string]interface{}

var (
	// ErrNoFileSpecified is returned when no file path is provided.
	ErrNoFileSpecified = errors.New("no file specified")

	// ErrFileNotFound is returned when the specified file cannot be located or accessed.
	ErrFileNotFound = errors.New("file not found")

	// ErrNoMetadataExtracted indicates that no metadata could be extracted from the file.
	ErrNoMetadataExtracted = errors.New("no metadata extracted")
)

// NewMetaExtractor creates a new MetaExtractor instance with the given options.
func NewMetaExtractor(opts Options) *MetaExtractor {
	if opts.TridMatches <= 0 {
		opts.TridMatches = 5 // Default to 5 matches if not specified
	}

	exifToolOpts := []func(*exiftool.Exiftool) error{
		exiftool.ExtractAllBinaryMetadata(),
		exiftool.ExtractEmbedded(),
	}

	if opts.ExifToolPath != "" {
		exifToolOpts = append(exifToolOpts, exiftool.SetExiftoolBinaryPath(opts.ExifToolPath))
	}

	return &MetaExtractor{
		trid: trid.NewTrid(trid.Options{
			Cmd:         opts.TridPath,
			Definitions: opts.TridDefs,
			Timeout:     opts.TridTimeout,
		}),
		tridMatches:  opts.TridMatches,
		exifToolOpts: exifToolOpts,
	}
}

// Extract examines the given file, extracting its metadata, determining its
// type, and gathering EXIF information if available. It returns a Metadata
// struct or an error.
func (me *MetaExtractor) Extract(filePath string) (Metadata, error) {
	var metadata Metadata

	if filePath == "" {
		return metadata, ErrNoFileSpecified
	}

	fileInfo, err := os.Stat(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return metadata, ErrFileNotFound
		}
		return metadata, err
	}

	metadata.Name = filepath.Base(filePath)
	metadata.Extension = strings.ToLower(filepath.Ext(filePath))
	metadata.Size = fileInfo.Size()

	if fileTime, err := getFileTimes(filePath); err == nil {
		metadata.Time = fileTime
	} else {
		return metadata, err
	}

	if fileTypes, err := me.tridAnalysis(filePath); err == nil {
		metadata.Types = fileTypes

		if len(fileTypes) > 0 {
			if strings.Contains(fileTypes[0].Extension, "/") {
				metadata.ExtMismatch = true
				es := strings.Split(strings.ReplaceAll(fileTypes[0].Extension, ".", ""), "/")
				for _, e := range es {
					if "."+e == metadata.Extension {
						metadata.ExtMismatch = false
						break
					}
				}
			} else if metadata.Extension != fileTypes[0].Extension {
				metadata.ExtMismatch = true
			}
		}
	} else {
		return metadata, err
	}

	if exifData, err := me.extractExifData(filePath); err == nil {
		metadata.Exif = exifData
	} else if errors.Is(err, ErrNoMetadataExtracted) {
		metadata.Exif = ExifMetadata{}
	} else {
		return metadata, err
	}

	return metadata, nil
}

// getFileTimes retrieves various timestamps associated with the file.
func getFileTimes(filePath string) (FileTime, error) {
	t, err := times.Stat(filePath)
	if err != nil {
		return FileTime{}, err
	}

	var changeTime time.Time
	if t.HasChangeTime() {
		changeTime = t.ChangeTime()
	}

	var birthTime time.Time
	if t.HasBirthTime() {
		birthTime = t.BirthTime()
	}

	return FileTime{
		AccessTime: t.AccessTime(),
		ModTime:    t.ModTime(),
		ChangeTime: changeTime,
		BirthTime:  birthTime,
	}, nil
}

// tridAnalysis performs file type analysis using TrID.
// It returns a slice of possible file types, sorted by likelihood.
func (me *MetaExtractor) tridAnalysis(filePath string) ([]trid.FileType, error) {
	return me.trid.Scan(filePath, me.tridMatches)
}

// extractExifData extracts EXIF metadata from the file using ExifTool.
// It returns a map of metadata fields or an error if extraction fails.
func (me *MetaExtractor) extractExifData(filePath string) (ExifMetadata, error) {
	et, err := exiftool.NewExiftool(me.exifToolOpts...)
	if err != nil {
		return nil, fmt.Errorf("error initializing ExifTool: %v", err)
	}
	defer et.Close()

	fileInfos := et.ExtractMetadata(filePath)
	if len(fileInfos) == 0 {
		return nil, ErrNoMetadataExtracted
	}

	if fileInfos[0].Err != nil {
		return nil, fmt.Errorf("error extracting metadata: %v", fileInfos[0].Err)
	}

	return fileInfos[0].Fields, nil
}
