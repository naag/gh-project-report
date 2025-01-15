package format

import (
	"fmt"
	"sort"
	"strings"

	"github.com/naag/gh-project-report/pkg/types"
	"github.com/olekukonko/tablewriter"
)

// PlainTableFormatter formats project diffs as a plain table
type PlainTableFormatter struct {
	options FormatterOptions
}

// NewPlainTableFormatter creates a new plain table formatter with the given options
func NewPlainTableFormatter(opts ...func(*FormatterOptions)) *PlainTableFormatter {
	options := DefaultOptions()
	for _, opt := range opts {
		opt(&options)
	}
	return &PlainTableFormatter{
		options: options,
	}
}

// Format formats the project diff as a plain table
func (f *PlainTableFormatter) Format(diff types.ProjectDiff) string {
	if len(diff.AddedItems) == 0 && len(diff.RemovedItems) == 0 && len(diff.ChangedItems) == 0 {
		return "No changes found in the project timeline."
	}

	doc := Document{
		Title: "Project Timeline Analysis",
	}

	// Timeline changes section
	timelineTable := &Table{
		Columns: []TableColumn{
			{Header: "Task", Alignment: AlignLeft},
			{Header: "Status", Alignment: AlignCenter},
			{Header: "Details", Alignment: AlignLeft},
			{Header: "Start Date", Alignment: AlignRight},
			{Header: "End Date", Alignment: AlignRight},
			{Header: "Duration", Alignment: AlignRight},
		},
	}

	// Added items
	for _, item := range diff.AddedItems {
		title := item.GetTitle()
		duration := formatHumanDuration(item.DateSpan.DurationDays())
		timelineTable.Rows = append(timelineTable.Rows, []string{
			title,
			"Added",
			"New task",
			formatDate(item.DateSpan.Start, f.options.DateFormat),
			formatDate(item.DateSpan.End, f.options.DateFormat),
			duration,
		})
	}

	// Removed items
	for _, item := range diff.RemovedItems {
		title := item.GetTitle()
		duration := formatHumanDuration(item.DateSpan.DurationDays())
		timelineTable.Rows = append(timelineTable.Rows, []string{
			title,
			"Removed",
			"Task removed",
			formatDate(item.DateSpan.Start, f.options.DateFormat),
			formatDate(item.DateSpan.End, f.options.DateFormat),
			duration,
		})
	}

	// Changed items
	for _, change := range diff.ChangedItems {
		title := change.After.GetTitle()

		// Handle timeline changes via DateSpan only
		if change.DateChange != nil {
			delay := calculateTimelineDelayLevel(
				change.DateChange.StartDaysDelta,
				change.DateChange.DurationDelta,
				f.options.ModerateDelayThreshold,
				f.options.HighDelayThreshold,
				f.options.ExtremeDelayThreshold,
			)
			details := formatTimelineDetails(change.DateChange, change.Before.DateSpan, change.After.DateSpan)
			afterDuration := formatHumanDuration(change.After.DateSpan.DurationDays())
			durationDiff := ""
			if change.DateChange.DurationDelta != 0 {
				durationDiff = fmt.Sprintf(" (%+d days)",
					change.DateChange.DurationDelta,
				)
			}

			timelineTable.Rows = append(timelineTable.Rows, []string{
				title,
				string(delay),
				details,
				formatDateWithChange(change.After.DateSpan.Start, change.Before.DateSpan.Start, f.options.DateFormat),
				formatDateWithChange(change.After.DateSpan.End, change.Before.DateSpan.End, f.options.DateFormat),
				fmt.Sprintf("%s%s", afterDuration, durationDiff),
			})
		}
	}

	if len(timelineTable.Rows) > 0 {
		doc.Sections = append(doc.Sections, Section{
			Title: "📅 Timeline Changes",
			Table: timelineTable,
		})
	}

	// Other changes section
	if hasFieldChanges(diff.ChangedItems) {
		// First, collect all unique field names that changed
		fieldNames := make(map[string]bool)
		for _, change := range diff.ChangedItems {
			for _, fieldChange := range change.FieldChanges {
				if fieldChange.Field != "start" && fieldChange.Field != "end" &&
					fieldChange.Field != "updated_at" && fieldChange.Field != "created_at" {
					fieldNames[fieldChange.Field] = true
				}
			}
		}

		// Create columns
		columns := []TableColumn{{Header: "Task", Alignment: AlignLeft}}
		// Sort field names for consistent column order
		var sortedFields []string
		for field := range fieldNames {
			sortedFields = append(sortedFields, field)
		}
		sort.Strings(sortedFields)
		for _, field := range sortedFields {
			columns = append(columns, TableColumn{Header: field, Alignment: AlignCenter})
		}

		otherTable := &Table{Columns: columns}

		// Add item changes
		for _, change := range diff.ChangedItems {
			if len(change.FieldChanges) > 0 {
				hasNonTimeChange := false
				row := make([]string, len(columns))
				row[0] = change.After.GetTitle()
				// Fill all fields with "-" by default
				for i := 1; i < len(columns); i++ {
					row[i] = "-"
				}

				// Fill in the actual changes
				for _, fieldChange := range change.FieldChanges {
					if fieldChange.Field != "start" && fieldChange.Field != "end" &&
						fieldChange.Field != "updated_at" && fieldChange.Field != "created_at" {
						hasNonTimeChange = true
						// Find the column index for this field
						for i, field := range sortedFields {
							if field == fieldChange.Field {
								row[i+1] = fmt.Sprintf("%v → %v", fieldChange.OldValue, fieldChange.NewValue)
								break
							}
						}
					}
				}

				// Only add the row if there are actual non-time changes
				if hasNonTimeChange {
					otherTable.Rows = append(otherTable.Rows, row)
				}
			}
		}

		if len(otherTable.Rows) > 0 {
			doc.Sections = append(doc.Sections, Section{
				Title: "📋 Other Changes",
				Table: otherTable,
			})
		}
	}

	return f.renderDocument(&doc)
}

