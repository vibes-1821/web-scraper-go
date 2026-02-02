package main

import (
	"fmt"
	"log"
	"net/url"
	"strings"
	"sync"

	"github.com/gocolly/colly/v2"
)

// WebCrawler implements a basic web crawler that follows links
type WebCrawler struct {
	collector    *colly.Collector
	visitedURLs  map[string]bool
	foundLinks   []string
	mu           sync.Mutex
	maxPages     int
	pagesVisited int
}

// NewWebCrawler creates a new web crawler
func NewWebCrawler(allowedDomains []string, maxPages int) *WebCrawler {
	wc := &WebCrawler{
		visitedURLs: make(map[string]bool),
		foundLinks:  make([]string, 0),
		maxPages:    maxPages,
	}

	wc.collector = colly.NewCollector(
		colly.AllowedDomains(allowedDomains...),
		colly.MaxDepth(3),
	)

	wc.setupCallbacks()
	return wc
}

func (wc *WebCrawler) setupCallbacks() {
	// Log each request
	wc.collector.OnRequest(func(r *colly.Request) {
		wc.mu.Lock()
		wc.pagesVisited++
		current := wc.pagesVisited
		wc.mu.Unlock()
		
		fmt.Printf("[%d] Crawling: %s\n", current, r.URL)
	})

	// Find and follow all links
	wc.collector.OnHTML("a[href]", func(e *colly.HTMLElement) {
		link := e.Attr("href")
		
		// Skip empty links, anchors, and javascript
		if link == "" || strings.HasPrefix(link, "#") || strings.HasPrefix(link, "javascript:") {
			return
		}

		// Convert to absolute URL
		absoluteURL := e.Request.AbsoluteURL(link)
		if absoluteURL == "" {
			return
		}

		// Normalize URL (remove fragments)
		parsedURL, err := url.Parse(absoluteURL)
		if err != nil {
			return
		}
		parsedURL.Fragment = ""
		normalizedURL := parsedURL.String()

		wc.mu.Lock()
		// Check if we've reached max pages
		if wc.pagesVisited >= wc.maxPages {
			wc.mu.Unlock()
			return
		}
		
		// Track unique links
		if !wc.visitedURLs[normalizedURL] {
			wc.visitedURLs[normalizedURL] = true
			wc.foundLinks = append(wc.foundLinks, normalizedURL)
			wc.mu.Unlock()
			
			// Visit the link
			e.Request.Visit(normalizedURL)
		} else {
			wc.mu.Unlock()
		}
	})

	// Handle errors
	wc.collector.OnError(func(r *colly.Response, err error) {
		log.Printf("Error crawling %s: %v", r.Request.URL, err)
	})

	// Log when a page is fully scraped
	wc.collector.OnScraped(func(r *colly.Response) {
		fmt.Printf("Completed: %s\n", r.Request.URL)
	})
}

// Crawl starts crawling from the given URL
func (wc *WebCrawler) Crawl(startURL string) error {
	return wc.collector.Visit(startURL)
}

// GetFoundLinks returns all discovered links
func (wc *WebCrawler) GetFoundLinks() []string {
	wc.mu.Lock()
	defer wc.mu.Unlock()
	
	// Return a copy
	links := make([]string, len(wc.foundLinks))
	copy(links, wc.foundLinks)
	return links
}

// GetPagesVisited returns the number of pages visited
func (wc *WebCrawler) GetPagesVisited() int {
	wc.mu.Lock()
	defer wc.mu.Unlock()
	return wc.pagesVisited
}

// runCrawlerExample demonstrates the web crawler
func runCrawlerExample() {
	fmt.Println("=== Web Crawler Example ===")
	fmt.Println("Crawling go-colly.org (max 10 pages)")
	fmt.Println()

	crawler := NewWebCrawler([]string{"go-colly.org"}, 10)
	
	err := crawler.Crawl("https://go-colly.org/")
	if err != nil {
		log.Fatal("Crawling failed:", err)
	}

	links := crawler.GetFoundLinks()
	fmt.Printf("\n=== Crawl Results ===\n")
	fmt.Printf("Pages visited: %d\n", crawler.GetPagesVisited())
	fmt.Printf("Links discovered: %d\n", len(links))
	
	fmt.Println("\nSample links found:")
	for i, link := range links {
		if i >= 10 {
			fmt.Printf("... and %d more\n", len(links)-10)
			break
		}
		fmt.Printf("  - %s\n", link)
	}
}
