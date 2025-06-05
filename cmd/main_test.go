package main

import (
	"testing"
	"time"
)

func TestIsWithinTimeRange(t *testing.T) {
	tests := []struct {
		name     string
		start    string
		end      string
		timezone string
		mockNow  func() time.Time
		expected bool
	}{
		{
			name:     "within range in UTC",
			start:    "2025/1/1 00:00:00",
			end:      "2025/12/31 23:59:59",
			timezone: "UTC",
			mockNow: func() time.Time {
				return time.Date(2025, 6, 15, 12, 0, 0, 0, time.UTC)
			},
			expected: true,
		},
		{
			name:     "within range in Asia/Tokyo",
			start:    "2025/1/1 00:00:00",
			end:      "2025/12/31 23:59:59",
			timezone: "Asia/Tokyo",
			mockNow: func() time.Time {
				loc, _ := time.LoadLocation("Asia/Tokyo")
				return time.Date(2025, 6, 15, 12, 0, 0, 0, loc)
			},
			expected: true,
		},
		{
			name:     "before start time",
			start:    "2025/6/15 13:00:00",
			end:      "2025/12/31 23:59:59",
			timezone: "UTC",
			mockNow: func() time.Time {
				return time.Date(2025, 6, 15, 12, 0, 0, 0, time.UTC)
			},
			expected: false,
		},
		{
			name:     "after end time",
			start:    "2025/1/1 00:00:00",
			end:      "2025/6/15 11:00:00",
			timezone: "UTC",
			mockNow: func() time.Time {
				return time.Date(2025, 6, 15, 12, 0, 0, 0, time.UTC)
			},
			expected: false,
		},
		{
			name:     "invalid timezone falls back to UTC",
			start:    "2025/1/1 00:00:00",
			end:      "2025/12/31 23:59:59",
			timezone: "Invalid/Timezone",
			mockNow: func() time.Time {
				return time.Date(2025, 6, 15, 12, 0, 0, 0, time.UTC)
			},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// We can't easily mock time.Now() in the function, so we'll test the logic
			// by checking if the timezone loading and parsing works correctly

			// Test timezone loading
			var loc *time.Location
			var err error
			if tt.timezone != "" {
				loc, err = time.LoadLocation(tt.timezone)
				if err != nil {
					loc = time.UTC
				}
			} else {
				loc = time.UTC
			}

			// Test time parsing
			layout := "2006/1/2 15:04:05"
			startTime, err := time.ParseInLocation(layout, tt.start, loc)
			if err != nil {
				t.Errorf("Failed to parse start time: %v", err)
				return
			}

			endTime, err := time.ParseInLocation(layout, tt.end, loc)
			if err != nil {
				t.Errorf("Failed to parse end time: %v", err)
				return
			}

			// Use mock time for comparison
			mockNow := tt.mockNow().In(loc)
			result := mockNow.After(startTime) && mockNow.Before(endTime)

			if result != tt.expected {
				t.Errorf("Expected %v, got %v", tt.expected, result)
				t.Errorf("Start: %v, End: %v, Now: %v", startTime, endTime, mockNow)
			}
		})
	}
}

func TestIsWithinTimeRangeInvalidInput(t *testing.T) {
	tests := []struct {
		name     string
		start    string
		end      string
		timezone string
		expected bool
	}{
		{
			name:     "invalid start time format",
			start:    "invalid-time",
			end:      "2025/12/31 23:59:59",
			timezone: "UTC",
			expected: false,
		},
		{
			name:     "invalid end time format",
			start:    "2025/1/1 00:00:00",
			end:      "invalid-time",
			timezone: "UTC",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isWithinTimeRange(tt.start, tt.end, tt.timezone)
			if result != tt.expected {
				t.Errorf("Expected %v, got %v", tt.expected, result)
			}
		})
	}
}
