package main

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

func openTestStore(t *testing.T) *Store {
	t.Helper()
	s, err := OpenStore(filepath.Join(t.TempDir(), "test.db"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { s.Close() })
	return s
}

func makeUser(t *testing.T, s *Store, name string) *User {
	t.Helper()
	u := &User{Email: name + "@example.com", Username: name, EmailVerified: true}
	if err := s.CreateUser(context.Background(), u); err != nil {
		t.Fatal(err)
	}
	return u
}

func TestStoreRoundTrip(t *testing.T) {
	ctx := context.Background()
	s := openTestStore(t)
	u := makeUser(t, s, "ana")

	a, err := s.Add(ctx, u.ID, Session{Date: "2026-10-02", Minutes: 25, Kind: "pomodoro"})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := s.Add(ctx, u.ID, Session{Date: "2026-10-02", Minutes: 40, Kind: "manual", Note: "chem"}); err != nil {
		t.Fatal(err)
	}

	totals, err := s.DailyTotals(ctx, u.ID)
	if err != nil {
		t.Fatal(err)
	}
	if got := totals["2026-10-02"]; got.Minutes != 65 || got.Sessions != 2 {
		t.Errorf("totals = %+v, want 65 min / 2 sessions", got)
	}

	other := makeUser(t, s, "bea")
	if ok, _ := s.Delete(ctx, other.ID, a.ID); ok {
		t.Error("another user deleted ana's session")
	}
	if totals, _ := s.DailyTotals(ctx, other.ID); len(totals) != 0 {
		t.Errorf("bea sees ana's data: %+v", totals)
	}
	if ok, _ := s.Delete(ctx, u.ID, a.ID); !ok {
		t.Error("delete reported not found")
	}
	day, _ := s.OnDate(ctx, u.ID, "2026-10-02")
	if len(day) != 1 || day[0].Note != "chem" {
		t.Errorf("after delete: %+v", day)
	}
}

func TestStoreTreatsInjectionAsData(t *testing.T) {
	ctx := context.Background()
	s := openTestStore(t)
	u := makeUser(t, s, "ana")
	s.Add(ctx, u.ID, Session{Date: "2026-10-02", Minutes: 25, Kind: "pomodoro"})

	payload := "x' OR '1'='1"
	if ok, err := s.Delete(ctx, payload, payload); err != nil || ok {
		t.Fatalf("delete with injection payload: ok=%v err=%v", ok, err)
	}
	if day, _ := s.OnDate(ctx, payload, payload); len(day) != 0 {
		t.Errorf("injection payload matched rows: %+v", day)
	}
	if found, err := s.UserByEmail(ctx, "' OR 1=1 --"); err != nil || found != nil {
		t.Errorf("user lookup with injection payload: %+v %v", found, err)
	}
	if totals, _ := s.DailyTotals(ctx, u.ID); totals["2026-10-02"].Sessions != 1 {
		t.Error("existing row was affected")
	}
}

func TestImportLegacyJSON(t *testing.T) {
	ctx := context.Background()
	s := openTestStore(t)
	path := filepath.Join(t.TempDir(), "sessions.json")
	os.WriteFile(path, []byte(`[{"id":"abc","date":"2026-09-30","minutes":50,"kind":"pomodoro","createdAt":"2026-09-30T15:00:00Z"}]`), 0o644)

	n, err := s.ImportLegacyJSON(ctx, path)
	if err != nil || n != 1 {
		t.Fatalf("import: n=%d err=%v", n, err)
	}
	if _, err := os.Stat(path + ".imported"); err != nil {
		t.Error("legacy file should be renamed after import")
	}
	if n, _ := s.ImportLegacyJSON(ctx, path); n != 0 {
		t.Error("second import should be a no-op")
	}
	u := makeUser(t, s, "ana")
	if n, err := s.ClaimLegacy(ctx, u.ID); err != nil || n != 1 {
		t.Fatalf("claim: n=%d err=%v", n, err)
	}
	day, _ := s.OnDate(ctx, u.ID, "2026-09-30")
	if len(day) != 1 || day[0].ID != "abc" || day[0].CreatedAt.Hour() != 15 {
		t.Errorf("imported row = %+v", day)
	}
}
