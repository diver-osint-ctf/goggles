package website

import (
	"encoding/csv"
	"fmt"
	"os"
	"strings"
)

// Website represents a website to check
type Website struct {
	URL        string
	CheckWords []string
}

// LoadFromCSV loads websites from a CSV file
func LoadFromCSV(filename string) ([]Website, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to open CSV file: %w", err)
	}
	defer file.Close()

	reader := csv.NewReader(file)
	records, err := reader.ReadAll()
	if err != nil {
		return nil, fmt.Errorf("failed to read CSV file: %w", err)
	}

	if len(records) == 0 {
		return nil, fmt.Errorf("CSV file is empty")
	}

	// Skip header row
	var websites []Website
	for i, record := range records[1:] {
		if len(record) < 2 {
			return nil, fmt.Errorf("invalid CSV format at line %d: expected at least 2 columns", i+2)
		}

		url := strings.TrimSpace(record[0])
		checkWordsStr := strings.TrimSpace(record[1])

		if url == "" {
			continue // Skip empty URLs
		}

		// Split check words by comma or newline
		var checkWords []string
		for _, word := range strings.Split(checkWordsStr, ",") {
			word = strings.TrimSpace(word)
			if word != "" {
				checkWords = append(checkWords, word)
			}
		}

		websites = append(websites, Website{
			URL:        url,
			CheckWords: checkWords,
		})
	}

	return websites, nil
}
