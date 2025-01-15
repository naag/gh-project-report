package format

import (
	"bytes"
	"strings"

	"github.com/olekukonko/tablewriter"
)

// CLITableRenderer renders tables in CLI format using tablewriter
type CLITableRenderer struct{}

// NewCLITableRenderer creates a new CLI table renderer
func NewCLITableRenderer() *CLITableRenderer {
	return &CLITableRenderer{}
}

// RenderTable converts a generic Table to CLI format
func (r *CLITableRenderer) RenderTable(t *Table) string {
	if len(t.Columns) == 0 {
		return ""
	}

	var buf bytes.Buffer
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

	// Configure table style
	table.SetAutoWrapText(false)
	table.SetAutoFormatHeaders(true)
	table.SetHeaderAlignment(tablewriter.ALIGN_LEFT)
	table.SetBorder(false)
	table.SetCenterSeparator("")
	table.SetColumnSeparator("│")
	table.SetRowSeparator("")
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

// RenderSection converts a generic Section to CLI format
func (r *CLITableRenderer) RenderSection(s *Section) string {
	var sb strings.Builder

	if s.Title != "" {
		sb.WriteString(s.Title + "\n\n")
	}

	if s.Table != nil {
		sb.WriteString(r.RenderTable(s.Table))
	} else if s.Text != "" {
		sb.WriteString(s.Text + "\n")
	}

	return sb.String()
}

// RenderDocument converts a generic Document to CLI format
func (r *CLITableRenderer) RenderDocument(d *Document) string {
	var sb strings.Builder

	if d.Title != "" {
		sb.WriteString(d.Title + "\n\n")
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
