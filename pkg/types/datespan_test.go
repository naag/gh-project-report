package types

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewDateSpan(t *testing.T) {
	tests := []struct {
		name      string
		start     string
		end       string
		wantError bool
	}{
		{
			name:      "valid_range",
			start:     "2024-01-01",
			end:       "2024-01-31",
			wantError: false,
		},
		{
			name:      "same_day",
			start:     "2024-01-01",
			end:       "2024-01-01",
			wantError: false,
		},
		{
			name:      "invalid_start",
			start:     "2024-13-01",
			end:       "2024-01-31",
			wantError: true,
		},
		{
			name:      "invalid_end",
			start:     "2024-01-01",
			end:       "2024-01-32",
			wantError: true,
		},
		{
			name:      "end_before_start",
			start:     "2024-01-31",
			end:       "2024-01-01",
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ds, err := NewDateSpan(tt.start, tt.end)
			if tt.wantError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.start, ds.Start.Format("2006-01-02"))
				assert.Equal(t, tt.end, ds.End.Format("2006-01-02"))
			}
		})
	}
}

func TestDateSpan_DurationDays(t *testing.T) {
	tests := []struct {
		name     string
		start    string
		end      string
		expected int
	}{
		{
			name:     "same_day",
			start:    "2024-01-01",
			end:      "2024-01-01",
			expected: 1,
		},
		{
			name:     "adjacent_days",
			start:    "2024-01-01",
			end:      "2024-01-02",
			expected: 2,
		},
		{
			name:     "one_week",
			start:    "2024-01-01",
			end:      "2024-01-07",
			expected: 7,
		},
		{
			name:     "across_month",
			start:    "2024-01-31",
			end:      "2024-02-02",
			expected: 3,
		},
		{
			name:     "across_year",
			start:    "2023-12-31",
			end:      "2024-01-01",
			expected: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ds := MustNewDateSpan(tt.start, tt.end)
			assert.Equal(t, tt.expected, ds.DurationDays())
		})
	}
}

func TestDateSpan_CompareTo(t *testing.T) {
	tests := []struct {
		name     string
		base     DateSpan
		other    DateSpan
		expected DateSpanChange
	}{
		{
			name:  "no_change",
			base:  MustNewDateSpan("2024-01-01", "2024-01-31"),
			other: MustNewDateSpan("2024-01-01", "2024-01-31"),
			expected: DateSpanChange{
				StartDaysDelta: 0,
				EndDaysDelta:   0,
				DurationDelta:  0,
			},
		},
		{
			name:  "delayed_same_duration",
			base:  MustNewDateSpan("2024-01-01", "2024-01-31"),
			other: MustNewDateSpan("2024-01-15", "2024-02-14"),
			expected: DateSpanChange{
				StartDaysDelta: 14,
				EndDaysDelta:   14,
				DurationDelta:  0,
			},
		},
		{
			name:  "earlier_extended",
			base:  MustNewDateSpan("2024-01-15", "2024-01-31"),
			other: MustNewDateSpan("2024-01-01", "2024-02-14"),
			expected: DateSpanChange{
				StartDaysDelta: -14,
				EndDaysDelta:   14,
				DurationDelta:  28,
			},
		},
		{
			name:  "same_start_extended",
			base:  MustNewDateSpan("2024-01-01", "2024-01-31"),
			other: MustNewDateSpan("2024-01-01", "2024-02-14"),
			expected: DateSpanChange{
				StartDaysDelta: 0,
				EndDaysDelta:   14,
				DurationDelta:  14,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			change := tt.base.CompareTo(tt.other)
			assert.Equal(t, tt.expected, change)
		})
	}
}
