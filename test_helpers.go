package main

import (
	"encoding/csv"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

// CreateMockServer creates a mock HTTP server with a custom handler
func CreateMockServer(handler http.Handler) *httptest.Server {
	return httptest.NewServer(handler)
}

// CreateMockServerWithRoutes creates a mock server with specific routes
func CreateMockServerWithRoutes(routes map[string]string) *httptest.Server {
	mux := http.NewServeMux()

	for path, content := range routes {
		path := path       // capture loop variable
		content := content // capture loop variable
		mux.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "text/html")
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(content))
		})
	}

	return httptest.NewServer(mux)
}

// GetFixture reads a test fixture from the testdata directory
func GetFixture(filename string) (string, error) {
	data, err := os.ReadFile(filepath.Join("testdata", filename))
	if err != nil {
		return "", fmt.Errorf("failed to read fixture %s: %w", filename, err)
	}
	return string(data), nil
}

// MustGetFixture reads a fixture or panics (for test setup)
func MustGetFixture(t *testing.T, filename string) string {
	t.Helper()
	content, err := GetFixture(filename)
	if err != nil {
		t.Fatalf("Failed to load fixture %s: %v", filename, err)
	}
	return content
}

// CompareCSV compares two CSV files for equality
func CompareCSV(t *testing.T, expectedFile, actualFile string) {
	t.Helper()

	expected, err := ReadCSVFile(expectedFile)
	assert.NoError(t, err, "Failed to read expected CSV")

	actual, err := ReadCSVFile(actualFile)
	assert.NoError(t, err, "Failed to read actual CSV")

	assert.Equal(t, len(expected), len(actual), "CSV row count mismatch")

	for i := range expected {
		assert.Equal(t, expected[i], actual[i], "CSV row %d mismatch", i)
	}
}

// ReadCSVFile reads a CSV file and returns all rows
func ReadCSVFile(filename string) ([][]string, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	reader := csv.NewReader(file)
	return reader.ReadAll()
}

// CreateTempCSV creates a temporary CSV file with given data
func CreateTempCSV(t *testing.T, data [][]string) string {
	t.Helper()

	tmpDir := t.TempDir()
	filename := filepath.Join(tmpDir, "test.csv")

	file, err := os.Create(filename)
	assert.NoError(t, err, "Failed to create temp CSV file")
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	for _, row := range data {
		err := writer.Write(row)
		assert.NoError(t, err, "Failed to write CSV row")
	}

	return filename
}

// FileExists checks if a file exists
func FileExists(filename string) bool {
	_, err := os.Stat(filename)
	return err == nil
}

// CreateMockServerWithStatus creates a server that returns a specific status code
func CreateMockServerWithStatus(statusCode int, content string) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		w.WriteHeader(statusCode)
		w.Write([]byte(content))
	}))
}

// ExtractDomain extracts the domain (including port) from a server URL for use with collectors
func ExtractDomain(serverURL string) string {
	// Remove http:// or https://
	domain := serverURL
	if len(domain) > 7 && domain[:7] == "http://" {
		domain = domain[7:]
	} else if len(domain) > 8 && domain[:8] == "https://" {
		domain = domain[8:]
	}
	// Remove any path components, keeping just host:port
	if idx := strings.Index(domain, "/"); idx != -1 {
		domain = domain[:idx]
	}
	return domain
}
