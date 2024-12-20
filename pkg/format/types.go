package format

import (
	"github.com/naag/gh-project-report/pkg/types"
)

// FormatterOptions contains configuration options for formatters
type FormatterOptions struct {
	DateFormat            string
	ModerateRiskThreshold int
	HighRiskThreshold     int
	ExtremeRiskThreshold  int
}

// Formatter interface defines methods that all formatters must implement
type Formatter interface {
	Format(diff types.ProjectDiff) string
}

// RiskLevel represents the risk level of a timeline change
type RiskLevel string

const (
	RiskLevelOnTrack  RiskLevel = "🔵 On track"
	RiskLevelAhead    RiskLevel = "🚀 Ahead of schedule"
	RiskLevelModerate RiskLevel = "🟠 Moderate risk"
	RiskLevelHigh     RiskLevel = "🔴 High risk"
	RiskLevelExtreme  RiskLevel = "🚫 Extreme risk"
)

// DefaultOptions returns the default formatter options
func DefaultOptions() FormatterOptions {
	return FormatterOptions{
		DateFormat:            "Jan 2, 2006",
		ModerateRiskThreshold: 7,  // 1 week
		HighRiskThreshold:     14, // 2 weeks
		ExtremeRiskThreshold:  30, // 1 month
	}
}

// WithDateFormat sets the date format option
func WithDateFormat(format string) func(*FormatterOptions) {
	return func(o *FormatterOptions) {
		o.DateFormat = format
	}
}

// WithModerateRiskThreshold sets the moderate risk threshold option
func WithModerateRiskThreshold(days int) func(*FormatterOptions) {
	return func(o *FormatterOptions) {
		o.ModerateRiskThreshold = days
	}
}

// WithHighRiskThreshold sets the high risk threshold option
func WithHighRiskThreshold(days int) func(*FormatterOptions) {
	return func(o *FormatterOptions) {
		o.HighRiskThreshold = days
	}
}

// WithExtremeRiskThreshold sets the extreme risk threshold option
func WithExtremeRiskThreshold(days int) func(*FormatterOptions) {
	return func(o *FormatterOptions) {
		o.ExtremeRiskThreshold = days
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
