package format

import (
	"fmt"
	"strings"

	"github.com/naag/gh-project-report/pkg/types"
)

// TextFormatter formats project diffs as plain text
type TextFormatter struct {
	options FormatterOptions
}

// NewTextFormatter creates a new text formatter with the given options
func NewTextFormatter(opts ...func(*FormatterOptions)) *TextFormatter {
	options := DefaultOptions()
	for _, opt := range opts {
		opt(&options)
	}
	return &TextFormatter{options: options}
}

// Format formats the project diff as plain text
func (f *TextFormatter) Format(diff types.ProjectDiff) string {
	if len(diff.AddedItems) == 0 && len(diff.RemovedItems) == 0 && len(diff.ChangedItems) == 0 {
		return "No changes found in the project timeline."
	}

	var sb strings.Builder

	// Added items
	if len(diff.AddedItems) > 0 {
		sb.WriteString("Added Items:\n")
		for _, item := range diff.AddedItems {
			title := item.GetTitle()
			duration := item.DateSpan.DurationDays()
			sb.WriteString(fmt.Sprintf("- %s\n", title))
			sb.WriteString(fmt.Sprintf("  Status: Added\n"))
			sb.WriteString(fmt.Sprintf("  Timeline: %s → %s (%s)\n",
				formatDate(item.DateSpan.Start, f.options.DateFormat),
				formatDate(item.DateSpan.End, f.options.DateFormat),
				formatHumanDuration(duration),
			))
			sb.WriteString(f.formatAttributes(item.Attributes))
			sb.WriteString("\n")
		}
	}

	// Removed items
	if len(diff.RemovedItems) > 0 {
		sb.WriteString("Removed Items:\n")
		for _, item := range diff.RemovedItems {
			title := item.GetTitle()
			duration := item.DateSpan.DurationDays()
			sb.WriteString(fmt.Sprintf("- %s\n", title))
			sb.WriteString(fmt.Sprintf("  Status: Removed\n"))
			sb.WriteString(fmt.Sprintf("  Timeline: %s → %s (%s)\n",
				formatDate(item.DateSpan.Start, f.options.DateFormat),
				formatDate(item.DateSpan.End, f.options.DateFormat),
				formatHumanDuration(duration),
			))
			sb.WriteString(f.formatAttributes(item.Attributes))
			sb.WriteString("\n")
		}
	}

	// Changed items
	if len(diff.ChangedItems) > 0 {
		sb.WriteString("Changed Items:\n")
		for _, change := range diff.ChangedItems {
			title := change.After.GetTitle()
			sb.WriteString(fmt.Sprintf("- %s\n", title))

			// Timeline changes
			if change.DateChange != nil {
				delay := calculateTimelineDelayLevel(
					change.DateChange.StartDaysDelta,
					change.DateChange.DurationDelta,
					f.options.ModerateDelayThreshold,
					f.options.HighDelayThreshold,
					f.options.ExtremeDelayThreshold,
				)
				sb.WriteString(fmt.Sprintf("  Timeline: %s %s\n",
					string(delay),
					formatHumanDuration(change.DateChange.DurationDelta),
				))
				sb.WriteString(fmt.Sprintf("  Before: %s → %s\n",
					formatDate(change.Before.DateSpan.Start, f.options.DateFormat),
					formatDate(change.Before.DateSpan.End, f.options.DateFormat),
				))
				sb.WriteString(fmt.Sprintf("  After:  %s → %s\n",
					formatDate(change.After.DateSpan.Start, f.options.DateFormat),
					formatDate(change.After.DateSpan.End, f.options.DateFormat),
				))
			}

			// Field changes
			if len(change.FieldChanges) > 0 {
				sb.WriteString("  Changes:\n")
				for _, fieldChange := range change.FieldChanges {
					if fieldChange.Field == "updated_at" || fieldChange.Field == "created_at" {
						continue
					}
					sb.WriteString(fmt.Sprintf("    %s: %v → %v\n",
						fieldChange.Field,
						fieldChange.OldValue,
						fieldChange.NewValue,
					))
				}
			}
			sb.WriteString("\n")
		}
	}

	return sb.String()
}

// formatAttributes formats item attributes as a string
func (f *TextFormatter) formatAttributes(attrs map[string]interface{}) string {
	var sb strings.Builder
	for k, v := range attrs {
		if strings.ToLower(k) != "title" && strings.ToLower(k) != "created_at" && strings.ToLower(k) != "updated_at" {
			sb.WriteString(fmt.Sprintf("  %s: %v\n", k, v))
		}
	}
	return sb.String()
}
