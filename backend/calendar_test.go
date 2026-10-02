package main

import (
	"testing"
	"time"
)

func day(s string) time.Time {
	t, _ := time.Parse(dateLayout, s)
	return t
}

func TestLevelFor(t *testing.T) {
	cases := map[int]int{0: 0, 1: 1, 59: 1, 60: 2, 119: 2, 120: 3, 239: 3, 240: 4, 600: 4}
	for minutes, want := range cases {
		if got := levelFor(minutes); got != want {
			t.Errorf("levelFor(%d) = %d, want %d", minutes, got, want)
		}
	}
}

func TestBuildCalendarGrid(t *testing.T) {
	totals := map[string]DayTotal{
		"2026-03-04": {Minutes: 90, Sessions: 3},
		"2025-12-31": {Minutes: 500, Sessions: 1}, // outside the 2026 range
	}
	from, to, err := RangeFor("2026", day("2026-10-02"))
	if err != nil {
		t.Fatal(err)
	}
	cal := BuildCalendar(totals, from, to)
	if cal.To != "2026-10-02" {
		t.Errorf("current year should be capped at today, got %s", cal.To)
	}
	if cal.TotalMinutes != 90 || cal.ActiveDays != 1 {
		t.Errorf("totals = %d min / %d days, want 90 / 1", cal.TotalMinutes, cal.ActiveDays)
	}
	for _, w := range cal.Weeks {
		if len(w) != 7 {
			t.Fatalf("week has %d days", len(w))
		}
	}
	first := cal.Weeks[0][0]
	if day(first.Date).Weekday() != time.Sunday {
		t.Errorf("grid should start on Sunday, got %s", first.Date)
	}
	if first.InRange || first.Minutes != 0 {
		t.Errorf("padding day 2025-12-28 should be out of range and empty: %+v", first)
	}
	if cal.Months[0].Name != "Jan" || cal.Months[len(cal.Months)-1].Name != "Oct" {
		t.Errorf("unexpected month labels %+v", cal.Months)
	}
}

func TestRollingYear(t *testing.T) {
	from, to, _ := RangeFor("last", day("2026-10-02"))
	if from.Format(dateLayout) != "2025-10-03" || to.Format(dateLayout) != "2026-10-02" {
		t.Errorf("rolling range = %s..%s", from.Format(dateLayout), to.Format(dateLayout))
	}
	if _, _, err := RangeFor("abc", day("2026-10-02")); err == nil {
		t.Error("expected error for bad filter")
	}
}

func TestStreaks(t *testing.T) {
	totals := map[string]DayTotal{
		"2026-09-01": {Minutes: 30}, "2026-09-02": {Minutes: 30}, "2026-09-03": {Minutes: 30}, "2026-09-04": {Minutes: 30},
		"2026-09-30": {Minutes: 25}, "2026-10-01": {Minutes: 50},
	}
	st := BuildStats(totals, day("2026-10-02"))
	if st.CurrentStreak != 2 {
		t.Errorf("current streak = %d, want 2 (today empty shouldn't break it)", st.CurrentStreak)
	}
	if st.LongestStreak != 4 {
		t.Errorf("longest streak = %d, want 4", st.LongestStreak)
	}
	if st.WeekMinutes != 75 || st.BestDayMinutes != 50 {
		t.Errorf("week=%d best=%d", st.WeekMinutes, st.BestDayMinutes)
	}
}
