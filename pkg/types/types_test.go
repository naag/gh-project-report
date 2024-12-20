package types

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestItemHelpers(t *testing.T) {
	now := time.Now()
	item := Item{
		ID:       "test-1",
		DateSpan: MustNewDateSpan("2024-01-01", "2024-01-10"),
		Attributes: map[string]interface{}{
			"Title":      "Test Item",
			"status":     "In Progress",
			"created_at": now,
			"updated_at": now,
			"priority":   "high",
		},
	}

	t.Run("GetTitle", func(t *testing.T) {
		assert.Equal(t, "Test Item", item.GetTitle())
	})

	t.Run("GetStatus", func(t *testing.T) {
		assert.Equal(t, "In Progress", item.GetStatus())
	})

	t.Run("GetCreatedAt", func(t *testing.T) {
		assert.Equal(t, now, item.GetCreatedAt())
	})

	t.Run("GetUpdatedAt", func(t *testing.T) {
		assert.Equal(t, now, item.GetUpdatedAt())
	})

	t.Run("missing attributes", func(t *testing.T) {
		emptyItem := Item{
			ID:         "test-2",
			Attributes: map[string]interface{}{},
		}
		assert.Empty(t, emptyItem.GetTitle())
		assert.Empty(t, emptyItem.GetStatus())
		assert.True(t, emptyItem.GetCreatedAt().IsZero())
		assert.True(t, emptyItem.GetUpdatedAt().IsZero())
	})

	t.Run("wrong type attributes", func(t *testing.T) {
		wrongItem := Item{
			ID: "test-3",
			Attributes: map[string]interface{}{
				"Title":      123,               // Not a string
				"status":     true,              // Not a string
				"created_at": "not a time.Time", // Not a time.Time
				"updated_at": 456,               // Not a time.Time
			},
		}
		assert.Empty(t, wrongItem.GetTitle())
		assert.Empty(t, wrongItem.GetStatus())
		assert.True(t, wrongItem.GetCreatedAt().IsZero())
		assert.True(t, wrongItem.GetUpdatedAt().IsZero())
	})
}
