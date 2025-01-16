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

func (p *ProjectState) CompareTo(other *ProjectState) *ProjectDiff {
	diff := ProjectDiff{}

	// Find removed and changed items
	for _, oldItem := range p.Items {
		found := false
		for _, newItem := range other.Items {
			if oldItem.ID == newItem.ID {
				found = true
				itemDiff := oldItem.CompareTo(newItem)
				if itemDiff.HasChanges() {
					diff.ChangedItems = append(diff.ChangedItems, itemDiff)
				}
				break
			}
		}
		if !found {
			diff.RemovedItems = append(diff.RemovedItems, oldItem)
		}
	}

	// Find added items
	for _, newItem := range other.Items {
		found := false
		for _, oldItem := range p.Items {
			if newItem.ID == oldItem.ID {
				found = true
				break
			}
		}
		if !found {
			diff.AddedItems = append(diff.AddedItems, newItem)
		}
	}

	return &diff
}
