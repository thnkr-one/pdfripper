package main

import (
	"flag"
	"fmt"
	"log"
	"runtime"

	"github.com/thnkr-one/pdfripper/pdfripper"
)

func main() {
	inputFile := flag.String("input", "", "Input PDF file path (required)")
	outputDir := flag.String("output", "", "Output directory (default: PDF basename)")
	procCount := flag.Int("processes", 0, "Number of concurrent workers (default: number of CPU cores)")
	flag.Parse()

	if *inputFile == "" {
		flag.Usage()
		log.Fatal("Error: input PDF file is required (use -input)")
	}

	if *procCount < 1 {
		*procCount = runtime.NumCPU()
	}

	extractor, err := pdfripper.NewExtractor(*inputFile, *outputDir, *procCount)
	if err != nil {
		log.Fatalf("Error initializing extractor: %v", err)
	}

	if err := extractor.ExtractPages(); err != nil {
		log.Fatalf("Error extracting pages: %v", err)
	}

	fmt.Println("Extraction complete.")
}
