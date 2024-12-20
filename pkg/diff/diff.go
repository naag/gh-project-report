package diff

import (
	"fmt"

	"github.com/naag/gh-project-report/pkg/format"
	"github.com/naag/gh-project-report/pkg/storage"
	"github.com/naag/gh-project-report/pkg/types"
)

// CompareStates compares two project states and returns a formatted string
func CompareStates(oldState, newState *types.ProjectState) string {
	diff := types.CompareProjectStates(oldState, newState)
	formatter := format.NewTextFormatter()
	return formatter.Format(diff)
}

// CompareStatesHuman compares states using a human-readable time range
func CompareStatesHuman(timeRange string, projectNumber int) error {
	from, to, err := format.ParseHumanRange(timeRange)
	if err != nil {
		return fmt.Errorf("error parsing time range: %w", err)
	}

	// Create storage
	store, err := storage.NewStore("")
	if err != nil {
		return fmt.Errorf("failed to create storage: %w", err)
	}

	// Load states
	fromState, err := store.LoadState(projectNumber, from)
	if err != nil {
		return fmt.Errorf("failed to load from state: %w", err)
	}

	toState, err := store.LoadState(projectNumber, to)
	if err != nil {
		return fmt.Errorf("failed to load to state: %w", err)
	}

	// Compare states
	fmt.Print(CompareStates(fromState, toState))
	return nil
}
