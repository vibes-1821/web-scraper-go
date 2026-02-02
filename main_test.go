package main

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCleanPrice(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "removes leading and trailing whitespace",
			input:    "  $19.99  ",
			expected: "$19.99",
		},
		{
			name:     "handles newlines",
			input:    "$19.99\n$29.99",
			expected: "$19.99 $29.99",
		},
		{
			name:     "handles tabs",
			input:    "$19.99\t\tUSD",
			expected: "$19.99USD",
		},
		{
			name:     "collapses multiple spaces",
			input:    "$19.99    USD",
			expected: "$19.99 USD",
		},
		{
			name:     "handles empty string",
			input:    "",
			expected: "",
		},
		{
			name:     "handles whitespace only",
			input:    "   \n\t  ",
			expected: "",
		},
		{
			name:     "handles complex formatting",
			input:    "\n\t  $19.99  -  $29.99  \n",
			expected: "$19.99 - $29.99",
		},
		{
			name:     "preserves single space",
			input:    "$19.99 USD",
			expected: "$19.99 USD",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := cleanPrice(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestExportToCSV(t *testing.T) {
	t.Run("creates valid CSV file", func(t *testing.T) {
		tmpDir := t.TempDir()
		filename := filepath.Join(tmpDir, "test.csv")

		products := []Product{
			{
				Name:      "Test Product 1",
				Price:     "$19.99",
				URL:       "http://example.com/product1",
				Image:     "http://example.com/image1.jpg",
				ScrapedAt: time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC),
			},
		}

		err := exportToCSV(products, filename)
		require.NoError(t, err)
		assert.FileExists(t, filename)
	})

	t.Run("writes correct headers", func(t *testing.T) {
		tmpDir := t.TempDir()
		filename := filepath.Join(tmpDir, "test.csv")

		products := []Product{}
		err := exportToCSV(products, filename)
		require.NoError(t, err)

		rows, err := ReadCSVFile(filename)
		require.NoError(t, err)
		require.Len(t, rows, 1) // Only header

		expectedHeaders := []string{"Name", "Price", "URL", "Image", "Scraped At"}
		assert.Equal(t, expectedHeaders, rows[0])
	})

	t.Run("formats data correctly", func(t *testing.T) {
		tmpDir := t.TempDir()
		filename := filepath.Join(tmpDir, "test.csv")

		testTime := time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC)
		products := []Product{
			{
				Name:      "Test Product 1",
				Price:     "$19.99",
				URL:       "http://example.com/product1",
				Image:     "http://example.com/image1.jpg",
				ScrapedAt: testTime,
			},
			{
				Name:      "Test Product 2",
				Price:     "$29.99",
				URL:       "http://example.com/product2",
				Image:     "http://example.com/image2.jpg",
				ScrapedAt: testTime,
			},
		}

		err := exportToCSV(products, filename)
		require.NoError(t, err)

		rows, err := ReadCSVFile(filename)
		require.NoError(t, err)
		require.Len(t, rows, 3) // Header + 2 products

		// Check first product row
		assert.Equal(t, "Test Product 1", rows[1][0])
		assert.Equal(t, "$19.99", rows[1][1])
		assert.Equal(t, "http://example.com/product1", rows[1][2])
		assert.Equal(t, "http://example.com/image1.jpg", rows[1][3])
		assert.Equal(t, testTime.Format(time.RFC3339), rows[1][4])
	})

	t.Run("handles empty product list", func(t *testing.T) {
		tmpDir := t.TempDir()
		filename := filepath.Join(tmpDir, "test.csv")

		products := []Product{}
		err := exportToCSV(products, filename)
		require.NoError(t, err)

		rows, err := ReadCSVFile(filename)
		require.NoError(t, err)
		assert.Len(t, rows, 1) // Only header
	})

	t.Run("handles special characters in fields", func(t *testing.T) {
		tmpDir := t.TempDir()
		filename := filepath.Join(tmpDir, "test.csv")

		products := []Product{
			{
				Name:      "Product with \"quotes\" and, commas",
				Price:     "$19.99",
				URL:       "http://example.com/product",
				Image:     "http://example.com/image.jpg",
				ScrapedAt: time.Now(),
			},
		}

		err := exportToCSV(products, filename)
		require.NoError(t, err)

		rows, err := ReadCSVFile(filename)
		require.NoError(t, err)
		assert.Equal(t, "Product with \"quotes\" and, commas", rows[1][0])
	})

	t.Run("handles file write errors", func(t *testing.T) {
		// Try to write to an invalid path
		filename := "/invalid/path/that/does/not/exist/test.csv"

		products := []Product{
			{Name: "Test", Price: "$19.99"},
		}

		err := exportToCSV(products, filename)
		assert.Error(t, err)
	})
}

