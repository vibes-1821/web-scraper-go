# Agent Guide - Go Web Scraper Project

## Quick Start

```bash
# Install dependencies
go mod tidy

# Run basic scraper (outputs products.csv)
go run main.go

# Test advanced scraper
go run -run advanced_scraper.go

# Test crawler
go run -run crawler.go
```

## Project Structure

```
web-scraper-go/
├── main.go             # Basic scraper - CSV export, single-threaded
├── advanced_scraper.go # Advanced - JSON export, parallel, proxy support
├── crawler.go          # Link crawler - discovers URLs
├── go.mod              # Dependencies: Colly v2.1.0, Go 1.21
└── cache/              # Auto-created cache directory
```

## Core Components

### 1. Basic Scraper (main.go:23-119)
- **Entry**: `main()`
- **Target**: `https://scrapingcourse.com/ecommerce/`
- **Output**: `products.csv`
- **Selectors**:
  - Products: `li.product`
  - Name: `h2.woocommerce-loop-product__title`
  - Price: `span.price`
  - Pagination: `a.next.page-numbers`

### 2. Advanced Scraper (advanced_scraper.go:39-202)
- **Entry**: `runAdvancedExample()`
- **Class**: `Scraper` struct with mutex protection
- **Output**: `products_detailed.json`
- **Features**: Dual collectors, 4x parallel listings, 2x parallel details
- **Retry**: Auto-retry on 429/503 with 5s backoff

### 3. Web Crawler (crawler.go:24-108)
- **Entry**: `runCrawlerExample()`
- **Class**: `WebCrawler` struct
- **Default**: 10 pages max, depth 3
- **Target**: `go-colly.org`

## Key Data Structures

```go
// Basic product (main.go:15)
type Product struct {
    URL, Image, Name, Price string
    ScrapedAt time.Time
}

// Detailed product (advanced_scraper.go:17)
type ProductDetail struct {
    URL, Name, Price, Description string
    SKU, Category, ImageURL string
    InStock bool
    ScrapedAt time.Time
}

// Scraper state (advanced_scraper.go:30)
type Scraper struct {
    collector, detailCollector *colly.Collector
    products []ProductDetail
    mu sync.Mutex
    visited map[string]bool
}
```

## Critical Functions

```go
// CSV Export (main.go:137)
exportToCSV(products []Product, filename string) error

// JSON Export (advanced_scraper.go:212)
(s *Scraper) ExportToJSON(filename string) error

// Proxy Setup (advanced_scraper.go:76)
(s *Scraper) SetProxy(proxyURLs []string) error

// Start Scraping (advanced_scraper.go:183)
(s *Scraper) Scrape(startURL string) error
```

## Colly Configuration

### Rate Limiting
```go
// Basic (main.go:43)
c.Limit(&colly.LimitRule{
    Parallelism: 2,
    Delay: 1 * time.Second,
})

// Advanced (advanced_scraper.go:57)
s.collector.Limit(&colly.LimitRule{
    Parallelism: 4,
    Delay: 500 * time.Millisecond,
    RandomDelay: 500 * time.Millisecond,
})
```

### Headers (advanced_scraper.go:169-180)
```go
r.Headers.Set("User-Agent", "Mozilla/5.0...")
r.Headers.Set("Accept", "text/html...")
// + 8 more headers for anti-detection
```

## Common Modifications

### Change Target Site
```go
// main.go:100
c.Visit("https://your-site.com/products/")

// Update selectors in main.go:69-83
c.OnHTML("your.selector", func(e *colly.HTMLElement) {
    product.Name = e.ChildText("your.name.selector")
})
```

### Add Proxies
```go
// advanced_scraper.go:241-246 (uncomment)
proxies := []string{
    "http://proxy1:8080",
    "http://proxy2:8080",
}
scraper.SetProxy(proxies)
```

### Add Fields
```go
// Extend ProductDetail struct (advanced_scraper.go:17)
Rating float64 `json:"rating"`
Reviews int `json:"reviews"`

// Extract in callback (advanced_scraper.go:146)
Rating: e.ChildAttr("div.rating", "data-rating")
```

### Export Format
```go
// Add new export method to Scraper
func (s *Scraper) ExportToXML(filename string) error {
    // Implementation
}
```

## Error Handling

### HTTP Errors (advanced_scraper.go:106-115)
- **429 Too Many Requests**: 5s wait + retry
- **503 Service Unavailable**: 5s wait + retry
- **Other errors**: Log and continue

### Validation
- URL validation before visit (advanced_scraper.go:185)
- Empty data checks (main.go:79, advanced_scraper.go:159)
- Mutex protection for concurrent ops (advanced_scraper.go:34)

## Performance Tuning

### Speed Up
- Increase `Parallelism` (max 10 recommended)
- Decrease `Delay` (min 200ms recommended)
- Remove `RandomDelay`
- Disable cache: Remove `colly.CacheDir()`

### Reduce Detection
- Add more proxies
- Increase delays
- Add random user agents rotation
- Enable `RandomDelay`

## Testing & Debugging

```bash
# Check dependencies
go mod verify

# Run with verbose logging
go run main.go 2>&1 | tee scraper.log

# Test specific functions
go test -run TestExportToCSV

# Profile memory
go run -memprofile=mem.prof main.go

# Check for race conditions
go run -race main.go
```

## Common Issues

| Issue | Solution | Location |
|-------|----------|----------|
| 429 errors | Increase delays | main.go:43-47 |
| No data scraped | Check selectors | main.go:69-83 |
| Duplicate scraping | Check visited map | advanced_scraper.go:126-128 |
| Memory leak | Limit pages/depth | crawler.go:19,33 |
| Blocked by site | Add proxies | advanced_scraper.go:241 |

## Quick Commands

```bash
# Clean cache
rm -rf cache/

# Check output
head -20 products.csv
jq '.[0]' products_detailed.json

# Monitor progress
tail -f scraper.log

# Count products
wc -l products.csv
jq '. | length' products_detailed.json
```

## API Reference

### Colly Callbacks
- `OnRequest`: Before request (headers)
- `OnHTML`: Element found (extraction)
- `OnResponse`: After response (status)
- `OnError`: Error occurred (retry logic)
- `OnScraped`: Page complete (logging)

### Key Methods
- `c.Visit(url)`: Start scraping
- `c.Wait()`: Wait for async completion
- `e.ChildText(selector)`: Extract text
- `e.ChildAttr(selector, attr)`: Extract attribute
- `e.Request.Visit(url)`: Follow link

## Integration Points

```go
// Use as library
import "your-module/scraper"

s := scraper.NewScraper([]string{"domain.com"})
s.Scrape("https://domain.com/products")
products := s.GetProducts()

// Add to existing project
go get github.com/gocolly/colly/v2@v2.1.0
```

## Environment Variables

```bash
# Optional configurations
export HTTP_PROXY="http://proxy:8080"
export HTTPS_PROXY="http://proxy:8080"
export CACHE_DIR="./my-cache"
export MAX_PAGES="100"
```

---
**Note**: This guide assumes familiarity with Go and web scraping concepts. For detailed explanations, see OVERVIEW.md.