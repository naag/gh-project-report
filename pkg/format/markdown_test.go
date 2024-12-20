package format

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMarkdownRenderer_RenderTable(t *testing.T) {
	renderer := &MarkdownRenderer{}
	tests := []struct {
		name     string
		table    Table
		expected string
	}{
		{
			name: "empty table",
			table: Table{
				Columns: []TableColumn{},
				Rows:    [][]string{},
			},
			expected: "",
		},
		{
			name: "simple table with default alignment",
			table: Table{
				Columns: []TableColumn{
					{Header: "Name", Alignment: AlignLeft},
					{Header: "Age", Alignment: AlignRight},
				},
				Rows: [][]string{
					{"Alice", "25"},
					{"Bob", "30"},
				},
			},
			expected: `| Name | Age |
|:------|------:|
| Alice | 25 |
| Bob | 30 |
`,
		},
		{
			name: "table with mixed alignments",
			table: Table{
				Columns: []TableColumn{
					{Header: "Name", Alignment: AlignLeft},
					{Header: "Age", Alignment: AlignRight},
					{Header: "Status", Alignment: AlignCenter},
				},
				Rows: [][]string{
					{"Alice", "25"},
					{"Bob", "30", "Active"},
				},
			},
			expected: `| Name | Age | Status |
|:------|------:|:-----:|
| Alice | 25 | - |
| Bob | 30 | Active |
`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := renderer.RenderTable(&tt.table)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestMarkdownRenderer_RenderSection(t *testing.T) {
	renderer := &MarkdownRenderer{}
	tests := []struct {
		name     string
		section  Section
		expected string
	}{
		{
			name: "section with title only",
			section: Section{
				Title: "Test Section",
			},
			expected: "## Test Section\n\n",
		},
		{
			name: "section with text",
			section: Section{
				Title: "Test Section",
				Text:  "Some text content",
			},
			expected: "## Test Section\n\nSome text content\n",
		},
		{
			name: "section with table",
			section: Section{
				Title: "Test Section",
				Table: &Table{
					Columns: []TableColumn{
						{Header: "Name", Alignment: AlignLeft},
						{Header: "Age", Alignment: AlignRight},
					},
					Rows: [][]string{
						{"Alice", "25"},
					},
				},
			},
			expected: `## Test Section

| Name | Age |
|:------|------:|
| Alice | 25 |
`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := renderer.RenderSection(&tt.section)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestMarkdownRenderer_RenderDocument(t *testing.T) {
	renderer := &MarkdownRenderer{}
	tests := []struct {
		name     string
		doc      Document
		expected string
	}{
		{
			name:     "empty document",
			doc:      Document{},
			expected: "\n",
		},
		{
			name: "document with title only",
			doc: Document{
				Title: "Test Document",
			},
			expected: "# Test Document\n\n",
		},
		{
			name: "document with sections",
			doc: Document{
				Title: "Test Document",
				Sections: []Section{
					{
						Title: "Section 1",
						Text:  "Some text",
					},
					{
						Title: "Section 2",
						Table: &Table{
							Columns: []TableColumn{
								{Header: "Name", Alignment: AlignLeft},
								{Header: "Age", Alignment: AlignRight},
							},
							Rows: [][]string{
								{"Alice", "25"},
							},
						},
					},
				},
			},
			expected: `# Test Document

## Section 1

Some text

## Section 2

| Name | Age |
|:------|------:|
| Alice | 25 |

`,
		},
		{
			name: "document with multiple tables and sections",
			doc: Document{
				Title: "Project Analysis Report",
				Sections: []Section{
					{
						Title: "Timeline Changes",
						Table: &Table{
							Columns: []TableColumn{
								{Header: "Task", Alignment: AlignLeft},
								{Header: "Status", Alignment: AlignCenter},
								{Header: "Details", Alignment: AlignLeft},
								{Header: "Start Date", Alignment: AlignRight},
								{Header: "End Date", Alignment: AlignRight},
								{Header: "Duration", Alignment: AlignRight},
							},
							Rows: [][]string{
								{
									"Task 1",
									"🔵 On track",
									"Duration increased by 14 days",
									"Jan 15, 2024",
									"Feb 14, 2024 → Feb 28, 2024",
									"6 weeks (+14 days)",
								},
								{
									"Task 2",
									"🔴 High risk",
									"Start delayed by 21 days",
									"Jan 22, 2024 → Feb 12, 2024",
									"Feb 19, 2024 → Mar 11, 2024",
									"4 weeks (-21 days)",
								},
							},
						},
					},
					{
						Title: "Status Summary",
						Text:  "Overview of project status changes",
					},
					{
						Title: "Field Changes",
						Table: &Table{
							Columns: []TableColumn{
								{Header: "Task", Alignment: AlignLeft},
								{Header: "Status", Alignment: AlignCenter},
								{Header: "Priority", Alignment: AlignCenter},
								{Header: "Owner", Alignment: AlignCenter},
							},
							Rows: [][]string{
								{"Task 1", "🏗️ In Progress → ✅ Done", "-", "-"},
								{"Task 2", "-", "High → Medium", "Alice → Bob"},
							},
						},
					},
				},
			},
			expected: `# Project Analysis Report

## Timeline Changes

| Task | Status | Details | Start Date | End Date | Duration |
|:------|:-----:|:------|------:|------:|------:|
| Task 1 | 🔵 On track | Duration increased by 14 days | Jan 15, 2024 | Feb 14, 2024 → Feb 28, 2024 | 6 weeks (+14 days) |
| Task 2 | 🔴 High risk | Start delayed by 21 days | Jan 22, 2024 → Feb 12, 2024 | Feb 19, 2024 → Mar 11, 2024 | 4 weeks (-21 days) |

## Status Summary

Overview of project status changes

## Field Changes

| Task | Status | Priority | Owner |
|:------|:-----:|:-----:|:-----:|
| Task 1 | 🏗️ In Progress → ✅ Done | - | - |
| Task 2 | - | High → Medium | Alice → Bob |

`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := renderer.RenderDocument(&tt.doc)
			// Normalize line endings for comparison
			expected := strings.ReplaceAll(tt.expected, "\n", "\n")
			result = strings.ReplaceAll(result, "\n", "\n")
			assert.Equal(t, expected, result)
		})
	}
}
