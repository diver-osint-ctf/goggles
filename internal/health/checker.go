package health

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
	"unicode/utf8"

	"goggles/internal/website"
)

// Result represents the result of a health check
type Result struct {
	Website     website.Website
	OK          bool
	Status      int
	HasWord     bool
	Date        time.Time
	Error       string
	HTMLContent string // Store HTML content for logging
}

// Checker performs health checks on websites
type Checker struct {
	client   *http.Client
	verbose  bool // Enable HTML content logging
	testMode bool // Disable delays for testing
}

// NewChecker creates a new health checker
func NewChecker(verbose bool) *Checker {
	return &Checker{
		client: &http.Client{
			Timeout: 30 * time.Second,
			Transport: &http.Transport{
				MaxIdleConns:        100,
				MaxIdleConnsPerHost: 10,
				IdleConnTimeout:     90 * time.Second,
				DisableCompression:  false, // Enable compression like browsers
				ForceAttemptHTTP2:   true,  // Prefer HTTP/2 like modern browsers
			},
		},
		verbose:  verbose,
		testMode: false,
	}
}

// NewCheckerForTest creates a new health checker optimized for testing
func NewCheckerForTest(verbose bool) *Checker {
	checker := NewChecker(verbose)
	checker.testMode = true
	return checker
}

// Check performs a health check on a website
func (c *Checker) Check(site website.Website) Result {
	result := Result{
		Website: site,
		Date:    time.Now(),
	}

	// Add a small random delay to mimic human behavior (1-3 seconds), skip in test mode
	if !c.testMode {
		time.Sleep(time.Duration(1000+time.Now().UnixNano()%2000) * time.Millisecond)
	}

	// Create HTTP request with proper headers
	req, err := http.NewRequest("GET", site.URL, nil)
	if err != nil {
		result.Error = fmt.Sprintf("failed to create request: %v", err)
		return result
	}

	// Set headers to mimic a real browser
	req.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/135.0.0.0 Safari/537.36")
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.7")
	// Note: Accept-Encoding removed to avoid compression issues that cause garbled content
	req.Header.Set("Accept-Language", "ja,en-US;q=0.9,en;q=0.8,id;q=0.7")
	req.Header.Set("Cache-Control", "max-age=0")
	req.Header.Set("Priority", "u=0, i")
	req.Header.Set("Sec-Ch-Ua", `"Google Chrome";v="135", "Not-A.Brand";v="8", "Chromium";v="135"`)
	req.Header.Set("Sec-Ch-Ua-Mobile", "?0")
	req.Header.Set("Sec-Ch-Ua-Platform", `"macOS"`)
	req.Header.Set("Sec-Fetch-Dest", "document")
	req.Header.Set("Sec-Fetch-Mode", "navigate")
	req.Header.Set("Sec-Fetch-Site", "none")
	req.Header.Set("Sec-Fetch-User", "?1")
	req.Header.Set("Upgrade-Insecure-Requests", "1")

	// Perform the request
	resp, err := c.client.Do(req)
	if err != nil {
		result.Error = fmt.Sprintf("request failed: %v", err)
		return result
	}
	defer resp.Body.Close()

	result.Status = resp.StatusCode

	// Check if status code is successful (2xx)
	validStatusCode := resp.StatusCode >= 200 && resp.StatusCode < 300

	// Read response body to check for words and store content
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		result.Error = fmt.Sprintf("failed to read response body: %v", err)
		return result
	}
	content := string(body)

	// Store HTML content if verbose mode is enabled
	if c.verbose {
		result.HTMLContent = content
	}

	// Check if any of the check words are present
	result.HasWord = false
	if len(site.CheckWords) > 0 {
		for _, word := range site.CheckWords {
			if strings.Contains(content, word) {
				result.HasWord = true
				break
			}
		}
	} else {
		result.HasWord = true // If no check words specified, consider it as passed
	}

	result.OK = validStatusCode && result.HasWord

	// Save HTML content if verbose mode is enabled and health check failed
	if c.verbose && !result.OK {
		// Add debug information about why the check failed
		var reasons []string
		if !validStatusCode {
			reasons = append(reasons, fmt.Sprintf("Invalid status code: %d", result.Status))
		}
		if !result.HasWord {
			if len(site.CheckWords) > 0 {
				reasons = append(reasons, fmt.Sprintf("Check words not found: %v", site.CheckWords))
			}
		}

		c.saveHTMLLog(site.URL, content, result.Date, reasons, resp.Header.Get("Content-Type"), resp.Header.Get("Content-Encoding"), resp.Header.Get("Content-Length"))
	}

	return result
}

