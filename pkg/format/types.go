package format

import (
	"github.com/naag/gh-project-report/pkg/types"
)

// FormatterOptions contains configuration options for formatters
type FormatterOptions struct {
	DateFormat             string
	ModerateDelayThreshold int
	HighDelayThreshold     int
	ExtremeDelayThreshold  int
}

// Formatter interface defines methods that all formatters must implement
type Formatter interface {
	Format(diff types.ProjectDiff) string
}

// DelayLevel represents the delay level of a timeline change
type DelayLevel string

const (
	DelayLevelOnTrack  DelayLevel = "🔵 On track"
	DelayLevelAhead    DelayLevel = "🚀 Ahead of schedule"
	DelayLevelModerate DelayLevel = "🟠 Moderate delay"
	DelayLevelHigh     DelayLevel = "🔴 High delay"
	DelayLevelExtreme  DelayLevel = "🚫 Extreme delay"
)

// DefaultOptions returns the default formatter options
func DefaultOptions() FormatterOptions {
	return FormatterOptions{
		DateFormat:             "Jan 2, 2006",
		ModerateDelayThreshold: 7,  // 1 week
		HighDelayThreshold:     14, // 2 weeks
		ExtremeDelayThreshold:  30, // 1 month
	}
}

// WithDateFormat sets the date format option
func WithDateFormat(format string) func(*FormatterOptions) {
	return func(o *FormatterOptions) {
		o.DateFormat = format
	}
}

// WithModerateDelayThreshold sets the moderate delay threshold option
func WithModerateDelayThreshold(days int) func(*FormatterOptions) {
	return func(o *FormatterOptions) {
		o.ModerateDelayThreshold = days
	}
}

// WithHighDelayThreshold sets the high delay threshold option
func WithHighDelayThreshold(days int) func(*FormatterOptions) {
	return func(o *FormatterOptions) {
		o.HighDelayThreshold = days
	}
}

// WithExtremeDelayThreshold sets the extreme delay threshold option
func WithExtremeDelayThreshold(days int) func(*FormatterOptions) {
	return func(o *FormatterOptions) {
		o.ExtremeDelayThreshold = days
	}
}

// Alignment represents text alignment in table columns
type Alignment string

const (
	// AlignLeft aligns text to the left
	AlignLeft Alignment = "left"
	// AlignCenter centers the text
	AlignCenter Alignment = "center"
	// AlignRight aligns text to the right
	AlignRight Alignment = "right"
)

// TableColumn represents a column in a table with its formatting options
type TableColumn struct {
	Header    string    // Column header text
	Alignment Alignment // Column text alignment
}

// Table represents a generic table structure that can be rendered in different formats
type Table struct {
	Columns []TableColumn // Column definitions including headers and formatting
	Rows    [][]string    // Table rows (data only)
}

// Document represents a structured document with sections
type Document struct {
	Title    string
	Sections []Section
}

// Section represents a section in a document
type Section struct {
	Title string
	Table *Table // Optional table content
	Text  string // Optional text content
}
