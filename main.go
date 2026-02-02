package main

import (
	"encoding/csv"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/gocolly/colly/v2"
)

// Product represents a scraped product item
type Product struct {
	URL      string
	Image    string
	Name     string
	Price    string
	ScrapedAt time.Time
}

func main() {
	fmt.Println("Starting Go Web Scraper...")
	fmt.Println("Target: https://scrapingcourse.com/ecommerce/")

	// Slice to store scraped products
	var products []Product

	// Create a new collector with configuration
	c := colly.NewCollector(
		// Only allow scraping from the target domain
		colly.AllowedDomains("scrapingcourse.com"),
		// Enable URL revisiting (useful for pagination)
		colly.AllowURLRevisit(),
		// Set max depth for crawling
		colly.MaxDepth(2),
		// Cache responses to avoid repeated requests during development
		colly.CacheDir("./cache"),
	)

	// Set rate limiting to be a good citizen
	c.Limit(&colly.LimitRule{
		DomainGlob:  "*",
		Parallelism: 2,
		Delay:       1 * time.Second,
	})

	// Set custom headers to avoid being blocked
	c.OnRequest(func(r *colly.Request) {
		r.Headers.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36")
		r.Headers.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,*/*;q=0.8")
		r.Headers.Set("Accept-Language", "en-US,en;q=0.5")
		fmt.Printf("Visiting: %s\n", r.URL)
	})

	// Handle response errors
	c.OnError(func(r *colly.Response, err error) {
		log.Printf("Error scraping %s: %v", r.Request.URL, err)
	})

	// Handle successful responses
	c.OnResponse(func(r *colly.Response) {
		fmt.Printf("Got response from: %s [Status: %d]\n", r.Request.URL, r.StatusCode)
	})

	// Scrape product items from the product listing
	// Each product is in a <li> element with class "product"
	c.OnHTML("li.product", func(e *colly.HTMLElement) {
		product := Product{
			URL:       e.ChildAttr("a.woocommerce-LoopProduct-link", "href"),
			Image:     e.ChildAttr("img.product-image", "src"),
			Name:      e.ChildText("h2.woocommerce-loop-product__title"),
			Price:     cleanPrice(e.ChildText("span.price")),
			ScrapedAt: time.Now(),
		}

		// Only add if we got valid data
		if product.Name != "" {
			products = append(products, product)
			fmt.Printf("Found product: %s - %s\n", product.Name, product.Price)
		}
	})

	// Handle pagination - find and visit "next" page links
	c.OnHTML("a.next.page-numbers", func(e *colly.HTMLElement) {
		nextPage := e.Attr("href")
		if nextPage != "" {
			fmt.Printf("Found next page: %s\n", nextPage)
			e.Request.Visit(nextPage)
		}
	})

	// Callback when scraping is complete for a page
	c.OnScraped(func(r *colly.Response) {
		fmt.Printf("Finished scraping: %s\n", r.Request.URL)
	})

	// Start scraping from the main e-commerce page
	err := c.Visit("https://scrapingcourse.com/ecommerce/")
	if err != nil {
		log.Fatal("Failed to start scraping:", err)
	}

	// Wait for all requests to complete
	c.Wait()

	// Export results to CSV
	if len(products) > 0 {
		err = exportToCSV(products, "products.csv")
		if err != nil {
			log.Fatal("Failed to export to CSV:", err)
		}
		fmt.Printf("\nScraping complete! Found %d products.\n", len(products))
		fmt.Println("Data exported to products.csv")
	} else {
		fmt.Println("No products found.")
	}
}

// cleanPrice removes extra whitespace and normalizes price strings
func cleanPrice(price string) string {
	// Remove extra whitespace and newlines
	price = strings.TrimSpace(price)
	price = strings.ReplaceAll(price, "\n", " ")
	price = strings.ReplaceAll(price, "\t", "")
	
	// Collapse multiple spaces into one
	for strings.Contains(price, "  ") {
		price = strings.ReplaceAll(price, "  ", " ")
	}
	
	return price
}

// exportToCSV writes the scraped products to a CSV file
func exportToCSV(products []Product, filename string) error {
	file, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	// Write header row
	header := []string{"Name", "Price", "URL", "Image", "Scraped At"}
	if err := writer.Write(header); err != nil {
		return fmt.Errorf("failed to write header: %w", err)
	}

	// Write product data
	for _, product := range products {
		row := []string{
			product.Name,
			product.Price,
			product.URL,
			product.Image,
			product.ScrapedAt.Format(time.RFC3339),
		}
		if err := writer.Write(row); err != nil {
			return fmt.Errorf("failed to write row: %w", err)
		}
	}

	return nil
}
