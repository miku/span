package crossref

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/segmentio/encoding/json"
)

func TestCreateSnapshot(t *testing.T) {
	// Create a temporary directory for test files
	tempDir, err := os.MkdirTemp("", "crossref-test-")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create test input files
	inputFiles, expectedDOIs := createTestInputFiles(t, tempDir)

	// Define output file path
	outputFile := filepath.Join(tempDir, "output.json")

	// Run the snapshot process
	opts := SnapshotOptions{
		InputFiles:     inputFiles,
		OutputFile:     outputFile,
		TempDir:        tempDir,
		BatchSize:      10,
		NumWorkers:     2,
		SortBufferSize: "10%",
		KeepTempFiles:  true,
		Verbose:        true,
	}

	if err := CreateSnapshot(opts); err != nil {
		t.Fatalf("CreateSnapshot failed: %v", err)
	}

	// Verify the output
	verifyOutput(t, outputFile, expectedDOIs)
}

// createTestInputFiles creates test input files with overlapping DOIs
// and returns the list of file paths and expected DOIs in the output
func createTestInputFiles(t *testing.T, dir string) ([]string, map[string]bool) {
	// Create 3 input files with different combinations of the DOIs
	inputContents := [][]string{
		{ // File 1
			createJSONRecord("10.1000/test1", 1610000000, "Content for test1 v1"),
			createJSONRecord("10.1000/test2", 1630000000, "Content for test2 v1"),
			createJSONRecord("10.1000/test3", 1640000000, "Content for test3 v1"),
			createJSONRecord("10.1000/test4", 1650000000, "Content for test4 v1"),
		},
		{ // File 2
			createJSONRecord("10.1000/test1", 1620000000, "Content for test1 v2"),
			createJSONRecord("10.1000/test3", 1630000000, "Content for test3 v2"),
			createJSONRecord("10.1000/test4", 1660000000, "Content for test4 v2"),
		},
		{ // File 3 - Some duplicates with older timestamps
			createJSONRecord("10.1000/test1", 1600000000, "Content for test1 v0"),
			createJSONRecord("10.1000/test3", 1620000000, "Content for test3 v0"),
		},
	}

	// Create the input files
	var inputFiles []string
	for i, contents := range inputContents {
		filename := filepath.Join(dir, fmt.Sprintf("input%d.json", i+1))
		file, err := os.Create(filename)
		if err != nil {
			t.Fatalf("Failed to create test file: %v", err)
		}

		for _, line := range contents {
			fmt.Fprintln(file, line)
		}
		file.Close()
		inputFiles = append(inputFiles, filename)
	}

	// Determine which records should be in the output (the latest for each DOI)
	expectedDOIs := map[string]bool{
		"10.1000/test1:1620000000": true, // Latest version of test1
		"10.1000/test2:1630000000": true, // Latest version of test2
		"10.1000/test3:1640000000": true, // Latest version of test3
		"10.1000/test4:1660000000": true, // Latest version of test4
	}

	return inputFiles, expectedDOIs
}

// createJSONRecord creates a JSON record with the given DOI, timestamp, and content
func createJSONRecord(doi string, timestamp int64, content string) string {
	return fmt.Sprintf(`{"DOI":"%s","indexed":{"timestamp":%d},"content":"%s"}`, doi, timestamp, content)
}

// verifyOutput verifies that the output file contains only the latest version of each DOI
func verifyOutput(t *testing.T, outputFile string, expectedDOIs map[string]bool) {
	// Open the output file
	file, err := os.Open(outputFile)
	if err != nil {
		t.Fatalf("Failed to open output file: %v", err)
	}
	defer file.Close()

	// Read and check each line
	scanner := bufio.NewScanner(file)
	foundDOIs := make(map[string]bool)

	for scanner.Scan() {
		line := scanner.Text()

		// Parse the JSON record
		var record struct {
			DOI     string `json:"DOI"`
			Indexed struct {
				Timestamp int64 `json:"timestamp"`
			} `json:"indexed"`
		}

		if err := json.Unmarshal([]byte(line), &record); err != nil {
			t.Fatalf("Failed to parse JSON record: %v\nLine: %s", err, line)
		}

		// Create the key in the format expected by the test
		key := fmt.Sprintf("%s:%d", record.DOI, record.Indexed.Timestamp)

		// Check if this is an expected DOI:timestamp
		if !expectedDOIs[key] {
			t.Errorf("Unexpected DOI:timestamp in output: %s", key)
		}

		foundDOIs[key] = true
	}

	if err := scanner.Err(); err != nil {
		t.Fatalf("Error reading output file: %v", err)
	}

	// Check if all expected DOIs were found
	for key := range expectedDOIs {
		if !foundDOIs[key] {
			t.Errorf("Expected DOI:timestamp not found in output: %s", key)
		}
	}

	// Check the count
	if len(foundDOIs) != len(expectedDOIs) {
		t.Errorf("Expected %d unique DOIs, found %d", len(expectedDOIs), len(foundDOIs))
	}
}

// TestSortAndFilter specifically tests the sort and filter stage
func TestSortAndFilter(t *testing.T) {
	// Create a temporary directory
	tempDir, err := os.MkdirTemp("", "crossref-sort-test-")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create a sample index file
	indexFile := filepath.Join(tempDir, "index.tsv")
	file, err := os.Create(indexFile)
	if err != nil {
		t.Fatalf("Failed to create index file: %v", err)
	}

	// Write sample data to the index file
	// Format: filename \t lineNumber \t timestamp \t DOI
	samples := []string{
		"file1.json\t0\t1610000000\t10.1000/test1",
		"file2.json\t5\t1620000000\t10.1000/test1",
		"file1.json\t10\t1630000000\t10.1000/test2",
		"file3.json\t15\t1640000000\t10.1000/test3",
		"file2.json\t20\t1630000000\t10.1000/test3",
		"file1.json\t25\t1650000000\t10.1000/test4",
		"file2.json\t30\t1660000000\t10.1000/test4",
	}

	for _, sample := range samples {
		fmt.Fprintln(file, sample)
	}
	file.Close()

	// Create output file for the line numbers
	lineNumsFile := filepath.Join(tempDir, "linenums.txt")

	// Run the sort and filter
	if err := identifyLatestVersions(indexFile, lineNumsFile, "10%", true); err != nil {
		time.Sleep(30 * time.Second)
		t.Fatalf("identifyLatestVersions failed: %v", err)
	}

	// Verify the line numbers file
	content, err := os.ReadFile(lineNumsFile)
	if err != nil {
		t.Fatalf("Failed to read line numbers file: %v", err)
	}

	lines := strings.Split(strings.TrimSpace(string(content)), "\n")
	expected := map[string]bool{
		"file2.json\t5":  true, // test1, latest version
		"file1.json\t10": true, // test2, only version
		"file3.json\t15": true, // test3, latest version
		"file2.json\t30": true, // test4, latest version
	}

	if len(lines) != len(expected) {
		t.Errorf("Expected %d lines, got %d", len(expected), len(lines))
	}

	for _, line := range lines {
		if !expected[line] {
			t.Errorf("Unexpected line in output: %s", line)
		}
	}
}
