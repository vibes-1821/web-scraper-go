package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/url"
	"os"
	"sync"
	"time"

	"github.com/gocolly/colly/v2"
	"github.com/gocolly/colly/v2/proxy"
)

// ProductDetail represents detailed product information
type ProductDetail struct {
	URL         string    `json:"url"`
	Name        string    `json:"name"`
	Price       string    `json:"price"`
	Description string    `json:"description"`
	SKU         string    `json:"sku"`
	Category    string    `json:"category"`
	ImageURL    string    `json:"image_url"`
	InStock     bool      `json:"in_stock"`
	ScrapedAt   time.Time `json:"scraped_at"`
}

// Scraper holds the scraper configuration and state
type Scraper struct {
	collector   *colly.Collector
	detailCollector *colly.Collector
	products    []ProductDetail
	mu          sync.Mutex
	visited     map[string]bool
}

// NewScraper creates a new scraper with advanced configuration
func NewScraper(allowedDomains []string) *Scraper {
	s := &Scraper{
		products: make([]ProductDetail, 0),
		visited:  make(map[string]bool),
	}

	// Main collector for listing pages
	s.collector = colly.NewCollector(
		colly.AllowedDomains(allowedDomains...),
		colly.MaxDepth(3),
		colly.Async(true), // Enable async for parallel scraping
		colly.CacheDir("./cache"),
	)

	// Separate collector for detail pages (for more granular control)
	s.detailCollector = s.collector.Clone()

	// Configure rate limiting
	s.collector.Limit(&colly.LimitRule{
		DomainGlob:  "*",
		Parallelism: 4, // Allow 4 concurrent requests
		Delay:       500 * time.Millisecond,
		RandomDelay: 500 * time.Millisecond, // Random delay to seem more human
	})

	s.detailCollector.Limit(&colly.LimitRule{
		DomainGlob:  "*",
		Parallelism: 2,
		Delay:       1 * time.Second,
	})

	s.setupCallbacks()

	return s
}

// SetProxy configures proxy rotation for the scraper
func (s *Scraper) SetProxy(proxyURLs []string) error {
	if len(proxyURLs) == 0 {
		return nil
	}

	rp, err := proxy.RoundRobinProxySwitcher(proxyURLs...)
	if err != nil {
		return fmt.Errorf("failed to create proxy switcher: %w", err)
	}

	s.collector.SetProxyFunc(rp)
	s.detailCollector.SetProxyFunc(rp)

	return nil
}

// setupCallbacks configures all the collector callbacks
func (s *Scraper) setupCallbacks() {
	// Set headers to avoid detection
	s.collector.OnRequest(func(r *colly.Request) {
		s.setHeaders(r)
		log.Printf("[LIST] Visiting: %s", r.URL)
	})

	s.detailCollector.OnRequest(func(r *colly.Request) {
		s.setHeaders(r)
		log.Printf("[DETAIL] Visiting: %s", r.URL)
	})

	// Handle errors with retry logic
	s.collector.OnError(func(r *colly.Response, err error) {
		log.Printf("[ERROR] %s: %v (Status: %d)", r.Request.URL, err, r.StatusCode)
		
		// Retry on certain errors
		if r.StatusCode == 429 || r.StatusCode == 503 {
			log.Printf("[RETRY] Will retry %s after delay", r.Request.URL)
			time.Sleep(5 * time.Second)
			r.Request.Retry()
		}
	})

	s.detailCollector.OnError(func(r *colly.Response, err error) {
		log.Printf("[ERROR] Detail page %s: %v", r.Request.URL, err)
	})

	// Parse product listings
	s.collector.OnHTML("li.product", func(e *colly.HTMLElement) {
		productURL := e.ChildAttr("a.woocommerce-LoopProduct-link", "href")
		
		s.mu.Lock()
		if !s.visited[productURL] && productURL != "" {
			s.visited[productURL] = true
			s.mu.Unlock()
			
			// Visit detail page with the detail collector
			s.detailCollector.Visit(productURL)
		} else {
			s.mu.Unlock()
		}
	})

	// Handle pagination
	s.collector.OnHTML("a.next.page-numbers", func(e *colly.HTMLElement) {
		nextURL := e.Attr("href")
		if nextURL != "" {
			e.Request.Visit(nextURL)
		}
	})

	// Parse product detail pages
	s.detailCollector.OnHTML("div.product", func(e *colly.HTMLElement) {
		product := ProductDetail{
			URL:         e.Request.URL.String(),
			Name:        e.ChildText("h1.product_title"),
			Price:       e.ChildText("p.price span.woocommerce-Price-amount"),
			Description: e.ChildText("div.woocommerce-product-details__short-description"),
			SKU:         e.ChildText("span.sku"),
			Category:    e.ChildText("span.posted_in a"),
			ImageURL:    e.ChildAttr("img.wp-post-image", "src"),
			InStock:     e.ChildAttr("p.stock", "class") != "out-of-stock",
			ScrapedAt:   time.Now(),
		}

		if product.Name != "" {
			s.mu.Lock()
			s.products = append(s.products, product)
			s.mu.Unlock()
			log.Printf("[FOUND] %s - %s", product.Name, product.Price)
		}
	})
}

