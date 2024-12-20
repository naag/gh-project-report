package storage

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/naag/gh-project-report/pkg/types"
	"github.com/stretchr/testify/assert"
)

func TestSaveAndLoadState(t *testing.T) {
	// Create temporary directory for test
	tempDir, err := os.MkdirTemp("", "gh-project-report-test")
	assert.NoError(t, err)
	defer os.RemoveAll(tempDir)

	// Create store
	store, err := NewStore(tempDir)
	assert.NoError(t, err)

	// Create test state
	now := time.Now()
	dateSpan := types.MustNewDateSpan("2024-01-01", "2024-01-10")
	state := &types.ProjectState{
		Timestamp:     now,
		ProjectNumber: 123,
		Items: []types.Item{
			{
				ID:       "test-1",
				DateSpan: dateSpan,
				Attributes: map[string]interface{}{
					"Title":      "Test Item",
					"status":     "In Progress",
					"created_at": now,
					"updated_at": now,
					"priority":   "high",
				},
			},
		},
	}

	// Save state
	filename, err := store.SaveState(state)
	assert.NoError(t, err)
	assert.NotEmpty(t, filename)

	// Verify file path format
	expectedPath := filepath.Join(tempDir, "states", "project=123", fmt.Sprintf("%d.json", now.Unix()))
	assert.Equal(t, expectedPath, filename)

	// Load state
	loadedState, err := store.LoadState(123, now)
	assert.NoError(t, err)
	assert.NotNil(t, loadedState)

	// Compare states
	assert.Equal(t, state.ProjectNumber, loadedState.ProjectNumber)
	assert.Equal(t, len(state.Items), len(loadedState.Items))

	// Compare items
	assert.Equal(t, state.Items[0].ID, loadedState.Items[0].ID)
	assert.Equal(t, state.Items[0].DateSpan, loadedState.Items[0].DateSpan)
	assert.Equal(t, state.Items[0].Attributes["Title"], loadedState.Items[0].Attributes["Title"])

	// Compare attributes individually to avoid time comparison issues
	for key, value := range state.Items[0].Attributes {
		if key != "created_at" && key != "updated_at" {
			assert.Equal(t, value, loadedState.Items[0].Attributes[key])
		}
	}
}

