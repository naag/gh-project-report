package format

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDefaultOptions(t *testing.T) {
	opts := DefaultOptions()

	assert.Equal(t, "Jan 2, 2006", opts.DateFormat)
	assert.Equal(t, 7, opts.ModerateRiskThreshold)
	assert.Equal(t, 14, opts.HighRiskThreshold)
	assert.Equal(t, 30, opts.ExtremeRiskThreshold)
}

func TestOptionFunctions(t *testing.T) {
	t.Run("WithDateFormat", func(t *testing.T) {
		opts := DefaultOptions()
		WithDateFormat("2006-01-02")(&opts)
		assert.Equal(t, "2006-01-02", opts.DateFormat)
	})

	t.Run("WithModerateRiskThreshold", func(t *testing.T) {
		opts := DefaultOptions()
		WithModerateRiskThreshold(10)(&opts)
		assert.Equal(t, 10, opts.ModerateRiskThreshold)
	})

	t.Run("WithHighRiskThreshold", func(t *testing.T) {
		opts := DefaultOptions()
		WithHighRiskThreshold(21)(&opts)
		assert.Equal(t, 21, opts.HighRiskThreshold)
	})

	t.Run("WithExtremeRiskThreshold", func(t *testing.T) {
		opts := DefaultOptions()
		WithExtremeRiskThreshold(45)(&opts)
		assert.Equal(t, 45, opts.ExtremeRiskThreshold)
	})

	t.Run("chaining options", func(t *testing.T) {
		opts := DefaultOptions()
		WithModerateRiskThreshold(10)(&opts)
		WithHighRiskThreshold(21)(&opts)
		WithExtremeRiskThreshold(45)(&opts)
		WithDateFormat("2006-01-02")(&opts)

		assert.Equal(t, 10, opts.ModerateRiskThreshold)
		assert.Equal(t, 21, opts.HighRiskThreshold)
		assert.Equal(t, 45, opts.ExtremeRiskThreshold)
		assert.Equal(t, "2006-01-02", opts.DateFormat)
	})
}

func TestRiskLevelConstants(t *testing.T) {
	assert.Equal(t, RiskLevel("🔵 On track"), RiskLevelOnTrack)
	assert.Equal(t, RiskLevel("🚀 Ahead of schedule"), RiskLevelAhead)
	assert.Equal(t, RiskLevel("🟠 Moderate risk"), RiskLevelModerate)
	assert.Equal(t, RiskLevel("🔴 High risk"), RiskLevelHigh)
	assert.Equal(t, RiskLevel("🚫 Extreme risk"), RiskLevelExtreme)
}
