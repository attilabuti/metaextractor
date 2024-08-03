# MetaExtractor

MetaExtractor is a Go package that provides comprehensive metadata extraction from files. It combines file system information, TrID file type analysis, and EXIF data extraction to give you a complete picture of your files.

## Installation

To use this package, you need to have Go installed on your system. You also need to have the [TrID command-line tool](https://mark0.net/soft-trid-e.html) and [ExifTool](https://exiftool.org/) installed and accessible in your system's PATH.

1. First, install the TrID command-line tool. You can download it from the [official TrID website](https://mark0.net/soft-trid-e.html).

2. Install ExifTool. You can download it from the [official ExifTool website](https://exiftool.org/).

3. Install the MetaExtractor Go package:

```bash
go get github.com/attilabuti/metaextractor
```

## Usage

```go
package main

import (
	"fmt"
	"log"

	"github.com/attilabuti/metaextractor"
)

func main() {
	// Create a new MetaExtractor instance
	me := metaextractor.NewMetaExtractor(metaextractor.Options{
		TridPath:     "/path/to/trid",
		TridDefs:     "/path/to/triddefs.trd",
		TridTimeout:  5 * time.Second,
		TridMatches:  5,
		ExifToolPath: "/path/to/exiftool",
	})

	// Extract metadata from a file
	metadata, err := me.Extract("/path/to/your/file")
	if err != nil {
		log.Fatalf("Error extracting metadata: %v", err)
	}

	// Print the extracted metadata
	fmt.Printf("File Name: %s\n", metadata.Name)
	fmt.Printf("File Size: %d bytes\n", metadata.Size)
	fmt.Printf("File Extension: %s\n", metadata.Extension)
	fmt.Printf("Extension Mismatch: %v\n", metadata.ExtMismatch)
	fmt.Printf("Last Modified: %v\n", metadata.Time.ModTime)

	if len(metadata.Types) > 0 {
		fmt.Printf("Detected File Type: %s (%s)\n", metadata.Types[0].Type, metadata.Types[0].Mime)
	}

	fmt.Println("EXIF Metadata:")
	for key, value := range metadata.Exif {
		fmt.Printf("  %s: %v\n", key, value)
	}
}
```

## Options

The Options struct allows you to configure the MetaExtractor:
- TridPath: Path to the TrID executable
- TridDefs: Path to the TrID definitions file
- TridTimeout: Maximum duration allowed for TrID execution
- TridMatches: Maximum number of file type matches to return from TrID
- ExifToolPath: Path to the ExifTool executable

Make sure to set these paths correctly according to your system configuration.

## Issues

Submit the [issues](https://github.com/attilabuti/metaextractor/issues) if you find any bug or have any suggestion.

## Contribution

Fork the [repo](https://github.com/attilabuti/metaextractor) and submit pull requests.

## License

This extension is licensed under the [MIT License](https://github.com/attilabuti/metaextractor/blob/main/LICENSE).