func TestValidateState(t *testing.T) {
	tests := []struct {
		name      string
		state     *types.ProjectState
		wantError bool
	}{
		{
			name: "valid state",
			state: &types.ProjectState{
				ProjectNumber: 123,
				Items: []types.Item{
					{
						ID: "test-1",
						Attributes: map[string]interface{}{
							"Title": "Test Item",
						},
					},
				},
			},
			wantError: false,
		},
		{
			name: "missing project number",
			state: &types.ProjectState{
				Items: []types.Item{
					{
						ID: "test-1",
						Attributes: map[string]interface{}{
							"Title": "Test Item",
						},
					},
				},
			},
			wantError: true,
		},
		{
			name: "missing item ID",
			state: &types.ProjectState{
				ProjectNumber: 123,
				Items: []types.Item{
					{
						Attributes: map[string]interface{}{
							"Title": "Test Item",
						},
					},
				},
			},
			wantError: true,
		},
		{
			name: "missing item title",
			state: &types.ProjectState{
				ProjectNumber: 123,
				Items: []types.Item{
					{
						ID:         "test-1",
						Attributes: map[string]interface{}{},
					},
				},
			},
			wantError: true,
		},
		{
			name: "nil field value",
			state: &types.ProjectState{
				ProjectNumber: 123,
				Items: []types.Item{
					{
						ID: "test-1",
						Attributes: map[string]interface{}{
							"Title":    "Test Item",
							"priority": nil,
						},
					},
				},
			},
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateState(tt.state)
			if tt.wantError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestFindClosestState(t *testing.T) {
	// Create a temporary directory for test files
	tempDir, err := os.MkdirTemp("", "storage_test")
	assert.NoError(t, err)
	defer os.RemoveAll(tempDir)

	store, err := NewStore(tempDir)
	assert.NoError(t, err)

	// Create test states
	timestamps := []time.Time{
		time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
		time.Date(2024, 1, 2, 0, 0, 0, 0, time.UTC),
		time.Date(2024, 1, 3, 0, 0, 0, 0, time.UTC),
	}

	for _, ts := range timestamps {
		state := &types.ProjectState{
			Timestamp:     ts,
			ProjectNumber: 123,
			Items: []types.Item{
				{
					ID: "test-1",
					Attributes: map[string]interface{}{
						"Title": "Test Item",
					},
				},
			},
		}
		_, err := store.SaveState(state)
		assert.NoError(t, err)
	}

	tests := []struct {
		name      string
		target    time.Time
		wantTime  time.Time
		wantError bool
	}{
		{
			name:     "exact match",
			target:   timestamps[1],
			wantTime: timestamps[1],
		},
		{
			name:     "between timestamps - closer to lower",
			target:   timestamps[0].Add(8 * time.Hour),
			wantTime: timestamps[0],
		},
		{
			name:     "between timestamps - closer to higher",
			target:   timestamps[1].Add(16 * time.Hour),
			wantTime: timestamps[2],
		},
		{
			name:     "before all timestamps",
			target:   timestamps[0].Add(-24 * time.Hour),
			wantTime: timestamps[0],
		},
		{
			name:     "after all timestamps",
			target:   timestamps[2].Add(24 * time.Hour),
			wantTime: timestamps[2],
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			filename, err := store.FindClosestState(123, tt.target)
			if tt.wantError {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)

			state, err := store.LoadStateFile(filename)
			assert.NoError(t, err)
			assert.Equal(t, tt.wantTime, state.Timestamp)
		})
	}
}

func TestLoadState(t *testing.T) {
	// Create a temporary directory for test files
	tempDir, err := os.MkdirTemp("", "storage_test")
	assert.NoError(t, err)
	defer os.RemoveAll(tempDir)

	store, err := NewStore(tempDir)
	assert.NoError(t, err)

	// Create test state
	timestamp := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	state := &types.ProjectState{
		Timestamp:     timestamp,
		ProjectNumber: 123,
		Items: []types.Item{
			{
				ID: "test-1",
				Attributes: map[string]interface{}{
					"Title": "Test Item",
				},
			},
		},
	}

	// Save state
	_, err = store.SaveState(state)
	assert.NoError(t, err)

	// Load state
	loadedState, err := store.LoadState(123, timestamp)
	assert.NoError(t, err)
	assert.Equal(t, state.Timestamp, loadedState.Timestamp)
	assert.Equal(t, state.ProjectNumber, loadedState.ProjectNumber)
	assert.Equal(t, state.Items[0].ID, loadedState.Items[0].ID)
	assert.Equal(t, state.Items[0].Attributes["Title"], loadedState.Items[0].Attributes["Title"])
}

func TestLoadStateErrors(t *testing.T) {
	// Create a temporary directory for test files
	tempDir, err := os.MkdirTemp("", "storage_test")
	assert.NoError(t, err)
	defer os.RemoveAll(tempDir)

	store, err := NewStore(tempDir)
	assert.NoError(t, err)

	// Test loading non-existent project
	_, err = store.LoadState(999, time.Now())
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to read project directory")
}

func TestStoreInProjectDirectory(t *testing.T) {
	// Create a temporary project directory
	tempDir, err := os.MkdirTemp("", "gh-project-report-project")
	assert.NoError(t, err)
	defer os.RemoveAll(tempDir)

	// Get the real path (resolves symlinks)
	realTempDir, err := filepath.EvalSymlinks(tempDir)
	assert.NoError(t, err)

	// Change to the project directory
	originalDir, err := os.Getwd()
	assert.NoError(t, err)
	err = os.Chdir(realTempDir)
	assert.NoError(t, err)
	defer os.Chdir(originalDir)

	// Create store without specifying a base directory
	store, err := NewStore("")
	assert.NoError(t, err)

	// Create test state
	now := time.Now()
	state := &types.ProjectState{
		Timestamp:     now,
		ProjectNumber: 123,
		Items: []types.Item{
			{
				ID: "test-1",
				Attributes: map[string]interface{}{
					"Title": "Test Item",
				},
			},
		},
	}

	// Save state
	filename, err := store.SaveState(state)
	assert.NoError(t, err)

	// Get the real path of the saved file
	realFilename, err := filepath.EvalSymlinks(filename)
	assert.NoError(t, err)

	// Verify file is in the project directory
	expectedPath := filepath.Join(realTempDir, "states", "project=123", fmt.Sprintf("%d.json", now.Unix()))
	assert.Equal(t, expectedPath, realFilename)

	// Verify file exists
	_, err = os.Stat(realFilename)
	assert.NoError(t, err)

	// Load state
	loadedState, err := store.LoadState(123, now)
	assert.NoError(t, err)
	assert.Equal(t, state.ProjectNumber, loadedState.ProjectNumber)
	assert.Equal(t, state.Items[0].ID, loadedState.Items[0].ID)
}
