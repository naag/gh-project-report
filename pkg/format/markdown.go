package format

import (
	"fmt"
	"strings"
	"time"

	"github.com/naag/gh-project-report/pkg/types"
)

// TableFormatter formats project diffs as a markdown table
type TableFormatter struct {
	options  FormatterOptions
	renderer *MarkdownRenderer
}

// NewTableFormatter creates a new table formatter with the given options
func NewTableFormatter(opts ...func(*FormatterOptions)) *TableFormatter {
	options := DefaultOptions()
	for _, opt := range opts {
		opt(&options)
	}
	return &TableFormatter{
		options:  options,
		renderer: &MarkdownRenderer{},
	}
}

// Format formats the project diff as a markdown table
func (f *TableFormatter) Format(diff types.ProjectDiff) string {
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
			risk := calculateTimelineRiskLevel(
				change.DateChange.StartDaysDelta,
				change.DateChange.DurationDelta,
				f.options.ModerateRiskThreshold,
				f.options.HighRiskThreshold,
				f.options.ExtremeRiskThreshold,
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
				string(risk),
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
		otherTable := &Table{
			Columns: []TableColumn{
				{Header: "Task", Alignment: AlignLeft},
				{Header: "Status", Alignment: AlignCenter},
				{Header: "Priority", Alignment: AlignCenter},
				{Header: "Owner", Alignment: AlignCenter},
			},
		}

		for _, change := range diff.ChangedItems {
			if len(change.FieldChanges) > 0 {
				title := change.After.GetTitle()
				row := []string{title, "-", "-", "-"}

				for _, fieldChange := range change.FieldChanges {
					// Skip start/end fields as they should be handled via DateSpan
					if fieldChange.Field == "start" || fieldChange.Field == "end" {
						continue
					}

					switch fieldChange.Field {
					case "status":
						row[1] = fmt.Sprintf("%v → %v", fieldChange.OldValue, fieldChange.NewValue)
					case "priority":
						row[2] = fmt.Sprintf("%v → %v", fieldChange.OldValue, fieldChange.NewValue)
					case "owner":
						row[3] = fmt.Sprintf("%v → %v", fieldChange.OldValue, fieldChange.NewValue)
					}
				}

				// Only add the row if there are actual changes (not just start/end)
				if row[1] != "-" || row[2] != "-" || row[3] != "-" {
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

	return f.renderer.RenderDocument(&doc)
}

// hasFieldChanges checks if there are any field changes in the changed items
func hasFieldChanges(changes []types.ItemDiff) bool {
	for _, change := range changes {
		if len(change.FieldChanges) > 0 {
			return true
		}
	}
	return false
}

// formatTimelineDetails formats the timeline change details
func formatTimelineDetails(change *types.DateSpanChange, before, after types.DateSpan) string {
	var parts []string
	if change.StartDaysDelta != 0 {
		verb := "delayed"
		if change.StartDaysDelta < 0 {
			verb = "moved earlier"
		}
		duration := formatHumanDuration(abs(change.StartDaysDelta))
		part := fmt.Sprintf("start %s by %s", verb, duration)
		parts = append(parts, part)
	}
	if change.DurationDelta != 0 && change.EndDaysDelta != 0 && change.EndDaysDelta != change.StartDaysDelta {
		verb := "increased"
		if change.DurationDelta < 0 {
			verb = "decreased"
		}
		duration := formatHumanDuration(abs(change.DurationDelta))
		part := fmt.Sprintf("duration %s by %s", verb, duration)
		parts = append(parts, part)
	}
	if len(parts) == 0 {
		return "No timeline changes"
	}
	result := strings.Join(parts, ", ")
	return strings.ToUpper(result[:1]) + result[1:]
}

// abs returns the absolute value of an integer
func abs(n int) int {
	if n < 0 {
		return -n
	}
	return n
}

// formatDateWithChange formats a date with its change, if any
func formatDateWithChange(after, before time.Time, format string) string {
	if after.Equal(before) {
		return formatDate(after, format)
	}
	return fmt.Sprintf("%s → %s",
		formatDate(before, format),
		formatDate(after, format),
	)
}

// MarkdownRenderer handles rendering generic types into markdown format
type MarkdownRenderer struct{}

// RenderTable converts a generic Table to markdown format
func (r *MarkdownRenderer) RenderTable(t *Table) string {
	if len(t.Columns) == 0 {
		return ""
	}

	var sb strings.Builder

	// Write headers
	sb.WriteString("|")
	for _, col := range t.Columns {
		sb.WriteString(" " + col.Header + " |")
	}
	sb.WriteString("\n")

	// Write separator with alignment indicators
	sb.WriteString("|")
	for _, col := range t.Columns {
		switch col.Alignment {
		case AlignLeft:
			sb.WriteString(":------|")
		case AlignRight:
			sb.WriteString("------:|")
		case AlignCenter:
			sb.WriteString(":-----:|")
		default:
			sb.WriteString("------|")
		}
	}
	sb.WriteString("\n")

	// Write rows
	for _, row := range t.Rows {
		sb.WriteString("|")
		// Ensure row has same number of columns as headers
		for i := range t.Columns {
			value := "-"
			if i < len(row) {
				value = row[i]
			}
			sb.WriteString(" " + value + " |")
		}
		sb.WriteString("\n")
	}

	return sb.String()
}

// RenderSection converts a generic Section to markdown format
func (r *MarkdownRenderer) RenderSection(s *Section) string {
	var sb strings.Builder

	if s.Title != "" {
		sb.WriteString("## " + s.Title + "\n\n")
	}

	if s.Table != nil {
		sb.WriteString(r.RenderTable(s.Table))
	} else if s.Text != "" {
		sb.WriteString(s.Text + "\n")
	}

	return sb.String()
}

// RenderDocument converts a generic Document to markdown format
func (r *MarkdownRenderer) RenderDocument(d *Document) string {
	var sb strings.Builder

	if d.Title != "" {
		sb.WriteString("# " + d.Title + "\n\n")
	}

	for _, section := range d.Sections {
		sb.WriteString(r.RenderSection(&section) + "\n")
	}

	// Always add a final newline for empty documents
	if d.Title == "" && len(d.Sections) == 0 {
		sb.WriteString("\n")
	}

	return sb.String()
}
