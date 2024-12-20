package format

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestFormatHumanDuration(t *testing.T) {
	tests := []struct {
		name     string
		days     int
		expected string
	}{
		{
			name:     "zero_days",
			days:     0,
			expected: "no change",
		},
		{
			name:     "single_day",
			days:     1,
			expected: "1 day",
		},
		{
			name:     "multiple_days",
			days:     5,
			expected: "5 days",
		},
		{
			name:     "one_week",
			days:     7,
			expected: "1 week",
		},
		{
			name:     "one_week_and_days",
			days:     9,
			expected: "1 week 2 days",
		},
		{
			name:     "multiple_weeks",
			days:     14,
			expected: "2 weeks",
		},
		{
			name:     "multiple_weeks_and_days",
			days:     17,
			expected: "2 weeks 3 days",
		},
		{
			name:     "one_month",
			days:     30,
			expected: "1 month",
		},
		{
			name:     "one_month_and_weeks",
			days:     37,
			expected: "1 month 1 week",
		},
		{
			name:     "multiple_months",
			days:     90,
			expected: "3 months",
		},
		{
			name:     "multiple_months_and_weeks",
			days:     95,
			expected: "3 months",
		},
		{
			name:     "one_year",
			days:     365,
			expected: "1 year",
		},
		{
			name:     "one_year_and_months",
			days:     395,
			expected: "1 year 1 month",
		},
		{
			name:     "multiple_years",
			days:     730,
			expected: "2 years",
		},
		{
			name:     "multiple_years_and_months",
			days:     760,
			expected: "2 years 1 month",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := formatHumanDuration(tt.days)
			assert.Equal(t, tt.expected, got)
		})
	}
}

func TestParseHumanRange(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		wantError bool
		checkFunc func(t *testing.T, from, to time.Time, err error)
	}{
		{
			name:      "valid range",
			input:     "2024-01-01 → 2024-01-31",
			wantError: false,
			checkFunc: func(t *testing.T, from, to time.Time, err error) {
				assert.NoError(t, err)
				assert.Equal(t, "2024-01-01", from.Format("2006-01-02"))
				assert.Equal(t, "2024-01-31", to.Format("2006-01-02"))
			},
		},
		{
			name:      "invalid format",
			input:     "2024-01-01 to 2024-01-31",
			wantError: true,
		},
		{
			name:      "invalid from date",
			input:     "invalid → 2024-01-31",
			wantError: true,
		},
		{
			name:      "invalid to date",
			input:     "2024-01-01 → invalid",
			wantError: true,
		},
		{
			name:      "reversed dates",
			input:     "2024-01-31 → 2024-01-01",
			wantError: true,
		},
		{
			name:      "last 12 hours",
			input:     "last 12 hours",
			wantError: false,
			checkFunc: func(t *testing.T, from, to time.Time, err error) {
				assert.NoError(t, err)
				duration := to.Sub(from)
				assert.Equal(t, 12*time.Hour, duration)
				assert.True(t, to.After(time.Now().Add(-time.Second))) // to should be very close to now
			},
		},
		{
			name:      "last 2 days",
			input:     "last 2 days",
			wantError: false,
			checkFunc: func(t *testing.T, from, to time.Time, err error) {
				assert.NoError(t, err)
				duration := to.Sub(from)
				assert.Equal(t, 48*time.Hour, duration)
				assert.True(t, to.After(time.Now().Add(-time.Second)))
			},
		},
		{
			name:      "last 1 week",
			input:     "last 1 week",
			wantError: false,
			checkFunc: func(t *testing.T, from, to time.Time, err error) {
				assert.NoError(t, err)
				duration := to.Sub(from)
				assert.Equal(t, 7*24*time.Hour, duration)
				assert.True(t, to.After(time.Now().Add(-time.Second)))
			},
		},
		{
			name:      "invalid relative format",
			input:     "last",
			wantError: true,
		},
		{
			name:      "invalid time unit",
			input:     "last 2 decades",
			wantError: true,
		},
		{
			name:      "invalid number",
			input:     "last abc hours",
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			from, to, err := ParseHumanRange(tt.input)
			if tt.wantError {
				assert.Error(t, err)
				return
			}
			if tt.checkFunc != nil {
				tt.checkFunc(t, from, to, err)
			}
		})
	}
}

