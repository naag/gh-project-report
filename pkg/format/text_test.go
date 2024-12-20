package format

import (
	"testing"

	"github.com/naag/gh-project-report/pkg/types"
	"github.com/stretchr/testify/assert"
)

func TestTextFormatter(t *testing.T) {
	diff := createTestDiff()
	formatter := NewTextFormatter()

	output := formatter.Format(diff)

	// Test key aspects of the text output
	assert.Contains(t, output, "New Task")
	assert.Contains(t, output, "Removed Task")
	assert.Contains(t, output, "Changed Task")
	assert.Contains(t, output, "Status: Added")
	assert.Contains(t, output, "Status: Removed")
	assert.Contains(t, output, "status: Todo → In Progress")
	assert.Contains(t, output, "priority: Medium → High")
	assert.Contains(t, output, string(RiskLevelModerate)) // Moderate risk emoji for 8 days delay
}

func TestTextFormatterNoChanges(t *testing.T) {
	emptyDiff := types.ProjectDiff{}
	formatter := NewTextFormatter()
	output := formatter.Format(emptyDiff)
	assert.NotContains(t, output, "Changes")
}

func TestTextFormatterCustomOptions(t *testing.T) {
	diff := createTestDiff()

	t.Run("custom risk thresholds", func(t *testing.T) {
		formatter := NewTextFormatter(
			WithModerateRiskThreshold(10),
			WithHighRiskThreshold(15),
		)
		output := formatter.Format(diff)
		assert.Contains(t, output, string(RiskLevelOnTrack)) // 8 days < 10 day threshold
	})

	t.Run("custom date format", func(t *testing.T) {
		formatter := NewTextFormatter(
			WithDateFormat("2006-01-02"),
		)
		output := formatter.Format(diff)
		assert.Contains(t, output, "2024-01-01")
		assert.Contains(t, output, "2024-01-31")
	})
}