// setHeaders sets browser-like headers on requests
func (s *Scraper) setHeaders(r *colly.Request) {
	r.Headers.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36")
	r.Headers.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8")
	r.Headers.Set("Accept-Language", "en-US,en;q=0.9")
	r.Headers.Set("Accept-Encoding", "gzip, deflate, br")
	r.Headers.Set("Connection", "keep-alive")
	r.Headers.Set("Upgrade-Insecure-Requests", "1")
	r.Headers.Set("Sec-Fetch-Dest", "document")
	r.Headers.Set("Sec-Fetch-Mode", "navigate")
	r.Headers.Set("Sec-Fetch-Site", "none")
	r.Headers.Set("Cache-Control", "max-age=0")
}

// Scrape starts the scraping process from the given URL
func (s *Scraper) Scrape(startURL string) error {
	// Validate URL
	_, err := url.Parse(startURL)
	if err != nil {
		return fmt.Errorf("invalid URL: %w", err)
	}

	log.Printf("Starting scrape from: %s", startURL)
	
	err = s.collector.Visit(startURL)
	if err != nil {
		return fmt.Errorf("failed to visit start URL: %w", err)
	}

	// Wait for async collectors to finish
	s.collector.Wait()
	s.detailCollector.Wait()

	return nil
}

// GetProducts returns the scraped products
func (s *Scraper) GetProducts() []ProductDetail {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.products
}

// ExportToJSON exports scraped data to a JSON file
func (s *Scraper) ExportToJSON(filename string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	file, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	
	if err := encoder.Encode(s.products); err != nil {
		return fmt.Errorf("failed to encode JSON: %w", err)
	}

	return nil
}

// Example usage demonstrating the advanced scraper
func runAdvancedExample() {
	fmt.Println("=== Advanced Go Web Scraper ===")
	fmt.Println("Features: Parallel scraping, Rate limiting, Proxy support, JSON export")
	fmt.Println()

	// Create scraper with allowed domains
	scraper := NewScraper([]string{"scrapingcourse.com"})

	// Optional: Configure proxies (uncomment to use)
	// proxies := []string{
	// 	"http://proxy1.example.com:8080",
	// 	"http://proxy2.example.com:8080",
	// }
	// scraper.SetProxy(proxies)

	// Start scraping
	startTime := time.Now()
	err := scraper.Scrape("https://scrapingcourse.com/ecommerce/")
	if err != nil {
		log.Fatal("Scraping failed:", err)
	}

	elapsed := time.Since(startTime)
	products := scraper.GetProducts()

	fmt.Printf("\n=== Results ===\n")
	fmt.Printf("Products found: %d\n", len(products))
	fmt.Printf("Time elapsed: %s\n", elapsed)

	// Export to JSON
	if len(products) > 0 {
		err = scraper.ExportToJSON("products_detailed.json")
		if err != nil {
			log.Printf("Failed to export JSON: %v", err)
		} else {
			fmt.Println("Data exported to products_detailed.json")
		}
	}
}
