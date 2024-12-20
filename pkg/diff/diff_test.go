package diff

import (
	"testing"
	"time"

	"github.com/naag/gh-project-report/pkg/types"
	"github.com/stretchr/testify/assert"
)

func TestCompareStates(t *testing.T) {
	now := time.Now()
	tests := []struct {
		name      string
		fromState *types.ProjectState
		toState   *types.ProjectState
		want      types.ProjectDiff
	}{
		{
			name: "item modified and new item added",
			fromState: &types.ProjectState{
				Items: []types.Item{
					{
						ID:       "1",
						DateSpan: types.MustNewDateSpan("2024-01-01", "2024-01-10"),
						Attributes: map[string]interface{}{
							"title":  "Task 1",
							"status": "Todo",
						},
					},
				},
			},
			toState: &types.ProjectState{
				Items: []types.Item{
					{
						ID:       "1",
						DateSpan: types.MustNewDateSpan("2024-01-01", "2024-01-15"),
						Attributes: map[string]interface{}{
							"title":  "Task 1",
							"status": "In Progress",
						},
					},
					{
						ID: "2",
						Attributes: map[string]interface{}{
							"title":      "Task 2",
							"status":     "Todo",
							"created_at": now,
							"updated_at": now,
						},
					},
				},
			},
			want: types.ProjectDiff{
				AddedItems: []types.Item{
					{
						ID: "2",
						Attributes: map[string]interface{}{
							"title":      "Task 2",
							"status":     "Todo",
							"created_at": now,
							"updated_at": now,
						},
					},
				},
				ChangedItems: []types.ItemDiff{
					{
						ItemID: "1",
						Before: types.Item{
							ID:       "1",
							DateSpan: types.MustNewDateSpan("2024-01-01", "2024-01-10"),
							Attributes: map[string]interface{}{
								"title":  "Task 1",
								"status": "Todo",
							},
						},
						After: types.Item{
							ID:       "1",
							DateSpan: types.MustNewDateSpan("2024-01-01", "2024-01-15"),
							Attributes: map[string]interface{}{
								"title":  "Task 1",
								"status": "In Progress",
							},
						},
						DateChange: &types.DateSpanChange{
							StartDaysDelta: 0,
							EndDaysDelta:   5,
							DurationDelta:  5,
						},
						FieldChanges: []types.FieldChange{
							{
								Field:    "status",
								OldValue: "Todo",
								NewValue: "In Progress",
							},
						},
					},
				},
			},
		},
		{
			name: "item deleted",
			fromState: &types.ProjectState{
				Items: []types.Item{
					{
						ID:       "1",
						DateSpan: types.MustNewDateSpan("2024-01-01", "2024-01-10"),
						Attributes: map[string]interface{}{
							"title":  "Task 1",
							"status": "Todo",
						},
					},
				},
			},
			toState: &types.ProjectState{
				Items: []types.Item{},
			},
			want: types.ProjectDiff{
				RemovedItems: []types.Item{
					{
						ID:       "1",
						DateSpan: types.MustNewDateSpan("2024-01-01", "2024-01-10"),
						Attributes: map[string]interface{}{
							"title":  "Task 1",
							"status": "Todo",
						},
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := types.CompareProjectStates(tt.fromState, tt.toState)

			// Compare added items
			assert.Equal(t, len(tt.want.AddedItems), len(got.AddedItems))
			for i := range tt.want.AddedItems {
				assert.Equal(t, tt.want.AddedItems[i].ID, got.AddedItems[i].ID)
				assert.Equal(t, tt.want.AddedItems[i].DateSpan, got.AddedItems[i].DateSpan)
				assert.Equal(t, tt.want.AddedItems[i].Attributes, got.AddedItems[i].Attributes)
			}

			// Compare removed items
			assert.Equal(t, len(tt.want.RemovedItems), len(got.RemovedItems))
			for i := range tt.want.RemovedItems {
				assert.Equal(t, tt.want.RemovedItems[i].ID, got.RemovedItems[i].ID)
				assert.Equal(t, tt.want.RemovedItems[i].DateSpan, got.RemovedItems[i].DateSpan)
				assert.Equal(t, tt.want.RemovedItems[i].Attributes, got.RemovedItems[i].Attributes)
			}

			// Compare changed items
			assert.Equal(t, len(tt.want.ChangedItems), len(got.ChangedItems))
			for i := range tt.want.ChangedItems {
				// Ignore timestamp in comparison
				got.ChangedItems[i].Timestamp = time.Time{}

				assert.Equal(t, tt.want.ChangedItems[i].ItemID, got.ChangedItems[i].ItemID)
				assert.Equal(t, tt.want.ChangedItems[i].DateChange, got.ChangedItems[i].DateChange)
				assert.Equal(t, tt.want.ChangedItems[i].FieldChanges, got.ChangedItems[i].FieldChanges)
				assert.Equal(t, tt.want.ChangedItems[i].Before, got.ChangedItems[i].Before)
				assert.Equal(t, tt.want.ChangedItems[i].After, got.ChangedItems[i].After)
			}
		})
	}
}