func TestParseRelativeDuration(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		want      time.Duration
		wantError bool
	}{
		{
			name:  "minutes",
			input: "30 minutes",
			want:  30 * time.Minute,
		},
		{
			name:  "minute singular",
			input: "1 minute",
			want:  time.Minute,
		},
		{
			name:  "hours",
			input: "12 hours",
			want:  12 * time.Hour,
		},
		{
			name:  "hour singular",
			input: "1 hour",
			want:  time.Hour,
		},
		{
			name:  "days",
			input: "3 days",
			want:  72 * time.Hour,
		},
		{
			name:  "day singular",
			input: "1 day",
			want:  24 * time.Hour,
		},
		{
			name:  "weeks",
			input: "2 weeks",
			want:  14 * 24 * time.Hour,
		},
		{
			name:  "week singular",
			input: "1 week",
			want:  7 * 24 * time.Hour,
		},
		{
			name:      "invalid format",
			input:     "invalid",
			wantError: true,
		},
		{
			name:      "invalid number",
			input:     "abc hours",
			wantError: true,
		},
		{
			name:      "invalid unit",
			input:     "2 centuries",
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseRelativeDuration(tt.input)
			if tt.wantError {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestCalculateDelayLevel(t *testing.T) {
	tests := []struct {
		name          string
		durationDelta int
		moderateDelay int
		highDelay     int
		extremeDelay  int
		expectedLevel DelayLevel
	}{
		{
			name:          "on track",
			durationDelta: 0,
			moderateDelay: 7,
			highDelay:     14,
			extremeDelay:  30,
			expectedLevel: DelayLevelOnTrack,
		},
		{
			name:          "ahead of schedule",
			durationDelta: -5,
			moderateDelay: 7,
			highDelay:     14,
			extremeDelay:  30,
			expectedLevel: DelayLevelAhead,
		},
		{
			name:          "moderate delay",
			durationDelta: 10,
			moderateDelay: 7,
			highDelay:     14,
			extremeDelay:  30,
			expectedLevel: DelayLevelModerate,
		},
		{
			name:          "high delay",
			durationDelta: 20,
			moderateDelay: 7,
			highDelay:     14,
			extremeDelay:  30,
			expectedLevel: DelayLevelHigh,
		},
		{
			name:          "extreme delay",
			durationDelta: 35,
			moderateDelay: 7,
			highDelay:     14,
			extremeDelay:  30,
			expectedLevel: DelayLevelExtreme,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			level := calculateDelayLevel(tt.durationDelta, tt.moderateDelay, tt.highDelay, tt.extremeDelay)
			assert.Equal(t, tt.expectedLevel, level)
		})
	}
}

func TestCalculateTimelineDelayLevel(t *testing.T) {
	tests := []struct {
		name          string
		startDelta    int
		durationDelta int
		moderateDelay int
		highDelay     int
		extremeDelay  int
		expectedLevel DelayLevel
	}{
		{
			name:          "on track - no changes",
			startDelta:    0,
			durationDelta: 0,
			moderateDelay: 7,
			highDelay:     14,
			extremeDelay:  30,
			expectedLevel: DelayLevelOnTrack,
		},
		{
			name:          "ahead of schedule - earlier start",
			startDelta:    -5,
			durationDelta: 0,
			moderateDelay: 7,
			highDelay:     14,
			extremeDelay:  30,
			expectedLevel: DelayLevelAhead,
		},
		{
			name:          "ahead of schedule - shorter duration",
			startDelta:    0,
			durationDelta: -5,
			moderateDelay: 7,
			highDelay:     14,
			extremeDelay:  30,
			expectedLevel: DelayLevelOnTrack,
		},
		{
			name:          "ahead of schedule - earlier start and shorter duration",
			startDelta:    -5,
			durationDelta: -3,
			moderateDelay: 7,
			highDelay:     14,
			extremeDelay:  30,
			expectedLevel: DelayLevelAhead,
		},
		{
			name:          "moderate delay - delayed start",
			startDelta:    10,
			durationDelta: 0,
			moderateDelay: 7,
			highDelay:     14,
			extremeDelay:  30,
			expectedLevel: DelayLevelModerate,
		},
		{
			name:          "moderate delay - increased duration",
			startDelta:    0,
			durationDelta: 10,
			moderateDelay: 7,
			highDelay:     14,
			extremeDelay:  30,
			expectedLevel: DelayLevelModerate,
		},
		{
			name:          "high delay - delayed start",
			startDelta:    20,
			durationDelta: 0,
			moderateDelay: 7,
			highDelay:     14,
			extremeDelay:  30,
			expectedLevel: DelayLevelHigh,
		},
		{
			name:          "high delay - increased duration",
			startDelta:    0,
			durationDelta: 20,
			moderateDelay: 7,
			highDelay:     14,
			extremeDelay:  30,
			expectedLevel: DelayLevelHigh,
		},
		{
			name:          "extreme delay - delayed start",
			startDelta:    35,
			durationDelta: 0,
			moderateDelay: 7,
			highDelay:     14,
			extremeDelay:  30,
			expectedLevel: DelayLevelExtreme,
		},
		{
			name:          "extreme delay - increased duration",
			startDelta:    0,
			durationDelta: 35,
			moderateDelay: 7,
			highDelay:     14,
			extremeDelay:  30,
			expectedLevel: DelayLevelExtreme,
		},
		{
			name:          "use max of start delay and duration increase",
			startDelta:    10,
			durationDelta: 20,
			moderateDelay: 7,
			highDelay:     14,
			extremeDelay:  30,
			expectedLevel: DelayLevelHigh,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			level := calculateTimelineDelayLevel(tt.startDelta, tt.durationDelta, tt.moderateDelay, tt.highDelay, tt.extremeDelay)
			assert.Equal(t, tt.expectedLevel, level)
		})
	}
}
