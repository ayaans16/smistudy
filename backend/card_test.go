package main

import (
	"encoding/xml"
	"io"
	"net/http"
	"strings"
	"testing"
)

// getCard fetches a card as an anonymous visitor (like GitHub's image proxy would).
func getCard(t *testing.T, env *testEnv, path string) (int, http.Header, string) {
	t.Helper()
	res, err := http.Get(env.srv.URL + path)
	if err != nil {
		t.Fatal(err)
	}
	defer res.Body.Close()
	body, _ := io.ReadAll(res.Body)
	return res.StatusCode, res.Header, string(body)
}

// assertValidXML fails if the SVG isn't well-formed (which also catches injected markup
// that breaks the document).
func assertValidXML(t *testing.T, svg string) {
	t.Helper()
	dec := xml.NewDecoder(strings.NewReader(svg))
	for {
		_, err := dec.Token()
		if err == io.EOF {
			return
		}
		if err != nil {
			t.Fatalf("card is not valid XML: %v\n%s", err, svg)
		}
	}
}

func TestProfileCard(t *testing.T) {
	env := newTestEnv(t)
	ana := env.client(t)
	ana.signUp("ana@example.com", "ana", "correct horse battery")
	ana.do("PATCH", "/api/me", `{"displayName":"Ana"}`)
	for _, date := range []string{"2026-10-01", "2026-10-02"} {
		ana.do("POST", "/api/sessions", `{"date":"`+date+`","minutes":150,"kind":"pomodoro"}`)
	}

	// Private: placeholder card, no stats, same as for a username that doesn't exist.
	_, _, private := getCard(t, env, "/api/users/ana/card.svg?today=2026-10-02")
	_, _, missing := getCard(t, env, "/api/users/nobody/card.svg?today=2026-10-02")
	if private != missing || strings.Contains(private, "5h") {
		t.Errorf("private card should match the not-found card and show no stats:\n%s", private)
	}
	assertValidXML(t, private)

	ana.do("PATCH", "/api/me", `{"profilePublic":true}`)
	code, header, svg := getCard(t, env, "/api/users/ana/card.svg?today=2026-10-02")
	if code != http.StatusOK || !strings.HasPrefix(header.Get("Content-Type"), "image/svg+xml") {
		t.Fatalf("card = %d %s", code, header.Get("Content-Type"))
	}
	if !strings.Contains(header.Get("Cache-Control"), "max-age=") {
		t.Errorf("card should be cacheable, got Cache-Control %q", header.Get("Cache-Control"))
	}
	for _, want := range []string{"Ana&#39;s study stats", "@ana", "5h", "2 days", "2.5h on 2026-10-02"} {
		if !strings.Contains(svg, want) {
			t.Errorf("card missing %q", want)
		}
	}
	if strings.Contains(svg, "ana@example.com") {
		t.Error("card leaks the email address")
	}
	assertValidXML(t, svg)

	// 17 weeks of cells, ending today.
	if n := strings.Count(svg, "<rect x=") - 1; n < 16*7 || n > 17*7 {
		t.Errorf("heatmap has %d cells", n)
	}

	_, _, dark := getCard(t, env, "/api/users/ana/card.svg?today=2026-10-02&theme=dark")
	if !strings.Contains(dark, `filter id="glow"`) || !strings.Contains(dark, "#141a14") {
		t.Error("dark theme not applied")
	}
}

func TestProfileCardEscapesDisplayName(t *testing.T) {
	env := newTestEnv(t)
	c := env.client(t)
	c.signUp("eve@example.com", "eve", "correct horse battery")
	c.do("PATCH", "/api/me", `{"displayName":"<script>alert(1)</script>&\"","profilePublic":true}`)

	_, _, svg := getCard(t, env, "/api/users/eve/card.svg")
	if strings.Contains(svg, "<script") {
		t.Fatalf("display name was injected into the SVG:\n%s", svg)
	}
	if !strings.Contains(svg, "&lt;script&gt;") {
		t.Error("expected the escaped display name in the card")
	}
	assertValidXML(t, svg)
}

func TestFormatCardHours(t *testing.T) {
	cases := map[int]string{0: "0h", 45: "45m", 60: "1h", 150: "2.5h", 6000: "100h", 6030: "100h"}
	for minutes, want := range cases {
		if got := formatCardHours(minutes); got != want {
			t.Errorf("formatCardHours(%d) = %q, want %q", minutes, got, want)
		}
	}
	if got := truncateRunes("abcdefghij", 5); got != "abcd…" {
		t.Errorf("truncateRunes = %q", got)
	}
}
