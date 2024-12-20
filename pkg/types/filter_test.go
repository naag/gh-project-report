package types

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// createTestState creates a test project state with predefined items
func createTestState() *ProjectState {
	now := time.Now()
	return &ProjectState{
		Filename:      "test.json",
		Timestamp:     now,
		ProjectNumber: 123,
		ProjectID:     "PVT_123",
		Organization:  "test-org",
		Items: []Item{
			{
				ID:       "1",
				DateSpan: MustNewDateSpan("2024-01-01", "2024-01-31"),
				Attributes: map[string]interface{}{
					"Title":    "Task 1",
					"Team":     "UI",
					"Priority": "High",
				},
			},
			{
				ID:       "2",
				DateSpan: MustNewDateSpan("2024-02-01", "2024-02-28"),
				Attributes: map[string]interface{}{
					"Title":    "Task 2",
					"Team":     "Backend",
					"Priority": "Medium",
				},
			},
			{
				ID:       "3",
				DateSpan: MustNewDateSpan("2024-03-01", "2024-03-31"),
				Attributes: map[string]interface{}{
					"Title":    "Task 3",
					"Team":     "UI",
					"Priority": "Low",
				},
			},
		},
	}
}

func TestFilterState(t *testing.T) {
	state := createTestState()

	tests := []struct {
		name          string
		filter        string
		wantErr       bool
		errMsg        string
		expectedCount int
		expectedIDs   []string
	}{
		{
			name:          "empty filter returns original state",
			filter:        "",
			wantErr:       false,
			expectedCount: 3,
			expectedIDs:   []string{"1", "2", "3"},
		},
		{
			name:          "filter by team UI",
			filter:        "Team=UI",
			wantErr:       false,
			expectedCount: 2,
			expectedIDs:   []string{"1", "3"},
		},
		{
			name:          "filter by team Backend",
			filter:        "Team=Backend",
			wantErr:       false,
			expectedCount: 1,
			expectedIDs:   []string{"2"},
		},
		{
			name:          "filter by priority High",
			filter:        "Priority=High",
			wantErr:       false,
			expectedCount: 1,
			expectedIDs:   []string{"1"},
		},
		{
			name:          "filter with no matches",
			filter:        "Team=DevOps",
			wantErr:       false,
			expectedCount: 0,
			expectedIDs:   []string{},
		},
		{
			name:          "filter by non-existent attribute",
			filter:        "NonExistent=Value",
			wantErr:       false,
			expectedCount: 0,
			expectedIDs:   []string{},
		},
		{
			name:    "invalid filter format",
			filter:  "InvalidFilter",
			wantErr: true,
			errMsg:  "invalid filter format: \"InvalidFilter\" (must be attribute=value)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			filtered, err := state.FilterState(tt.filter)
			if tt.wantErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.expectedCount, len(filtered.Items))

			// Check that all expected IDs are present
			actualIDs := make([]string, len(filtered.Items))
			for i, item := range filtered.Items {
				actualIDs[i] = item.ID
			}
			assert.ElementsMatch(t, tt.expectedIDs, actualIDs)

			// Check that metadata is preserved
			assert.Equal(t, state.Filename, filtered.Filename)
			assert.Equal(t, state.Timestamp, filtered.Timestamp)
			assert.Equal(t, state.ProjectNumber, filtered.ProjectNumber)
			assert.Equal(t, state.ProjectID, filtered.ProjectID)
			assert.Equal(t, state.Organization, filtered.Organization)
		})
	}
}

func TestFilterState_Integration(t *testing.T) {
	// Create two states with some overlapping items
	oldState := createTestState()
	newState := createTestState()

	// Modify some items in the new state
	newState.Items[0].Attributes["Team"] = "Backend"  // Move Task 1 from UI to Backend
	newState.Items[1].Attributes["Priority"] = "High" // Change Task 2's priority to High

	// Filter both states by Team=UI and compare
	oldFiltered, err := oldState.FilterState("Team=UI")
	require.NoError(t, err)
	newFiltered, err := newState.FilterState("Team=UI")
	require.NoError(t, err)

	diff := CompareProjectStates(oldFiltered, newFiltered)

	// Task 1 should be in RemovedItems (moved from UI to Backend)
	assert.Equal(t, 1, len(diff.RemovedItems))
	assert.Equal(t, "1", diff.RemovedItems[0].ID)

	// Task 3 should remain unchanged
	assert.Equal(t, 0, len(diff.ChangedItems))

	// Filter both states by Priority=High and compare
	oldFiltered, err = oldState.FilterState("Priority=High")
	require.NoError(t, err)
	newFiltered, err = newState.FilterState("Priority=High")
	require.NoError(t, err)

	diff = CompareProjectStates(oldFiltered, newFiltered)

	// Task 2 should be in AddedItems (priority changed to High)
	assert.Equal(t, 1, len(diff.AddedItems))
	assert.Equal(t, "2", diff.AddedItems[0].ID)
}
