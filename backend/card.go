package main

import (
	"fmt"
	"html"
	"net/http"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"
)

// Profile cards are embeddable SVG images (e.g. in a GitHub profile README):
//
//	[![smistudy](https://smistudy.ca/api/users/ayaan/card.svg)](https://smistudy.ca/u/ayaan)
//
// They show the same information as the public profile page, so they only render for
// public profiles. Private and unknown users get an identical placeholder card, so the
// endpoint can't be used to discover which usernames exist.

const (
	cardWidth  = 495
	cardHeight = 195
	cardWeeks  = 17 // heatmap columns
	cardCell   = 10
	cardGap    = 3
	cardCache  = 30 * time.Minute
)

type cardTheme struct {
	bg, border, text, muted, accent string
	heat                            [5]string
	glow                            bool
	smiBody, smiShade               string
}

var cardThemes = map[string]cardTheme{
	"light": {
		bg: "#ffffff", border: "#e4e8d9", text: "#28311f", muted: "#6d7762", accent: "#5b961f",
		heat:    [5]string{"#ebede4", "#d5ebb0", "#aad873", "#7cbb3d", "#4d881b"},
		smiBody: "#d4eaae", smiShade: "#8fbf55",
	},
	"dark": {
		bg: "#141a14", border: "#232d22", text: "#e5eedb", muted: "#8f9c84", accent: "#b4f06a",
		heat:    [5]string{"#1a2119", "#25431b", "#3b7225", "#69b23a", "#b4f06a"},
		glow:    true,
		smiBody: "#c8f58f", smiShade: "#8fd15a",
	},
}

type cardData struct {
	Title, Subtitle string
	Stats           [4][2]string // label, value
	Weeks           [][]Day      // nil for the placeholder card
}

func (a *App) handleProfileCard(w http.ResponseWriter, r *http.Request) {
	theme, ok := cardThemes[r.URL.Query().Get("theme")]
	if !ok {
		theme = cardThemes["light"]
	}
	today, ok := todayParam(w, r)
	if !ok {
		return
	}

	var u *User
	if name, err := normalizeUsername(r.PathValue("username")); err == nil {
		if u, err = a.store.UserByUsername(r.Context(), name); err != nil {
			serverError(w, err)
			return
		}
	}

	data := cardData{
		Title:    "smistudy",
		Subtitle: "This profile is private or doesn't exist",
	}
	if u != nil && u.ProfilePublic {
		totals, ok := a.totals(w, r, u.ID)
		if !ok {
			return
		}
		data = profileCardData(u, totals, today)
	}

	// Override the API's default no-store: image proxies (like GitHub's) should cache cards.
	w.Header().Set("Content-Type", "image/svg+xml; charset=utf-8")
	w.Header().Set("Cache-Control", "public, max-age="+strconv.Itoa(int(cardCache.Seconds())))
	w.WriteHeader(http.StatusOK)
	fmt.Fprint(w, renderCard(data, theme))
}

func profileCardData(u *User, totals map[string]DayTotal, today time.Time) cardData {
	stats := BuildStats(totals, today)
	name := u.DisplayName
	if name == "" {
		name = u.Username
	}
	// The rolling window ends today and starts on the Sunday cardWeeks-1 weeks earlier.
	from := today.AddDate(0, 0, -int(today.Weekday())-(cardWeeks-1)*7)
	cal := BuildCalendar(totals, from, today)
	return cardData{
		Title:    truncateRunes(name, 22) + "'s study stats",
		Subtitle: "@" + u.Username + " · smistudy.ca",
		Stats: [4][2]string{
			{"Total", formatCardHours(stats.TotalMinutes)},
			{"This week", formatCardHours(stats.WeekMinutes)},
			{"Streak", pluralDays(stats.CurrentStreak)},
			{"Best streak", pluralDays(stats.LongestStreak)},
		},
		Weeks: cal.Weeks,
	}
}

