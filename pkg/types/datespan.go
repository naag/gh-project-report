package types

import (
	"fmt"
	"time"
)

// DateSpan represents a span of time with a start and end date
type DateSpan struct {
	Start time.Time
	End   time.Time
}

// NewDateSpan creates a DateSpan from string dates in YYYY-MM-DD format
func NewDateSpan(start, end string) (DateSpan, error) {
	startTime, err := time.Parse("2006-01-02", start)
	if err != nil {
		return DateSpan{}, fmt.Errorf("invalid start date: %w", err)
	}
	endTime, err := time.Parse("2006-01-02", end)
	if err != nil {
		return DateSpan{}, fmt.Errorf("invalid end date: %w", err)
	}
	if endTime.Before(startTime) {
		return DateSpan{}, fmt.Errorf("end date %s is before start date %s", end, start)
	}
	return DateSpan{Start: startTime, End: endTime}, nil
}

// MustNewDateSpan creates a DateSpan and panics if the dates are invalid
func MustNewDateSpan(start, end string) DateSpan {
	tr, err := NewDateSpan(start, end)
	if err != nil {
		panic(err)
	}
	return tr
}

// DurationDays returns the duration in days, including both start and end days
func (ds DateSpan) DurationDays() int {
	return int(ds.End.Sub(ds.Start).Hours()/24) + 1
}

// DateSpanChange represents how a time range has changed
type DateSpanChange struct {
	StartDaysDelta int // positive = moved later, negative = moved earlier
	EndDaysDelta   int // positive = extended, negative = shortened
	DurationDelta  int // change in duration in days
}

// CompareTo compares this range to another and returns the changes
func (ds DateSpan) CompareTo(other DateSpan) DateSpanChange {
	startDelta := int(other.Start.Sub(ds.Start).Hours() / 24)
	endDelta := int(other.End.Sub(ds.End).Hours() / 24)
	return DateSpanChange{
		StartDaysDelta: startDelta,
		EndDaysDelta:   endDelta,
		DurationDelta:  other.DurationDays() - ds.DurationDays(),
	}
}

// Equal returns true if this DateSpan is equal to the other DateSpan
func (ds DateSpan) Equal(other DateSpan) bool {
	return ds.Start.Equal(other.Start) && ds.End.Equal(other.End)
}
