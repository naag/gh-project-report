package format

import (
	"strings"
)

// MarkdownTableRenderer renders tables in markdown format
type MarkdownTableRenderer struct{}

// NewMarkdownTableRenderer creates a new markdown table renderer
func NewMarkdownTableRenderer() *MarkdownTableRenderer {
	return &MarkdownTableRenderer{}
}

// RenderTable converts a generic Table to markdown format
func (r *MarkdownTableRenderer) RenderTable(t *Table) string {
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
func (r *MarkdownTableRenderer) RenderSection(s *Section) string {
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
func (r *MarkdownTableRenderer) RenderDocument(d *Document) string {
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
