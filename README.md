# Go Web Scraper

A Golang web scraper built with [Colly](https://github.com/gocolly/colly) - the popular Go scraping framework.

Based on the [ZenRows Web Scraping in Golang Tutorial](https://www.zenrows.com/blog/web-scraping-golang).

## Features

- Product scraping from e-commerce sites
- CSV and JSON export
- Rate limiting to avoid being blocked
- Parallel scraping for better performance
- Proxy rotation support
- Web crawling capabilities
- Request retries on errors
- Custom User-Agent headers

## Project Structure

```
web-scraper-go/
├── go.mod                  # Go module definition
├── main.go                 # Basic scraper with CSV export
├── advanced_scraper.go     # Advanced scraper with parallel requests & JSON
├── crawler.go              # Web crawler example
└── README.md               # This file
```

## Installation

1. Make sure you have Go 1.19+ installed
2. Clone or navigate to this directory
3. Install dependencies:

```bash
go mod tidy
```

## Usage

### Basic Product Scraper

Run the main scraper to collect product data and export to CSV:

```bash
go run main.go
```

This will:
- Scrape products from `https://scrapingcourse.com/ecommerce/`
- Follow pagination automatically
- Export results to `products.csv`

### Advanced Scraper

To use the advanced scraper with parallel requests and JSON export:

```go
// In your code, call:
runAdvancedExample()
```

Or create a new file:

```go
package main

func main() {
    runAdvancedExample()
}
```

### Web Crawler

To crawl a website and discover links:

```go
package main

func main() {
    runCrawlerExample()
}
```

## Configuration Options

### Rate Limiting

```go
c.Limit(&colly.LimitRule{
    DomainGlob:  "*",
    Parallelism: 2,              // Concurrent requests
    Delay:       1 * time.Second, // Delay between requests
    RandomDelay: 500 * time.Millisecond,
})
```

### Proxy Rotation

```go
proxies := []string{
    "http://proxy1:8080",
    "http://proxy2:8080",
}
scraper.SetProxy(proxies)
```

### Custom Headers

Headers are automatically set to mimic a real browser. You can customize them in the `setHeaders` function.

## Key Colly Callbacks

| Callback | Description |
|----------|-------------|
| `OnRequest` | Called before making a request |
| `OnResponse` | Called after receiving a response |
| `OnHTML` | Called when an HTML element matches a CSS selector |
| `OnXML` | Called when an XML element matches an XPath |
| `OnError` | Called when an error occurs |
| `OnScraped` | Called after scraping a page |

## Example Output

### CSV Output (products.csv)

```csv
Name,Price,URL,Image,Scraped At
"Product Name","$19.99","https://...","https://...","2024-01-15T10:30:00Z"
```

### JSON Output (products_detailed.json)

```json
[
  {
    "url": "https://...",
    "name": "Product Name",
    "price": "$19.99",
    "description": "...",
    "sku": "SKU123",
    "category": "Category",
    "image_url": "https://...",
    "in_stock": true,
    "scraped_at": "2024-01-15T10:30:00Z"
  }
]
```

## Best Practices

1. **Respect robots.txt**: Check the site's robots.txt before scraping
2. **Rate limiting**: Always add delays between requests
3. **Error handling**: Implement retries for transient errors
4. **Caching**: Use `colly.CacheDir()` during development
5. **User-Agent**: Set realistic browser headers
6. **Be ethical**: Don't overload target servers

## Dependencies

- [gocolly/colly](https://github.com/gocolly/colly) - Web scraping framework

## Resources

- [Colly Documentation](https://go-colly.org/docs/)
- [Colly Examples](https://go-colly.org/docs/examples/basic/)
- [GoQuery](https://github.com/PuerkitoBio/goquery) - HTML parsing (used by Colly)

## License

MIT
