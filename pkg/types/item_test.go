package types

import (
	"sort"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestItemHelpers(t *testing.T) {
	now := time.Now()
	item := Item{
		ID:       "test-1",
		DateSpan: MustNewDateSpan("2024-01-01", "2024-01-10"),
		Attributes: map[string]interface{}{
			"Title":      "Test Item",
			"status":     "In Progress",
			"created_at": now,
			"updated_at": now,
			"priority":   "high",
		},
	}

	t.Run("GetTitle", func(t *testing.T) {
		assert.Equal(t, "Test Item", item.GetTitle())
	})

	t.Run("GetStatus", func(t *testing.T) {
		assert.Equal(t, "In Progress", item.GetStatus())
	})

	t.Run("GetCreatedAt", func(t *testing.T) {
		assert.Equal(t, now, item.GetCreatedAt())
	})

	t.Run("GetUpdatedAt", func(t *testing.T) {
		assert.Equal(t, now, item.GetUpdatedAt())
	})

	t.Run("missing attributes", func(t *testing.T) {
		emptyItem := Item{
			ID:         "test-2",
			Attributes: map[string]interface{}{},
		}
		assert.Empty(t, emptyItem.GetTitle())
		assert.Empty(t, emptyItem.GetStatus())
		assert.True(t, emptyItem.GetCreatedAt().IsZero())
		assert.True(t, emptyItem.GetUpdatedAt().IsZero())
	})

	t.Run("wrong type attributes", func(t *testing.T) {
		wrongItem := Item{
			ID: "test-3",
			Attributes: map[string]interface{}{
				"Title":      123,               // Not a string
				"status":     true,              // Not a string
				"created_at": "not a time.Time", // Not a time.Time
				"updated_at": 456,               // Not a time.Time
			},
		}
		assert.Empty(t, wrongItem.GetTitle())
		assert.Empty(t, wrongItem.GetStatus())
		assert.True(t, wrongItem.GetCreatedAt().IsZero())
		assert.True(t, wrongItem.GetUpdatedAt().IsZero())
	})
}

func TestItemComparison(t *testing.T) {
	// Create base item
	baseItem := Item{
		ID:       "test-1",
		DateSpan: MustNewDateSpan("2024-01-01", "2024-01-10"),
		Attributes: map[string]interface{}{
			"title":  "Original Title",
			"status": "open",
		},
	}

	tests := []struct {
		name     string
		before   Item
		after    Item
		wantDiff ItemDiff
	}{
		{
			name:   "no changes",
			before: baseItem,
			after:  baseItem,
			wantDiff: ItemDiff{
				ItemID:       "test-1",
				Before:       baseItem,
				After:        baseItem,
				DateChange:   nil,
				FieldChanges: nil,
			},
		},
		{
			name:   "date change only",
			before: baseItem,
			after: Item{
				ID:       "test-1",
				DateSpan: MustNewDateSpan("2024-01-01", "2024-01-15"),
				Attributes: map[string]interface{}{
					"title":  "Original Title",
					"status": "open",
				},
			},
			wantDiff: ItemDiff{
				ItemID: "test-1",
				DateChange: &DateSpanChange{
					StartDaysDelta: 0, // Start date didn't change
					EndDaysDelta:   5, // End date moved 5 days later
					DurationDelta:  5, // Duration increased by 5 days
				},
				FieldChanges: nil,
			},
		},
		{
			name: "attribute changes",
			before: Item{
				ID: "test-1",
				Attributes: map[string]interface{}{
					"title":  "Original Title",
					"status": "open",
				},
			},
			after: Item{
				ID: "test-1",
				Attributes: map[string]interface{}{
					"title":  "New Title",
					"status": "closed",
				},
			},
			wantDiff: ItemDiff{
				ItemID: "test-1",
				FieldChanges: []FieldChange{
					{
						Field:    "status",
						OldValue: "open",
						NewValue: "closed",
					},
					{
						Field:    "title",
						OldValue: "Original Title",
						NewValue: "New Title",
					},
				},
			},
		},
		{
			name: "deleted attributes",
			before: Item{
				ID: "test-1",
				Attributes: map[string]interface{}{
					"title":    "Original Title",
					"status":   "open",
					"priority": "high",
				},
			},
			after: Item{
				ID: "test-1",
				Attributes: map[string]interface{}{
					"title": "Original Title",
				},
			},
			wantDiff: ItemDiff{
				ItemID: "test-1",
				FieldChanges: []FieldChange{
					{
						Field:    "priority",
						OldValue: "high",
						NewValue: nil,
					},
					{
						Field:    "status",
						OldValue: "open",
						NewValue: nil,
					},
				},
			},
		},
		{
			name: "added attributes",
			before: Item{
				ID: "test-1",
				Attributes: map[string]interface{}{
					"title": "Original Title",
				},
			},
			after: Item{
				ID: "test-1",
				Attributes: map[string]interface{}{
					"title":    "Original Title",
					"priority": "high",
					"status":   "open",
				},
			},
			wantDiff: ItemDiff{
				ItemID: "test-1",
				FieldChanges: []FieldChange{
					{
						Field:    "priority",
						OldValue: nil,
						NewValue: "high",
					},
					{
						Field:    "status",
						OldValue: nil,
						NewValue: "open",
					},
				},
			},
		},
		{
			name: "combined date and attribute changes",
			before: Item{
				ID:       "test-1",
				DateSpan: MustNewDateSpan("2024-01-01", "2024-01-10"),
				Attributes: map[string]interface{}{
					"title":    "Original Title",
					"status":   "open",
					"priority": "low",
				},
			},
			after: Item{
				ID:       "test-1",
				DateSpan: MustNewDateSpan("2024-01-05", "2024-01-15"),
				Attributes: map[string]interface{}{
					"title":  "New Title",
					"status": "closed",
				},
			},
			wantDiff: ItemDiff{
				ItemID: "test-1",
				DateChange: &DateSpanChange{
					StartDaysDelta: 4, // Start date moved 4 days later
					EndDaysDelta:   5, // End date moved 5 days later
					DurationDelta:  1, // Duration increased by 1 day
				},
				FieldChanges: []FieldChange{
					{
						Field:    "priority",
						OldValue: "low",
						NewValue: nil,
					},
					{
						Field:    "status",
						OldValue: "open",
						NewValue: "closed",
					},
					{
						Field:    "title",
						OldValue: "Original Title",
						NewValue: "New Title",
					},
				},
			},
		},
		{
			name: "nil maps",
			before: Item{
				ID: "test-1",
			},
			after: Item{
				ID: "test-1",
				Attributes: map[string]interface{}{
					"title": "New Title",
				},
			},
			wantDiff: ItemDiff{
				ItemID: "test-1",
				FieldChanges: []FieldChange{
					{
						Field:    "title",
						OldValue: nil,
						NewValue: "New Title",
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.before.CompareTo(tt.after)

			// Ignore timestamp in comparison
			got.Timestamp = time.Time{}
			tt.wantDiff.Before = tt.before
			tt.wantDiff.After = tt.after

			assert.Equal(t, tt.wantDiff, got)
		})
	}
}

func TestItemDiffHelpers(t *testing.T) {
	// Create test items
	oldItem := Item{
		ID:       "test-1",
		DateSpan: MustNewDateSpan("2024-01-01", "2024-01-10"),
		Attributes: map[string]interface{}{
			"Title":    "Test Item",
			"Status":   "In Progress",
			"Priority": "High",
		},
	}

	newItem := Item{
		ID:       "test-1",
		DateSpan: MustNewDateSpan("2024-01-05", "2024-01-15"), // Changed date
		Attributes: map[string]interface{}{
			"Title":    "Test Item Updated",
			"Status":   "Done",
			"Priority": "High",
		},
	}

	diff := oldItem.CompareTo(newItem)

	t.Run("GetDateChange", func(t *testing.T) {
		dateChange := diff.GetDateChange()
		assert.NotNil(t, dateChange)
		assert.Equal(t, 4, dateChange.StartDaysDelta)
		assert.Equal(t, 5, dateChange.EndDaysDelta)
		assert.Equal(t, 1, dateChange.DurationDelta)
	})

	t.Run("GetDateChange with no change", func(t *testing.T) {
		sameItem := oldItem
		noDiff := oldItem.CompareTo(sameItem)
		assert.Nil(t, noDiff.GetDateChange())
	})

	t.Run("HasChanges", func(t *testing.T) {
		assert.True(t, diff.HasChanges())
	})

	t.Run("HasDateChange", func(t *testing.T) {
		assert.True(t, diff.HasDateChange())
		assert.Equal(t, 4, diff.GetDateChange().StartDaysDelta)
		assert.Equal(t, 5, diff.GetDateChange().EndDaysDelta)
		assert.Equal(t, 1, diff.GetDateChange().DurationDelta)

		// Test with no changes
		noDiff := oldItem.CompareTo(oldItem)
		assert.False(t, noDiff.HasDateChange())
	})

	t.Run("GetAttributeChanges", func(t *testing.T) {
		changes := diff.GetAttributeChanges()
		assert.Len(t, changes, 2)

		// Sort changes by field name for consistent testing
		sort.Slice(changes, func(i, j int) bool {
			return changes[i].Field < changes[j].Field
		})

		// Status changed from "In Progress" to "Done"
		assert.Equal(t, "Status", changes[0].Field)
		assert.Equal(t, "In Progress", changes[0].OldValue)
		assert.Equal(t, "Done", changes[0].NewValue)

		// Title changed
		assert.Equal(t, "Title", changes[1].Field)
		assert.Equal(t, "Test Item", changes[1].OldValue)
		assert.Equal(t, "Test Item Updated", changes[1].NewValue)
	})

	t.Run("GetChangeForField", func(t *testing.T) {
		// Test existing field change
		statusChange := diff.GetChangeForField("Status")
		assert.NotNil(t, statusChange)
		assert.Equal(t, "Status", statusChange.Field)
		assert.Equal(t, "In Progress", statusChange.OldValue)
		assert.Equal(t, "Done", statusChange.NewValue)

		// Test field with no change
		priorityChange := diff.GetChangeForField("Priority")
		assert.Nil(t, priorityChange)

		// Test non-existent field
		nonExistentChange := diff.GetChangeForField("NonExistent")
		assert.Nil(t, nonExistentChange)
	})

	t.Run("GetChangedFieldNames", func(t *testing.T) {
		changedFields := diff.GetChangedFieldNames()
		assert.Len(t, changedFields, 2)

		// Sort for consistent testing
		sort.Strings(changedFields)
		assert.Equal(t, []string{"Status", "Title"}, changedFields)

		// Test with no changes
		noDiff := oldItem.CompareTo(oldItem)
		assert.Empty(t, noDiff.GetChangedFieldNames())
	})
}
