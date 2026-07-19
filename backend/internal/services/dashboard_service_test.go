package services

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestChangePct_BaselineZero_NilRegardlessOfCurrent(t *testing.T) {
	assert.Nil(t, changePct(0, 0), "both zero -> nil, frontend shows —")
	assert.Nil(t, changePct(5, 0), "prev zero, cur nonzero -> nil, frontend shows \"baru\"")
}

func TestChangePct_NonZeroBaseline_RoundedSignedPercent(t *testing.T) {
	pct := changePct(15, 10)
	if assert.NotNil(t, pct) {
		assert.Equal(t, 50, *pct, "(15-10)/10 = 50%")
	}
	down := changePct(5, 10)
	if assert.NotNil(t, down) {
		assert.Equal(t, -50, *down, "decrease is a negative percent, not an absolute value")
	}
}

func TestPercentOf_ZeroTotal_Nil(t *testing.T) {
	assert.Nil(t, percentOf(0, 0))
	assert.Nil(t, percentOf(3, 0), "a nonzero part with a zero total is still undefined, not a divide-by-zero 0")
}

func TestPercentOf_RoundsToNearest(t *testing.T) {
	pct := percentOf(1, 3)
	if assert.NotNil(t, pct) {
		assert.Equal(t, 33, *pct)
	}
}

func TestComparisonPeriod_SameLengthImmediatelyPreceding(t *testing.T) {
	from := time.Date(2026, 7, 1, 0, 0, 0, 0, time.UTC)
	to := time.Date(2026, 7, 30, 0, 0, 0, 0, time.UTC) // 30-day window
	prevFrom, prevTo := comparisonPeriod(from, to)
	assert.Equal(t, time.Date(2026, 6, 30, 0, 0, 0, 0, time.UTC), prevTo, "prev period ends the day before `from`")
	assert.Equal(t, time.Date(2026, 6, 1, 0, 0, 0, 0, time.UTC), prevFrom, "prev period is the same length (30 days)")
}
