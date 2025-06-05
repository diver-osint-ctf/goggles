package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"goggles/internal/config"
	"goggles/internal/discord"
	"goggles/internal/health"
	"goggles/internal/website"
)

func main() {
	// Parse command line flags
	verbose := flag.Bool("v", false, "verbose mode: save HTML content when health checks fail")
	configFile := flag.String("config", "config.yaml", "path to config file")
	urlsFile := flag.String("urls", "urls.csv", "path to URLs CSV file")
	singleURL := flag.String("s", "", "single URL mode: fetch and print HTML content for the specified URL")
	headerOnly := flag.Bool("header", false, "header mode: when used with -s, fetch and print only HTTP headers")
	flag.Parse()

	// Single URL mode
	if *singleURL != "" {
		checker := health.NewChecker(false) // No verbose mode needed for single URL

		if *headerOnly {
			// Fetch headers only
			headers, err := checker.FetchHeaders(*singleURL)
			if err != nil {
				log.Fatalf("Failed to fetch headers from %s: %v", *singleURL, err)
			}
			fmt.Print(headers)
		} else {
			// Fetch HTML content
			html, err := checker.FetchHTML(*singleURL)
			if err != nil {
				log.Fatalf("Failed to fetch HTML from %s: %v", *singleURL, err)
			}
			fmt.Print(html)
		}
		return
	}

	// Load configuration
	cfg, err := config.Load(*configFile)
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Load websites
	websites, err := website.LoadFromCSV(*urlsFile)
	if err != nil {
		log.Fatalf("Failed to load websites: %v", err)
	}

	if len(websites) == 0 {
		log.Println("No websites to check")
		return
	}

	// Check if current time is within the configured time range
	if !isWithinTimeRange(cfg.Start, cfg.End, cfg.Timezone) {
		log.Println("Current time is outside the configured range")
		return
	}

	// Create logs directory if verbose mode is enabled
	if *verbose {
		if err := os.MkdirAll("logs", 0755); err != nil {
			log.Printf("Failed to create logs directory: %v", err)
		}
	}

	// Perform health checks
	checker := health.NewChecker(*verbose)
	results := make([]health.Result, len(websites))

	// Set default threads if not specified or invalid
	threads := cfg.Threads
	if threads <= 0 {
		threads = 1
	}
	log.Printf("Using %d threads for parallel processing", threads)

	// Create channels for worker pool
	jobs := make(chan int, len(websites))
	resultsChan := make(chan struct {
		index  int
		result health.Result
	}, len(websites))

	// Start worker goroutines
	for w := 0; w < threads; w++ {
		go func() {
			for i := range jobs {
				site := websites[i]
				log.Printf("Checking %s...", site.URL)
				result := checker.Check(site)

				// Log the result for debugging
				if result.OK {
					log.Printf("✓ %s - OK (status: %d, has words: %t)", site.URL, result.Status, result.HasWord)
				} else {
					reasons := []string{}
					if result.Status < 200 || result.Status >= 300 {
						reasons = append(reasons, fmt.Sprintf("status: %d", result.Status))
					}
					if !result.HasWord && len(site.CheckWords) > 0 {
						reasons = append(reasons, fmt.Sprintf("missing words: %v", site.CheckWords))
					}
					if result.Error != "" {
						reasons = append(reasons, fmt.Sprintf("error: %s", result.Error))
					}
					log.Printf("✗ %s - FAILED (%s)", site.URL, strings.Join(reasons, ", "))
				}

				resultsChan <- struct {
					index  int
					result health.Result
				}{i, result}
			}
		}()
	}

	// Send jobs to workers
	for i := range websites {
		jobs <- i
	}
	close(jobs)

	// Collect results
	for i := 0; i < len(websites); i++ {
		res := <-resultsChan
		results[res.index] = res.result
	}

	// Send Discord notification
	discordClient := discord.NewClient(cfg.DiscordWebhookURL)

	// Send summary
	failures := getFailures(results)
	err = discordClient.SendSummary(failures, len(websites))
	if err != nil {
		log.Printf("Failed to send summary: %v", err)
	}

	// Send detailed report if there are failures
	if len(failures) > 0 {
		err = discordClient.SendReport(failures, cfg.Mentions)
		if err != nil {
			log.Printf("Failed to send report: %v", err)
		}
	}

	log.Println("Health check completed")
}

func isWithinTimeRange(start, end, timezone string) bool {
	layout := "2006/1/2 15:04:05"

	// Load the specified timezone
	loc, err := time.LoadLocation(timezone)
	if err != nil {
		log.Printf("Failed to load timezone %s: %v, using UTC", timezone, err)
		loc = time.UTC
	}

	// Parse start and end times in the specified timezone
	startTime, err := time.ParseInLocation(layout, start, loc)
	if err != nil {
		log.Printf("Failed to parse start time: %v", err)
		return false
	}

	endTime, err := time.ParseInLocation(layout, end, loc)
	if err != nil {
		log.Printf("Failed to parse end time: %v", err)
		return false
	}

	// Get current time in the specified timezone
	now := time.Now().In(loc)
	return now.After(startTime) && now.Before(endTime)
}

func getFailures(results []health.Result) []health.Result {
	var failures []health.Result
	for _, result := range results {
		if !result.OK {
			failures = append(failures, result)
		}
	}
	return failures
}
