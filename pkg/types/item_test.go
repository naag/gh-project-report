package types

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

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
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.before.CompareTo(tt.after)

			// Ignore timestamp in comparison
			got.Timestamp = time.Time{}

			assert.Equal(t, tt.wantDiff.ItemID, got.ItemID)
			assert.Equal(t, tt.wantDiff.DateChange, got.DateChange)
			assert.Equal(t, tt.wantDiff.FieldChanges, got.FieldChanges)
		})
	}
}

func TestItemDiffHelpers(t *testing.T) {
	item1 := Item{
		ID:       "test-1",
		DateSpan: MustNewDateSpan("2024-01-01", "2024-01-10"),
		Attributes: map[string]interface{}{
			"title":  "Original Title",
			"status": "open",
		},
	}

	item2 := Item{
		ID:       "test-1",
		DateSpan: MustNewDateSpan("2024-01-01", "2024-01-15"),
		Attributes: map[string]interface{}{
			"title":    "New Title",
			"priority": "high",
		},
	}

	diff := item1.CompareTo(item2)

	t.Run("HasChanges", func(t *testing.T) {
		assert.True(t, diff.HasChanges())
	})

	t.Run("HasDateChange", func(t *testing.T) {
		assert.True(t, diff.HasDateChange())
		assert.NotNil(t, diff.DateChange)
		assert.Equal(t, 0, diff.DateChange.StartDaysDelta)
		assert.Equal(t, 5, diff.DateChange.EndDaysDelta)
		assert.Equal(t, 5, diff.DateChange.DurationDelta)
	})

	t.Run("GetAttributeChanges", func(t *testing.T) {
		attrChanges := diff.GetAttributeChanges()
		assert.Len(t, attrChanges, 3) // title changed, status removed, priority added
	})

	t.Run("GetChangeForField", func(t *testing.T) {
		titleChange := diff.GetChangeForField("title")
		assert.NotNil(t, titleChange)
		assert.Equal(t, "Original Title", titleChange.OldValue)
		assert.Equal(t, "New Title", titleChange.NewValue)

		statusChange := diff.GetChangeForField("status")
		assert.NotNil(t, statusChange)
		assert.Equal(t, "open", statusChange.OldValue)
		assert.Nil(t, statusChange.NewValue)

		priorityChange := diff.GetChangeForField("priority")
		assert.NotNil(t, priorityChange)
		assert.Nil(t, priorityChange.OldValue)
		assert.Equal(t, "high", priorityChange.NewValue)
	})

	t.Run("GetChangedFieldNames", func(t *testing.T) {
		fieldNames := diff.GetChangedFieldNames()
		assert.ElementsMatch(t, []string{"title", "status", "priority"}, fieldNames)
	})
}
