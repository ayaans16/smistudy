package main

import "sort"

// goalProgress adds up study time from startDate onwards and reports the day the running
// total first reached target ("" if it hasn't yet).
func goalProgress(totals map[string]DayTotal, startDate string, target int) (minutes int, reachedOn string) {
	dates := make([]string, 0, len(totals))
	for d := range totals {
		if d >= startDate { // YYYY-MM-DD strings sort chronologically
			dates = append(dates, d)
		}
	}
	sort.Strings(dates)
	for _, d := range dates {
		minutes += totals[d].Minutes
		if reachedOn == "" && minutes >= target {
			reachedOn = d
		}
	}
	return minutes, reachedOn
}