func TestProductStruct(t *testing.T) {
	t.Run("creates product with all fields", func(t *testing.T) {
		now := time.Now()
		product := Product{
			URL:       "http://example.com/product",
			Image:     "http://example.com/image.jpg",
			Name:      "Test Product",
			Price:     "$19.99",
			ScrapedAt: now,
		}

		assert.Equal(t, "http://example.com/product", product.URL)
		assert.Equal(t, "http://example.com/image.jpg", product.Image)
		assert.Equal(t, "Test Product", product.Name)
		assert.Equal(t, "$19.99", product.Price)
		assert.Equal(t, now, product.ScrapedAt)
	})

	t.Run("handles empty fields", func(t *testing.T) {
		product := Product{}

		assert.Equal(t, "", product.URL)
		assert.Equal(t, "", product.Image)
		assert.Equal(t, "", product.Name)
		assert.Equal(t, "", product.Price)
		assert.True(t, product.ScrapedAt.IsZero())
	})
}

func TestExportToCSVIntegration(t *testing.T) {
	t.Run("creates file that can be read back", func(t *testing.T) {
		tmpDir := t.TempDir()
		filename := filepath.Join(tmpDir, "products.csv")

		testTime := time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC)
		originalProducts := []Product{
			{
				Name:      "Product A",
				Price:     "$10.00",
				URL:       "http://example.com/a",
				Image:     "http://example.com/a.jpg",
				ScrapedAt: testTime,
			},
			{
				Name:      "Product B",
				Price:     "$20.00",
				URL:       "http://example.com/b",
				Image:     "http://example.com/b.jpg",
				ScrapedAt: testTime,
			},
		}

		// Export to CSV
		err := exportToCSV(originalProducts, filename)
		require.NoError(t, err)

		// Read back and verify
		rows, err := ReadCSVFile(filename)
		require.NoError(t, err)
		require.Len(t, rows, 3) // Header + 2 products

		// Verify header
		assert.Equal(t, []string{"Name", "Price", "URL", "Image", "Scraped At"}, rows[0])

		// Verify data
		assert.Equal(t, "Product A", rows[1][0])
		assert.Equal(t, "$10.00", rows[1][1])
		assert.Equal(t, "Product B", rows[2][0])
		assert.Equal(t, "$20.00", rows[2][1])
	})
}

func TestCleanPriceEdgeCases(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "unicode characters",
			input:    "€19.99",
			expected: "€19.99",
		},
		{
			name:     "multiple newlines",
			input:    "\n\n\n$19.99\n\n",
			expected: "$19.99",
		},
		{
			name:     "mixed whitespace types",
			input:    " \t\n$19.99 \n\t ",
			expected: "$19.99",
		},
		{
			name:     "price range",
			input:    "$19.99  -  $29.99",
			expected: "$19.99 - $29.99",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := cleanPrice(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// Test that CSV file is properly closed
func TestExportToCSVFileHandling(t *testing.T) {
	tmpDir := t.TempDir()
	filename := filepath.Join(tmpDir, "test.csv")

	products := []Product{
		{Name: "Test", Price: "$19.99", ScrapedAt: time.Now()},
	}

	err := exportToCSV(products, filename)
	require.NoError(t, err)

	// If file wasn't closed properly, we wouldn't be able to open it again
	file, err := os.Open(filename)
	require.NoError(t, err)
	file.Close()

	// Also try to remove it
	err = os.Remove(filename)
	assert.NoError(t, err)
}
