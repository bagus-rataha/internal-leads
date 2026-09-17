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

func TestComparisonPeriod_WeekToDate_ShiftsBackExactlySevenDays(t *testing.T) {
	// Monday 2026-09-14 through Friday 2026-09-18 (5-day week-to-date range).
	from := time.Date(2026, 9, 14, 0, 0, 0, 0, time.UTC)
	to := time.Date(2026, 9, 18, 0, 0, 0, 0, time.UTC)

	prevFrom, prevTo := comparisonPeriod(from, to)

	assert.Equal(t, time.Date(2026, 9, 7, 0, 0, 0, 0, time.UTC), prevFrom, "should be last Monday, not an arbitrary 5-day block")
	assert.Equal(t, time.Date(2026, 9, 11, 0, 0, 0, 0, time.UTC), prevTo, "should be last Friday")
}

func TestComparisonPeriod_MonthToDate_SameDayPreviousMonth(t *testing.T) {
	// 2026-09-01 through 2026-09-18 (month-to-date, September has 30 days).
	from := time.Date(2026, 9, 1, 0, 0, 0, 0, time.UTC)
	to := time.Date(2026, 9, 18, 0, 0, 0, 0, time.UTC)

	prevFrom, prevTo := comparisonPeriod(from, to)

	assert.Equal(t, time.Date(2026, 8, 1, 0, 0, 0, 0, time.UTC), prevFrom)
	assert.Equal(t, time.Date(2026, 8, 18, 0, 0, 0, 0, time.UTC), prevTo)
}

func TestComparisonPeriod_MonthToDate_ClampsAtShorterPreviousMonth(t *testing.T) {
	// 2026-10-01 through 2026-10-31 (October has 31 days, September only has 30) -
	// the naive from.AddDate(0,-1,0) on the 31st would roll over into October 1st.
	from := time.Date(2026, 10, 1, 0, 0, 0, 0, time.UTC)
	to := time.Date(2026, 10, 31, 0, 0, 0, 0, time.UTC)

	prevFrom, prevTo := comparisonPeriod(from, to)

	assert.Equal(t, time.Date(2026, 9, 1, 0, 0, 0, 0, time.UTC), prevFrom)
	assert.Equal(t, time.Date(2026, 9, 30, 0, 0, 0, 0, time.UTC), prevTo, "must clamp to September's actual last day, not roll over into October")
}

func TestComparisonPeriod_SingleDay_StillComparesToYesterday(t *testing.T) {
	// "Hari ini" - already worked correctly before this change, must keep working.
	from := time.Date(2026, 9, 18, 0, 0, 0, 0, time.UTC)
	to := time.Date(2026, 9, 18, 0, 0, 0, 0, time.UTC)

	prevFrom, prevTo := comparisonPeriod(from, to)

	assert.Equal(t, time.Date(2026, 9, 17, 0, 0, 0, 0, time.UTC), prevFrom)
	assert.Equal(t, time.Date(2026, 9, 17, 0, 0, 0, 0, time.UTC), prevTo)
}

func TestComparisonPeriod_MonthStartsOnMonday_PrefersMonthToDateOverWeekToDate(t *testing.T) {
	// June 2026 starts on a Monday - from satisfies BOTH the week-to-date and
	// month-to-date conditions. Month-to-date must win (it's the more specific signal).
	from := time.Date(2026, 6, 1, 0, 0, 0, 0, time.UTC)
	to := time.Date(2026, 6, 3, 0, 0, 0, 0, time.UTC)

	prevFrom, prevTo := comparisonPeriod(from, to)

	assert.Equal(t, time.Date(2026, 5, 1, 0, 0, 0, 0, time.UTC), prevFrom, "must compare against May 1 (month-to-date), not a week-to-date result")
	assert.Equal(t, time.Date(2026, 5, 3, 0, 0, 0, 0, time.UTC), prevTo)
}

func TestComparisonPeriod_ArbitraryCustomRange_ImmediatelyPrecedingWindow(t *testing.T) {
	// A custom range that starts on a Wednesday (not Monday) and isn't the 1st -
	// must fall through to the original same-length-immediately-preceding rule.
	from := time.Date(2026, 9, 9, 0, 0, 0, 0, time.UTC) // Wednesday
	to := time.Date(2026, 9, 15, 0, 0, 0, 0, time.UTC)  // Tuesday, 7 days total

	prevFrom, prevTo := comparisonPeriod(from, to)

	assert.Equal(t, time.Date(2026, 9, 2, 0, 0, 0, 0, time.UTC), prevFrom, "7 days immediately before Sep 9")
	assert.Equal(t, time.Date(2026, 9, 8, 0, 0, 0, 0, time.UTC), prevTo)
}