// saveHTMLLog saves HTML content to logs directory
func (c *Checker) saveHTMLLog(siteURL, content string, timestamp time.Time, reasons []string, contentType, contentEncoding, contentLength string) {
	// Parse URL to get a safe filename
	parsedURL, err := url.Parse(siteURL)
	if err != nil {
		fmt.Printf("Failed to parse URL for logging: %v\n", err)
		return
	}

	// Create safe filename: timestamp_hostname_path.html
	hostname := parsedURL.Hostname()
	path := strings.ReplaceAll(parsedURL.Path, "/", "_")
	if path == "" || path == "_" {
		path = "index"
	}

	filename := fmt.Sprintf("%s_%s%s.html",
		timestamp.Format("2006-01-02_15-04-05"),
		hostname,
		path)

	// Remove any remaining unsafe characters
	filename = strings.ReplaceAll(filename, ":", "_")
	filename = strings.ReplaceAll(filename, "?", "_")
	filename = strings.ReplaceAll(filename, "&", "_")

	logPath := filepath.Join("logs", filename)

	// Prepare content with debug information
	debugInfo := fmt.Sprintf("<!-- DEBUG INFO -->\n<!-- URL: %s -->\n<!-- Timestamp: %s -->\n<!-- Content-Type: %s -->\n<!-- Content-Encoding: %s -->\n<!-- Content-Length: %s -->\n<!-- Failure reasons: %s -->\n<!-- Content is valid UTF-8: %t -->\n\n",
		siteURL,
		timestamp.Format("2006-01-02 15:04:05"),
		contentType,
		contentEncoding,
		contentLength,
		strings.Join(reasons, ", "),
		utf8.ValidString(content))

	// Handle encoding issues
	var finalContent string
	if utf8.ValidString(content) {
		finalContent = debugInfo + content
	} else {
		// If content is not valid UTF-8, try to clean it or save as binary info
		finalContent = debugInfo + fmt.Sprintf("<!-- WARNING: Content contains invalid UTF-8 characters -->\n<!-- Content length: %d bytes -->\n<!-- First 500 chars (cleaned): %s -->\n",
			len(content),
			strings.ToValidUTF8(content[:min(500, len(content))], ""))
	}

	err = os.WriteFile(logPath, []byte(finalContent), 0644)
	if err != nil {
		fmt.Printf("Failed to save HTML log to %s: %v\n", logPath, err)
		return
	}

	fmt.Printf("HTML content saved to: %s (reasons: %s)\n", logPath, strings.Join(reasons, ", "))
}

// min returns the minimum of two integers
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// FetchHTML fetches HTML content for a single URL and returns it as string
func (c *Checker) FetchHTML(url string) (string, error) {
	// Create HTTP request with proper headers
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return "", fmt.Errorf("failed to create request: %v", err)
	}

	// Set headers to mimic a real browser (same as Check method)
	req.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/135.0.0.0 Safari/537.36")
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.7")
	// Note: Accept-Encoding removed to avoid compression issues that cause garbled content
	req.Header.Set("Accept-Language", "ja,en-US;q=0.9,en;q=0.8,id;q=0.7")
	req.Header.Set("Cache-Control", "max-age=0")
	req.Header.Set("Priority", "u=0, i")
	req.Header.Set("Sec-Ch-Ua", `"Google Chrome";v="135", "Not-A.Brand";v="8", "Chromium";v="135"`)
	req.Header.Set("Sec-Ch-Ua-Mobile", "?0")
	req.Header.Set("Sec-Ch-Ua-Platform", `"macOS"`)
	req.Header.Set("Sec-Fetch-Dest", "document")
	req.Header.Set("Sec-Fetch-Mode", "navigate")
	req.Header.Set("Sec-Fetch-Site", "none")
	req.Header.Set("Sec-Fetch-User", "?1")
	req.Header.Set("Upgrade-Insecure-Requests", "1")

	// Perform the request
	resp, err := c.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("request failed: %v", err)
	}
	defer resp.Body.Close()

	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response body: %v", err)
	}

	return string(body), nil
}

// FetchHeaders fetches HTTP headers for a single URL and returns them as formatted string
func (c *Checker) FetchHeaders(url string) (string, error) {
	// Create HTTP request with proper headers
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return "", fmt.Errorf("failed to create request: %v", err)
	}

	// Set headers to mimic a real browser (same as Check method)
	req.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/135.0.0.0 Safari/537.36")
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.7")
	// Note: Accept-Encoding removed to avoid compression issues that cause garbled content
	req.Header.Set("Accept-Language", "ja,en-US;q=0.9,en;q=0.8,id;q=0.7")
	req.Header.Set("Cache-Control", "max-age=0")
	req.Header.Set("Priority", "u=0, i")
	req.Header.Set("Sec-Ch-Ua", `"Google Chrome";v="135", "Not-A.Brand";v="8", "Chromium";v="135"`)
	req.Header.Set("Sec-Ch-Ua-Mobile", "?0")
	req.Header.Set("Sec-Ch-Ua-Platform", `"macOS"`)
	req.Header.Set("Sec-Fetch-Dest", "document")
	req.Header.Set("Sec-Fetch-Mode", "navigate")
	req.Header.Set("Sec-Fetch-Site", "none")
	req.Header.Set("Sec-Fetch-User", "?1")
	req.Header.Set("Upgrade-Insecure-Requests", "1")
	req.Header.Set("Requests", "")
	req.Header.Set("Accept-Encoding", "gzip, deflate, br, zstd")

	// Perform the request
	resp, err := c.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("request failed: %v", err)
	}
	defer resp.Body.Close()

	// Format response headers
	var headerStr strings.Builder
	headerStr.WriteString(fmt.Sprintf("HTTP/%d.%d %s\n", resp.ProtoMajor, resp.ProtoMinor, resp.Status))

	// Sort header keys for consistent output
	var keys []string
	for key := range resp.Header {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	// Write headers in sorted order
	for _, key := range keys {
		values := resp.Header[key]
		for _, value := range values {
			headerStr.WriteString(fmt.Sprintf("%s: %s\n", key, value))
		}
	}

	return headerStr.String(), nil
}