// renderCard builds the SVG. Every piece of user-controlled text goes through esc,
// since SVG is markup and an unescaped display name could inject elements.
func renderCard(d cardData, t cardTheme) string {
	var b strings.Builder
	esc := html.EscapeString
	font := `font-family="'Segoe UI', Ubuntu, 'Helvetica Neue', Arial, sans-serif"`

	fmt.Fprintf(&b, `<svg xmlns="http://www.w3.org/2000/svg" width="%d" height="%d" viewBox="0 0 %d %d" role="img" aria-label="%s">`,
		cardWidth, cardHeight, cardWidth, cardHeight, esc(d.Title))
	fmt.Fprintf(&b, `<title>%s</title>`, esc(d.Title))
	if t.glow {
		b.WriteString(`<defs><filter id="glow" x="-50%" y="-50%" width="200%" height="200%">` +
			`<feGaussianBlur stdDeviation="1.6" result="b"/><feMerge><feMergeNode in="b"/><feMergeNode in="SourceGraphic"/></feMerge>` +
			`</filter></defs>`)
	}
	fmt.Fprintf(&b, `<rect x="0.5" y="0.5" width="%d" height="%d" rx="12" fill="%s" stroke="%s"/>`,
		cardWidth-1, cardHeight-1, t.bg, t.border)

	writeMascot(&b, t, 20, 16)
	fmt.Fprintf(&b, `<text x="78" y="40" %s font-size="18" font-weight="700" fill="%s">%s</text>`, font, t.text, esc(d.Title))
	fmt.Fprintf(&b, `<text x="78" y="60" %s font-size="12" fill="%s">%s</text>`, font, t.muted, esc(d.Subtitle))

	if d.Weeks == nil {
		return b.String() + `</svg>`
	}

	// 2×2 grid of stats on the left.
	for i, s := range d.Stats {
		x := 24 + (i%2)*110
		y := 104 + (i/2)*48
		fmt.Fprintf(&b, `<text x="%d" y="%d" %s font-size="11" fill="%s">%s</text>`, x, y, font, t.muted, esc(s[0]))
		fmt.Fprintf(&b, `<text x="%d" y="%d" %s font-size="18" font-weight="700" fill="%s">%s</text>`, x, y+21, font, t.accent, esc(s[1]))
	}

	// Mini heatmap on the right: the most recent cardWeeks columns.
	weeks := d.Weeks
	if len(weeks) > cardWeeks {
		weeks = weeks[len(weeks)-cardWeeks:]
	}
	left := cardWidth - 24 - cardWeeks*(cardCell+cardGap) + cardGap
	for wi, week := range weeks {
		for di, day := range week {
			if !day.InRange {
				continue
			}
			x := left + wi*(cardCell+cardGap)
			y := 84 + di*(cardCell+cardGap)
			glow := ""
			if t.glow && day.Level == 4 {
				glow = ` filter="url(#glow)"`
			}
			fmt.Fprintf(&b, `<rect x="%d" y="%d" width="%d" height="%d" rx="2" fill="%s"%s><title>%s</title></rect>`,
				x, y, cardCell, cardCell, t.heat[day.Level], glow, esc(cardDayTitle(day)))
		}
	}
	return b.String() + `</svg>`
}

// writeMascot draws the reading smiski (the site's logo), scaled down, at (x, y).
func writeMascot(b *strings.Builder, t cardTheme, x, y int) {
	fmt.Fprintf(b, `<g transform="translate(%d %d) scale(0.38)"`, x, y)
	if t.glow {
		b.WriteString(` filter="url(#glow)"`)
	}
	body := fmt.Sprintf(`fill="%s" stroke="%s" stroke-width="3"`, t.smiBody, t.smiShade)
	fmt.Fprintf(b, `><path %s d="M30 78 C26 104 34 122 60 122 C86 122 94 104 90 78 Z"/>`, body)
	fmt.Fprintf(b, `<path %s d="M60 14 C86 14 98 30 98 52 C98 74 82 86 60 86 C38 86 22 74 22 52 C22 30 34 14 60 14 Z"/>`, body)
	b.WriteString(`<circle cx="47" cy="54" r="4" fill="#2b3324"/><circle cx="73" cy="54" r="4" fill="#2b3324"/>`)
	b.WriteString(`<path d="M34 92 L60 98 L60 120 L34 114 Z" fill="#fff7ea" stroke="#d9b98f" stroke-width="2"/>`)
	b.WriteString(`<path d="M86 92 L60 98 L60 120 L86 114 Z" fill="#fff2df" stroke="#d9b98f" stroke-width="2"/>`)
	fmt.Fprintf(b, `<ellipse %s cx="33" cy="102" rx="6" ry="7"/><ellipse %s cx="87" cy="102" rx="6" ry="7"/></g>`, body, body)
}

func cardDayTitle(d Day) string {
	if d.Minutes == 0 {
		return "No study time on " + d.Date
	}
	return formatCardHours(d.Minutes) + " on " + d.Date
}

// formatCardHours: 0 → "0h", 45 → "45m", 150 → "2.5h", 6000 → "100h".
func formatCardHours(minutes int) string {
	switch {
	case minutes == 0:
		return "0h"
	case minutes < 60:
		return strconv.Itoa(minutes) + "m"
	case minutes >= 100*60:
		return strconv.Itoa(minutes/60) + "h"
	}
	return strings.TrimSuffix(strconv.FormatFloat(float64(minutes)/60, 'f', 1, 64), ".0") + "h"
}

func pluralDays(n int) string {
	if n == 1 {
		return "1 day"
	}
	return strconv.Itoa(n) + " days"
}

func truncateRunes(s string, n int) string {
	if utf8.RuneCountInString(s) <= n {
		return s
	}
	return string([]rune(s)[:n-1]) + "…"
}
