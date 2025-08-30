package query

import (
	"testing"
	"time"
)

func TestEndOfYear(t *testing.T) {
	endOf2024 := time.Date(2024, time.December, 31, 23, 59, 59, int(time.Second-time.Nanosecond), time.UTC)
	tests := []struct {
		name     string
		input    time.Time
		expected time.Time
	}{
		{
			name:     "from August",
			input:    time.Date(2024, time.August, 15, 12, 30, 45, 0, time.UTC),
			expected: endOf2024,
		},
		{
			name:     "from January",
			input:    time.Date(2024, time.January, 1, 0, 0, 0, 0, time.UTC),
			expected: endOf2024,
		},
		{
			name:     "from December",
			input:    time.Date(2024, time.December, 15, 18, 20, 30, 0, time.UTC),
			expected: endOf2024,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := endOfYear(tt.input)
			if !result.Equal(tt.expected) {
				t.Errorf("endOfYear(%v) = %v, expected %v",
					tt.input.Format("2006-01-02 15:04:05"),
					result.Format("2006-01-02 15:04:05"),
					tt.expected.Format("2006-01-02 15:04:05"))
			}
		})
	}
}
