package types

import (
	"fmt"
	"strings"
	"time"
)

// ProjectState represents the state of a project at a specific point in time
type ProjectState struct {
	Filename      string    `json:"filename"`
	Timestamp     time.Time `json:"timestamp"`
	ProjectNumber int       `json:"project_number,omitempty"`
	ProjectID     string    `json:"project_id,omitempty"`
	Organization  string    `json:"organization,omitempty"`
	Items         []Item    `json:"items"`
}

// ProjectDiff represents all changes between two project states
type ProjectDiff struct {
	AddedItems   []Item     // Items that are new in the target state
	RemovedItems []Item     // Items that were in source but not in target
	ChangedItems []ItemDiff // Items that exist in both states but changed
}

// FilterState returns a new ProjectState containing only items that match the filter
func (s *ProjectState) FilterState(filter string) (*ProjectState, error) {
	if filter == "" {
		return s, nil
	}

	// Parse filter in format "attribute=value"
	parts := strings.SplitN(filter, "=", 2)
	if len(parts) != 2 {
		return nil, fmt.Errorf("invalid filter format: %q (must be attribute=value)", filter)
	}
	attribute, value := parts[0], parts[1]

	// Create new state with filtered items
	filtered := &ProjectState{
		Filename:      s.Filename,
		Timestamp:     s.Timestamp,
		ProjectNumber: s.ProjectNumber,
		ProjectID:     s.ProjectID,
		Organization:  s.Organization,
		Items:         make([]Item, 0),
	}

	// Add items that match the filter
	for _, item := range s.Items {
		if itemValue, ok := item.Attributes[attribute]; ok {
			if fmt.Sprintf("%v", itemValue) == value {
				filtered.Items = append(filtered.Items, item)
			}
		}
	}

	return filtered, nil
}

// CompareProjectStates compares two project states and returns a ProjectDiff
func CompareProjectStates(oldState, newState *ProjectState) ProjectDiff {
	diff := ProjectDiff{}
	oldItems := make(map[string]Item)
	newItems := make(map[string]Item)

	// Create maps for easier lookup
	for _, item := range oldState.Items {
		oldItems[item.ID] = item
	}
	for _, item := range newState.Items {
		newItems[item.ID] = item
	}

	// Find removed items
	for id, oldItem := range oldItems {
		if _, exists := newItems[id]; !exists {
			diff.RemovedItems = append(diff.RemovedItems, oldItem)
		}
	}

	// Find added items
	for id, newItem := range newItems {
		if _, exists := oldItems[id]; !exists {
			diff.AddedItems = append(diff.AddedItems, newItem)
		}
	}

	// Find changed items
	for id, oldItem := range oldItems {
		if newItem, exists := newItems[id]; exists {
			itemDiff := oldItem.CompareTo(newItem)
			if itemDiff.HasChanges() {
				diff.ChangedItems = append(diff.ChangedItems, itemDiff)
			}
		}
	}

	return diff
}
