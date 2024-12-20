package types

import (
	"sort"
	"time"
)

// Item represents a single item at a point in time
type Item struct {
	ID         string
	DateSpan   DateSpan
	Attributes map[string]interface{}
}

// FieldChange represents what changed in a specific field
type FieldChange struct {
	Field    string
	OldValue interface{}
	NewValue interface{}
}

// ItemDiff captures the complete state change of an item
type ItemDiff struct {
	ItemID       string
	Timestamp    time.Time
	Before       Item
	After        Item
	DateChange   *DateSpanChange // Dedicated field for date changes
	FieldChanges []FieldChange   // Only for attribute changes
}

// CompareTo compares this item to another and returns an ItemDiff
func (i Item) CompareTo(other Item) ItemDiff {
	diff := ItemDiff{
		ItemID:    i.ID,
		Timestamp: time.Now(),
		Before:    i,
		After:     other,
	}

	// Compare DateSpan separately
	if !i.DateSpan.Equal(other.DateSpan) {
		dateChange := i.DateSpan.CompareTo(other.DateSpan)
		diff.DateChange = &dateChange
	}

	var changes []FieldChange

	// Check attribute changes and additions
	for key, newVal := range other.Attributes {
		oldVal, exists := i.Attributes[key]
		if !exists || oldVal != newVal {
			changes = append(changes, FieldChange{
				Field:    key,
				OldValue: oldVal,
				NewValue: newVal,
			})
		}
	}

	// Check for deleted attributes
	for key, oldVal := range i.Attributes {
		if _, exists := other.Attributes[key]; !exists {
			changes = append(changes, FieldChange{
				Field:    key,
				OldValue: oldVal,
				NewValue: nil,
			})
		}
	}

	// Sort field changes by field name for consistent ordering
	sort.Slice(changes, func(i, j int) bool {
		return changes[i].Field < changes[j].Field
	})

	diff.FieldChanges = changes
	return diff
}

// HasChanges returns true if any field changed
func (d ItemDiff) HasChanges() bool {
	return d.DateChange != nil || len(d.FieldChanges) > 0
}

// HasDateChange returns true if the DateSpan changed
func (d ItemDiff) HasDateChange() bool {
	return d.DateChange != nil
}

// GetDateChange returns the DateSpan change if it exists
func (d ItemDiff) GetDateChange() *DateSpanChange {
	return d.DateChange
}

// GetAttributeChanges returns all attribute changes
func (d ItemDiff) GetAttributeChanges() []FieldChange {
	return d.FieldChanges
}

// GetChangeForField returns the change for a specific attribute if it exists
func (d ItemDiff) GetChangeForField(fieldName string) *FieldChange {
	for _, change := range d.FieldChanges {
		if change.Field == fieldName {
			return &change
		}
	}
	return nil
}

// GetChangedFieldNames returns a slice of all changed attribute names
func (d ItemDiff) GetChangedFieldNames() []string {
	names := make([]string, len(d.FieldChanges))
	for i, change := range d.FieldChanges {
		names[i] = change.Field
	}
	return names
}

// Helper functions for accessing common attributes
func (i Item) GetTitle() string {
	if title, ok := i.Attributes["Title"].(string); ok {
		return title
	}
	return ""
}

func (i Item) GetStatus() string {
	if status, ok := i.Attributes["status"].(string); ok {
		return status
	}
	return ""
}

func (i Item) GetCreatedAt() time.Time {
	if createdAt, ok := i.Attributes["created_at"].(time.Time); ok {
		return createdAt
	}
	return time.Time{}
}

func (i Item) GetUpdatedAt() time.Time {
	if updatedAt, ok := i.Attributes["updated_at"].(time.Time); ok {
		return updatedAt
	}
	return time.Time{}
}
