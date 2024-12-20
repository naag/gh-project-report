package diff

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/naag/gh-project-report/pkg/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestStateFileToProjectDiff(t *testing.T) {
	// Create temporary directory for test files
	tempDir, err := os.MkdirTemp("", "gh-project-report-integration-test")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	// Create test states
	oldState := &types.ProjectState{
		Timestamp:     time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
		ProjectNumber: 123,
		Items: []types.Item{
			{
				ID:       "1",
				DateSpan: types.MustNewDateSpan("2024-01-01", "2024-01-10"),
				Attributes: map[string]interface{}{
					"Title":      "Existing Task",
					"status":     "Todo",
					"priority":   "Medium",
					"created_at": time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
					"updated_at": time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
				},
			},
			{
				ID:       "2",
				DateSpan: types.MustNewDateSpan("2024-01-05", "2024-01-15"),
				Attributes: map[string]interface{}{
					"Title":      "Task to be Removed",
					"status":     "Done",
					"created_at": time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
					"updated_at": time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
				},
			},
		},
	}

	newState := &types.ProjectState{
		Timestamp:     time.Date(2024, 1, 2, 0, 0, 0, 0, time.UTC),
		ProjectNumber: 123,
		Items: []types.Item{
			{
				ID:       "1",
				DateSpan: types.MustNewDateSpan("2024-01-01", "2024-01-15"), // Extended by 5 days
				Attributes: map[string]interface{}{
					"Title":      "Existing Task",
					"status":     "In Progress", // Changed status
					"priority":   "High",        // Changed priority
					"created_at": time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
					"updated_at": time.Date(2024, 1, 2, 0, 0, 0, 0, time.UTC),
				},
			},
			{
				ID:       "3",
				DateSpan: types.MustNewDateSpan("2024-01-10", "2024-01-20"),
				Attributes: map[string]interface{}{
					"Title":      "New Task",
					"status":     "Todo",
					"created_at": time.Date(2024, 1, 2, 0, 0, 0, 0, time.UTC),
					"updated_at": time.Date(2024, 1, 2, 0, 0, 0, 0, time.UTC),
				},
			},
		},
	}

	// Create project directory
	projectDir := filepath.Join(tempDir, "states", "project=123")
	err = os.MkdirAll(projectDir, 0755)
	require.NoError(t, err)

	// Write states to files
	oldStatePath := filepath.Join(projectDir, fmt.Sprintf("%d.json", oldState.Timestamp.Unix()))
	newStatePath := filepath.Join(projectDir, fmt.Sprintf("%d.json", newState.Timestamp.Unix()))

	oldStateData, err := json.MarshalIndent(oldState, "", "  ")
	require.NoError(t, err)
	err = os.WriteFile(oldStatePath, oldStateData, 0644)
	require.NoError(t, err)

	newStateData, err := json.MarshalIndent(newState, "", "  ")
	require.NoError(t, err)
	err = os.WriteFile(newStatePath, newStateData, 0644)
	require.NoError(t, err)

	// Load and compare states
	oldStateLoaded := &types.ProjectState{}
	err = json.Unmarshal(oldStateData, oldStateLoaded)
	require.NoError(t, err)

	newStateLoaded := &types.ProjectState{}
	err = json.Unmarshal(newStateData, newStateLoaded)
	require.NoError(t, err)

	// Get the diff
	diff := types.CompareProjectStates(oldStateLoaded, newStateLoaded)

	// Verify the diff structure
	assert.Len(t, diff.AddedItems, 1, "should have one added item")
	assert.Len(t, diff.RemovedItems, 1, "should have one removed item")
	assert.Len(t, diff.ChangedItems, 1, "should have one changed item")

	// Verify added item
	assert.Equal(t, "3", diff.AddedItems[0].ID)
	assert.Equal(t, "New Task", diff.AddedItems[0].GetTitle())

	// Verify removed item
	assert.Equal(t, "2", diff.RemovedItems[0].ID)
	assert.Equal(t, "Task to be Removed", diff.RemovedItems[0].GetTitle())

	// Verify changed item
	changedItem := diff.ChangedItems[0]
	assert.Equal(t, "1", changedItem.ItemID)
	assert.Equal(t, "Existing Task", changedItem.Before.GetTitle())

	// Verify timeline change
	assert.NotNil(t, changedItem.DateChange)
	assert.Equal(t, 0, changedItem.DateChange.StartDaysDelta, "start date should not have changed")
	assert.Equal(t, 5, changedItem.DateChange.EndDaysDelta, "end date should have moved 5 days later")
	assert.Equal(t, 5, changedItem.DateChange.DurationDelta, "duration should have increased by 5 days")

	// Verify field changes
	assert.Len(t, changedItem.FieldChanges, 3, "should have three field changes (status, priority, and updated_at)")

	// Sort field changes for consistent comparison
	statusChange := changedItem.GetChangeForField("status")
	require.NotNil(t, statusChange)
	assert.Equal(t, "Todo", statusChange.OldValue)
	assert.Equal(t, "In Progress", statusChange.NewValue)

	priorityChange := changedItem.GetChangeForField("priority")
	require.NotNil(t, priorityChange)
	assert.Equal(t, "Medium", priorityChange.OldValue)
	assert.Equal(t, "High", priorityChange.NewValue)

	updatedAtChange := changedItem.GetChangeForField("updated_at")
	require.NotNil(t, updatedAtChange)

	// Convert string times back to time.Time for comparison
	oldTime, err := time.Parse(time.RFC3339, updatedAtChange.OldValue.(string))
	require.NoError(t, err)
	newTime, err := time.Parse(time.RFC3339, updatedAtChange.NewValue.(string))
	require.NoError(t, err)

	assert.Equal(t, time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC), oldTime)
	assert.Equal(t, time.Date(2024, 1, 2, 0, 0, 0, 0, time.UTC), newTime)
}
