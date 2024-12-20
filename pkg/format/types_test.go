package format

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDefaultOptions(t *testing.T) {
	opts := DefaultOptions()

	assert.Equal(t, "Jan 2, 2006", opts.DateFormat)
	assert.Equal(t, 7, opts.ModerateDelayThreshold)
	assert.Equal(t, 14, opts.HighDelayThreshold)
	assert.Equal(t, 30, opts.ExtremeDelayThreshold)
}

func TestOptionFunctions(t *testing.T) {
	t.Run("WithDateFormat", func(t *testing.T) {
		opts := DefaultOptions()
		WithDateFormat("2006-01-02")(&opts)
		assert.Equal(t, "2006-01-02", opts.DateFormat)
	})

	t.Run("WithModerateDelayThreshold", func(t *testing.T) {
		opts := DefaultOptions()
		WithModerateDelayThreshold(10)(&opts)
		assert.Equal(t, 10, opts.ModerateDelayThreshold)
	})

	t.Run("WithHighDelayThreshold", func(t *testing.T) {
		opts := DefaultOptions()
		WithHighDelayThreshold(21)(&opts)
		assert.Equal(t, 21, opts.HighDelayThreshold)
	})

	t.Run("WithExtremeDelayThreshold", func(t *testing.T) {
		opts := DefaultOptions()
		WithExtremeDelayThreshold(45)(&opts)
		assert.Equal(t, 45, opts.ExtremeDelayThreshold)
	})

	t.Run("chaining options", func(t *testing.T) {
		opts := DefaultOptions()
		WithModerateDelayThreshold(10)(&opts)
		WithHighDelayThreshold(21)(&opts)
		WithExtremeDelayThreshold(45)(&opts)
		WithDateFormat("2006-01-02")(&opts)

		assert.Equal(t, 10, opts.ModerateDelayThreshold)
		assert.Equal(t, 21, opts.HighDelayThreshold)
		assert.Equal(t, 45, opts.ExtremeDelayThreshold)
		assert.Equal(t, "2006-01-02", opts.DateFormat)
	})
}

func TestDelayLevelConstants(t *testing.T) {
	assert.Equal(t, DelayLevel("🔵 On track"), DelayLevelOnTrack)
	assert.Equal(t, DelayLevel("🚀 Ahead of schedule"), DelayLevelAhead)
	assert.Equal(t, DelayLevel("🟠 Moderate delay"), DelayLevelModerate)
	assert.Equal(t, DelayLevel("🔴 High delay"), DelayLevelHigh)
	assert.Equal(t, DelayLevel("🚫 Extreme delay"), DelayLevelExtreme)
}