// renderDocument converts a Document to plain text format
func (f *PlainTableFormatter) renderDocument(d *Document) string {
	var sb strings.Builder

	if d.Title != "" {
		sb.WriteString(d.Title + "\n\n")
	}

	for _, section := range d.Sections {
		sb.WriteString(f.renderSection(&section) + "\n")
	}

	return sb.String()
}

// renderSection converts a Section to plain text format
func (f *PlainTableFormatter) renderSection(s *Section) string {
	var sb strings.Builder

	if s.Title != "" {
		sb.WriteString(s.Title + "\n\n")
	}

	if s.Table != nil {
		sb.WriteString(f.renderTable(s.Table))
	} else if s.Text != "" {
		sb.WriteString(s.Text + "\n")
	}

	return sb.String()
}

// renderTable converts a Table to plain text format using tablewriter
func (f *PlainTableFormatter) renderTable(t *Table) string {
	if len(t.Columns) == 0 {
		return ""
	}

	var buf strings.Builder
	table := tablewriter.NewWriter(&buf)

	// Set headers
	headers := make([]string, len(t.Columns))
	for i, col := range t.Columns {
		headers[i] = col.Header
	}
	table.SetHeader(headers)

	// Set alignments
	alignments := make([]int, len(t.Columns))
	for i, col := range t.Columns {
		switch col.Alignment {
		case AlignLeft:
			alignments[i] = tablewriter.ALIGN_LEFT
		case AlignRight:
			alignments[i] = tablewriter.ALIGN_RIGHT
		case AlignCenter:
			alignments[i] = tablewriter.ALIGN_CENTER
		default:
			alignments[i] = tablewriter.ALIGN_LEFT
		}
	}
	table.SetColumnAlignment(alignments)

	// Configure table style for plain text
	table.SetAutoWrapText(false)
	table.SetAutoFormatHeaders(true)
	table.SetHeaderAlignment(tablewriter.ALIGN_LEFT)
	table.SetBorder(true)
	table.SetCenterSeparator("")
	table.SetColumnSeparator("│")
	table.SetRowSeparator("-")
	table.SetHeaderLine(true)
	table.SetTablePadding("  ")
	table.SetNoWhiteSpace(true)

	// Add rows
	for _, row := range t.Rows {
		// Ensure row has same number of columns as headers
		paddedRow := make([]string, len(t.Columns))
		for i := range t.Columns {
			if i < len(row) {
				paddedRow[i] = row[i]
			} else {
				paddedRow[i] = "-"
			}
		}
		table.Append(paddedRow)
	}

	table.Render()
	return buf.String()
}
