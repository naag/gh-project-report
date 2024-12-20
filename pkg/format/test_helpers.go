package format

import (
	"time"

	"github.com/naag/gh-project-report/pkg/types"
)

// Helper function to create test data
func createTestDiff() types.ProjectDiff {
	now := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)

	return types.ProjectDiff{
		AddedItems: []types.Item{
			{
				ID:       "new-1",
				DateSpan: types.MustNewDateSpan("2024-01-01", "2024-01-31"),
				Attributes: map[string]interface{}{
					"Title":      "New Task",
					"status":     "Todo",
					"priority":   "High",
					"created_at": now,
					"updated_at": now,
				},
			},
		},
		RemovedItems: []types.Item{
			{
				ID:       "removed-1",
				DateSpan: types.MustNewDateSpan("2024-01-01", "2024-01-15"),
				Attributes: map[string]interface{}{
					"Title":      "Removed Task",
					"status":     "Done",
					"created_at": now,
					"updated_at": now,
				},
			},
		},
		ChangedItems: []types.ItemDiff{
			{
				ItemID: "changed-1",
				Before: types.Item{
					ID:       "changed-1",
					DateSpan: types.MustNewDateSpan("2024-01-01", "2024-01-15"),
					Attributes: map[string]interface{}{
						"Title":      "Changed Task",
						"status":     "Todo",
						"priority":   "Medium",
						"created_at": now,
						"updated_at": now,
					},
				},
				After: types.Item{
					ID:       "changed-1",
					DateSpan: types.MustNewDateSpan("2024-01-01", "2024-01-31"),
					Attributes: map[string]interface{}{
						"Title":      "Changed Task",
						"status":     "In Progress",
						"priority":   "High",
						"created_at": now,
						"updated_at": now.Add(24 * time.Hour),
					},
				},
				DateChange: &types.DateSpanChange{
					StartDaysDelta: 0,
					EndDaysDelta:   16,
					DurationDelta:  8, // Changed to be in the moderate risk range (7-14 days)
				},
				FieldChanges: []types.FieldChange{
					{
						Field:    "status",
						OldValue: "Todo",
						NewValue: "In Progress",
					},
					{
						Field:    "priority",
						OldValue: "Medium",
						NewValue: "High",
					},
					{
						Field:    "updated_at",
						OldValue: now,
						NewValue: now.Add(24 * time.Hour),
					},
				},
			},
		},
	}
}
