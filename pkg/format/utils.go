package format

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

// calculateDelayLevel determines the delay level based on duration delta and thresholds
func calculateDelayLevel(durationDelta, moderateDelay, highDelay, extremeDelay int) DelayLevel {
	if durationDelta < 0 {
		return DelayLevelAhead
	}
	if durationDelta == 0 {
		return DelayLevelOnTrack
	}
	if durationDelta >= extremeDelay {
		return DelayLevelExtreme
	}
	if durationDelta >= highDelay {
		return DelayLevelHigh
	}
	if durationDelta >= moderateDelay {
		return DelayLevelModerate
	}
	return DelayLevelOnTrack
}

// calculateTimelineDelayLevel determines the delay level based on both start delay and duration change
func calculateTimelineDelayLevel(startDaysDelta, durationDelta, moderateDelay, highDelay, extremeDelay int) DelayLevel {
	// If we're ahead of schedule (earlier start or shorter duration)
	if startDaysDelta < 0 && durationDelta <= 0 {
		return DelayLevelAhead
	}

	// Use the maximum of start delay and duration increase to determine delay
	maxDelay := startDaysDelta
	if durationDelta > startDaysDelta {
		maxDelay = durationDelta
	}

	if maxDelay == 0 {
		return DelayLevelOnTrack
	}
	if maxDelay >= extremeDelay {
		return DelayLevelExtreme
	}
	if maxDelay >= highDelay {
		return DelayLevelHigh
	}
	if maxDelay >= moderateDelay {
		return DelayLevelModerate
	}
	return DelayLevelOnTrack
}

// formatHumanDuration formats a duration in days into a human-readable string
func formatHumanDuration(days int) string {
	if days == 0 {
		return "no change"
	}

	years := days / 365
	remainingDays := days % 365
	months := remainingDays / 30
	remainingDays = remainingDays % 30
	weeks := remainingDays / 7
	remainingDays = remainingDays % 7

	if years > 0 {
		if months == 0 {
			return fmt.Sprintf("%d year%s", years, pluralize(years))
		}
		return fmt.Sprintf("%d year%s %d month%s", years, pluralize(years), months, pluralize(months))
	}

	if months > 0 {
		if weeks == 0 {
			return fmt.Sprintf("%d month%s", months, pluralize(months))
		}
		return fmt.Sprintf("%d month%s %d week%s", months, pluralize(months), weeks, pluralize(weeks))
	}

	if weeks > 0 {
		if remainingDays == 0 {
			return fmt.Sprintf("%d week%s", weeks, pluralize(weeks))
		}
		return fmt.Sprintf("%d week%s %d day%s", weeks, pluralize(weeks), remainingDays, pluralize(remainingDays))
	}

	return fmt.Sprintf("%d day%s", days, pluralize(days))
}

// pluralize returns "s" if n != 1, empty string otherwise
func pluralize(n int) string {
	if n == 1 {
		return ""
	}
	return "s"
}

// formatDate formats a time.Time using the specified format string
func formatDate(t time.Time, format string) string {
	return t.Format(format)
}

// ParseHumanRange parses a human-readable time range
func ParseHumanRange(timeRange string) (time.Time, time.Time, error) {
	// Handle relative time ranges
	if strings.HasPrefix(timeRange, "last ") {
		duration, err := parseRelativeDuration(strings.TrimPrefix(timeRange, "last "))
		if err != nil {
			return time.Time{}, time.Time{}, fmt.Errorf("invalid relative time range: %w", err)
		}
		now := time.Now()
		return now.Add(-duration), now, nil
	}

	// Handle explicit date ranges
	parts := strings.Split(timeRange, "→")
	if len(parts) != 2 {
		return time.Time{}, time.Time{}, fmt.Errorf("invalid time range format, expected 'from → to' or 'last X hours/days/weeks'")
	}

	fromStr := strings.TrimSpace(parts[0])
	toStr := strings.TrimSpace(parts[1])

	from, err := time.Parse("2006-01-02", fromStr)
	if err != nil {
		return time.Time{}, time.Time{}, fmt.Errorf("invalid from date: %w", err)
	}

	to, err := time.Parse("2006-01-02", toStr)
	if err != nil {
		return time.Time{}, time.Time{}, fmt.Errorf("invalid to date: %w", err)
	}

	if to.Before(from) {
		return time.Time{}, time.Time{}, fmt.Errorf("to date cannot be before from date")
	}

	return from, to, nil
}

// parseRelativeDuration parses strings like "12 hours", "2 days", "1 week"
func parseRelativeDuration(s string) (time.Duration, error) {
	parts := strings.Fields(s)
	if len(parts) != 2 {
		return 0, fmt.Errorf("invalid format, expected '<number> <unit>'")
	}

	amount, err := strconv.ParseFloat(parts[0], 64)
	if err != nil {
		return 0, fmt.Errorf("invalid number: %w", err)
	}

	unit := strings.ToLower(parts[1])
	// Handle plural forms
	unit = strings.TrimSuffix(unit, "s")

	switch unit {
	case "minute":
		return time.Duration(amount * float64(time.Minute)), nil
	case "hour":
		return time.Duration(amount * float64(time.Hour)), nil
	case "day":
		return time.Duration(amount * 24 * float64(time.Hour)), nil
	case "week":
		return time.Duration(amount * 7 * 24 * float64(time.Hour)), nil
	default:
		return 0, fmt.Errorf("unsupported time unit: %s", unit)
	}
}
