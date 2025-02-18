package pdfripper

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"sync"
)

// Extractor holds configuration for PDF extraction.
type Extractor struct {
	PDFFile      string // Path to the input PDF file.
	OutputDir    string // Directory to store extracted pages.
	ProcessCount int    // Number of concurrent workers to use.
}

// NewExtractor creates a new Extractor instance.
// If outputDir is empty, it defaults to a directory named after the PDF file (without extension).
// If processCount is less than 1, it defaults to the number of available CPU cores.
func NewExtractor(pdfFile, outputDir string, processCount int) (*Extractor, error) {
	if pdfFile == "" {
		return nil, errors.New("input PDF file must be specified")
	}

	if outputDir == "" {
		base := filepath.Base(pdfFile)
		ext := filepath.Ext(base)
		outputDir = base[:len(base)-len(ext)]
	}

	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return nil, fmt.Errorf("creating output directory: %w", err)
	}

	if processCount < 1 {
		processCount = runtime.NumCPU()
	}

	return &Extractor{
		PDFFile:      pdfFile,
		OutputDir:    outputDir,
		ProcessCount: processCount,
	}, nil
}

// getTotalPages uses the system-installed pdfinfo command to determine the number of pages.
func (e *Extractor) getTotalPages() (int, error) {
	cmd := exec.Command("pdfinfo", e.PDFFile)
	out, err := cmd.Output()
	if err != nil {
		return 0, fmt.Errorf("running pdfinfo: %w", err)
	}

	lines := strings.Split(string(out), "\n")
	for _, line := range lines {
		if strings.HasPrefix(line, "Pages:") {
			parts := strings.Fields(line)
			if len(parts) >= 2 {
				pages, err := strconv.Atoi(parts[1])
				if err != nil {
					return 0, fmt.Errorf("parsing pages count: %w", err)
				}
				return pages, nil
			}
		}
	}
	return 0, errors.New("could not determine number of pages from pdfinfo output")
}

// ExtractPages extracts text from each page using pdftotext and saves each page to a separate file.
func (e *Extractor) ExtractPages() error {
	totalPages, err := e.getTotalPages()
	if err != nil {
		return fmt.Errorf("getting total pages: %w", err)
	}
	fmt.Printf("Total pages: %d\n", totalPages)

	// Create a channel to distribute page numbers (1-indexed) to workers.
	pagesChan := make(chan int, totalPages)
	var wg sync.WaitGroup
	var mu sync.Mutex
	var firstErr error

	workerCount := e.ProcessCount
	if workerCount > totalPages {
		workerCount = totalPages
	}

	// Launch worker goroutines.
	for i := 0; i < workerCount; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for page := range pagesChan {
				outputFile := filepath.Join(e.OutputDir, fmt.Sprintf("page_%d.txt", page))
				// Use pdftotext to extract one page:
				// -f <page> sets the first page and -l <page> sets the last page.
				cmd := exec.Command("pdftotext", "-f", strconv.Itoa(page), "-l", strconv.Itoa(page), e.PDFFile, outputFile)
				if err := cmd.Run(); err != nil {
					mu.Lock()
					if firstErr == nil {
						firstErr = fmt.Errorf("extracting page %d: %w", page, err)
					}
					mu.Unlock()
					continue
				}
				fmt.Printf("Saved page %d to %s\n", page, outputFile)
			}
		}()
	}

	// Enqueue page numbers.
	for i := 1; i <= totalPages; i++ {
		pagesChan <- i
	}
	close(pagesChan)

	wg.Wait()
	return firstErr
}
