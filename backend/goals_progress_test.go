package main

import "testing"

func TestGoalProgress(t *testing.T) {
	totals := map[string]DayTotal{
		"2026-09-30": {Minutes: 300}, // before the goal started: doesn't count
		"2026-10-01": {Minutes: 60},
		"2026-10-02": {Minutes: 90},
		"2026-10-04": {Minutes: 120},
	}
	cases := []struct {
		start   string
		target  int
		minutes int
		reached string
	}{
		{"2026-10-01", 120, 270, "2026-10-02"}, // 60 + 90 crosses 120 on the 2nd
		{"2026-10-01", 270, 270, "2026-10-04"}, // exactly hitting the target counts
		{"2026-10-01", 600, 270, ""},           // not there yet
		{"2026-10-03", 60, 120, "2026-10-04"},  // a start date between sessions
		{"2026-11-01", 60, 0, ""},              // nothing logged since
	}
	for _, c := range cases {
		minutes, reached := goalProgress(totals, c.start, c.target)
		if minutes != c.minutes || reached != c.reached {
			t.Errorf("goalProgress(from %s, target %d) = %d, %q; want %d, %q",
				c.start, c.target, minutes, reached, c.minutes, c.reached)
		}
	}
}
