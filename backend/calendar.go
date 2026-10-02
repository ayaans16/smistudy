package main

import (
	"sort"
	"strconv"
	"time"
)

// Level thresholds in minutes, mirroring GitHub's five-step scale:
// 0 = nothing, 1 = under 1h, 2 = 1–2h, 3 = 2–4h, 4 = 4h+.
var levelThresholds = []int{1, 60, 120, 240}

func levelFor(minutes int) int {
	level := 0
	for i, t := range levelThresholds {
		if minutes >= t {
			level = i + 1
		}
	}
	return level
}

type Day struct {
	Date     string `json:"date"`
	Minutes  int    `json:"minutes"`
	Sessions int    `json:"sessions"`
	Level    int    `json:"level"`
	InRange  bool   `json:"inRange"` // false for padding cells outside the selected range
}

type MonthLabel struct {
	Name string `json:"name"`
	Week int    `json:"week"` // column index where the label starts
}

type Calendar struct {
	From         string       `json:"from"`
	To           string       `json:"to"`
	TotalMinutes int          `json:"totalMinutes"`
	ActiveDays   int          `json:"activeDays"`
	Weeks        [][]Day      `json:"weeks"` // columns of 7 days, Sunday first
	Months       []MonthLabel `json:"months"`
	Thresholds   []int        `json:"thresholds"`
}

// RangeFor resolves a filter into a date range.
// filter "last" (default) is the rolling year ending today, like GitHub's default view;
// a 4-digit year selects Jan 1 – Dec 31 of that year (capped at today for the current year).
func RangeFor(filter string, today time.Time) (time.Time, time.Time, error) {
	if filter == "" || filter == "last" {
		return today.AddDate(-1, 0, 1), today, nil
	}
	year, err := strconv.Atoi(filter)
	if err != nil || year < 1970 || year > 9999 {
		return time.Time{}, time.Time{}, errBadFilter
	}
	from := time.Date(year, 1, 1, 0, 0, 0, 0, time.UTC)
	to := time.Date(year, 12, 31, 0, 0, 0, 0, time.UTC)
	if to.After(today) {
		to = today
	}
	return from, to, nil
}

// BuildCalendar lays out the range as a GitHub-style grid of week columns.
func BuildCalendar(totals map[string]DayTotal, from, to time.Time) Calendar {
	cal := Calendar{From: from.Format(dateLayout), To: to.Format(dateLayout), Thresholds: levelThresholds}

	// Pad back to Sunday and forward to Saturday so every column has 7 cells.
	start := from.AddDate(0, 0, -int(from.Weekday()))
	end := to.AddDate(0, 0, 6-int(to.Weekday()))

	var week []Day
	lastMonth := -1
	for d := start; !d.After(end); d = d.AddDate(0, 0, 1) {
		key := d.Format(dateLayout)
		in := !d.Before(from) && !d.After(to)
		day := Day{Date: key, InRange: in}
		if in {
			t := totals[key]
			day.Minutes, day.Sessions, day.Level = t.Minutes, t.Sessions, levelFor(t.Minutes)
			cal.TotalMinutes += t.Minutes
			if t.Minutes > 0 {
				cal.ActiveDays++
			}
			// Label a month on the first column that contains an in-range day of it.
			if m := int(d.Month()); m != lastMonth && (d.Day() <= 7 || len(cal.Months) == 0) {
				cal.Months = append(cal.Months, MonthLabel{Name: d.Month().String()[:3], Week: len(cal.Weeks)})
				lastMonth = m
			}
		}
		week = append(week, day)
		if len(week) == 7 {
			cal.Weeks = append(cal.Weeks, week)
			week = nil
		}
	}
	// Drop a leading label that would collide with the next one (GitHub does the same).
	if len(cal.Months) > 1 && cal.Months[1].Week-cal.Months[0].Week < 2 {
		cal.Months = cal.Months[1:]
	}
	return cal
}

type Stats struct {
	TodayMinutes   int     `json:"todayMinutes"`
	WeekMinutes    int     `json:"weekMinutes"` // since Sunday
	TotalMinutes   int     `json:"totalMinutes"`
	CurrentStreak  int     `json:"currentStreak"`
	LongestStreak  int     `json:"longestStreak"`
	BestDay        string  `json:"bestDay,omitempty"`
	BestDayMinutes int     `json:"bestDayMinutes"`
	DailyAverage   float64 `json:"dailyAverage"` // minutes per active day
}

func BuildStats(totals map[string]DayTotal, today time.Time) Stats {
	var st Stats
	todayKey := today.Format(dateLayout)
	weekStart := today.AddDate(0, 0, -int(today.Weekday()))

	dates := make([]string, 0, len(totals))
	for date, t := range totals {
		if t.Minutes <= 0 {
			continue
		}
		dates = append(dates, date)
		st.TotalMinutes += t.Minutes
		if t.Minutes > st.BestDayMinutes || (t.Minutes == st.BestDayMinutes && date > st.BestDay) {
			st.BestDay, st.BestDayMinutes = date, t.Minutes
		}
		if d, err := time.Parse(dateLayout, date); err == nil && !d.Before(weekStart) && !d.After(today) {
			st.WeekMinutes += t.Minutes
		}
	}
	st.TodayMinutes = totals[todayKey].Minutes
	if len(dates) > 0 {
		st.DailyAverage = float64(st.TotalMinutes) / float64(len(dates))
	}

	// Longest run of consecutive study days.
	sort.Strings(dates)
	run := 0
	var prev time.Time
	for i, ds := range dates {
		d, err := time.Parse(dateLayout, ds)
		if err != nil {
			continue
		}
		if i > 0 && d.Sub(prev) == 24*time.Hour {
			run++
		} else {
			run = 1
		}
		st.LongestStreak = max(st.LongestStreak, run)
		prev = d
	}

	// Current streak counts back from today; an empty today doesn't break it yet.
	d := today
	if totals[todayKey].Minutes == 0 {
		d = d.AddDate(0, 0, -1)
	}
	for totals[d.Format(dateLayout)].Minutes > 0 {
		st.CurrentStreak++
		d = d.AddDate(0, 0, -1)
	}
	return st
}

// Years returns every year with study time plus the current year, newest first.
func Years(totals map[string]DayTotal, today time.Time) []int {
	seen := map[int]bool{today.Year(): true}
	for date, t := range totals {
		if t.Minutes <= 0 || len(date) < 4 {
			continue
		}
		if y, err := strconv.Atoi(date[:4]); err == nil {
			seen[y] = true
		}
	}
	years := make([]int, 0, len(seen))
	for y := range seen {
		years = append(years, y)
	}
	sort.Sort(sort.Reverse(sort.IntSlice(years)))
	return years
}
