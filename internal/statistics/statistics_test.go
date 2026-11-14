package statistics

import (
	"math"
	"testing"
	"time"

	"DayOneCLI/internal/journal"
)

func journalsOn(dates ...string) []journal.Journal {
	var entries []journal.Journal
	for _, date := range dates {
		entries = append(entries, journal.Journal{Date: date})
	}
	return entries
}

func TestCalculateStreak(t *testing.T) {
	now := time.Date(2026, 1, 2, 0, 30, 0, 0, time.FixedZone("KST", 9*60*60))
	for _, tt := range []struct {
		name  string
		dates []string
		want  int
	}{
		{"empty", nil, 0},
		{"today", []string{"2026-01-02"}, 1},
		{"yesterday and year boundary", []string{"2025-12-31", "2026-01-01"}, 2},
		{"gap", []string{"2026-01-02", "2025-12-31"}, 1},
		{"broken streak", []string{"2025-12-31"}, 0},
		{"unordered duplicates and future", []string{"2026-01-03", "2026-01-01", "2026-01-02", "2026-01-01"}, 2},
	} {
		t.Run(tt.name, func(t *testing.T) {
			if got := calculateStreak(journalsOn(tt.dates...), now); got != tt.want {
				t.Fatalf("streak = %d, want %d", got, tt.want)
			}
		})
	}
}

func TestCalculateLongestStreak(t *testing.T) {
	for _, tt := range []struct {
		name  string
		dates []string
		want  int
	}{
		{"empty", nil, 0},
		{"invalid dates", []string{"invalid", "2026-02-30"}, 0},
		{"leap day and duplicates", []string{"2024-03-01", "2024-02-28", "2024-02-29", "2024-02-29"}, 3},
		{"later streak beyond two years", []string{"2020-01-01", "2026-11-14", "2026-11-12", "2026-11-13"}, 3},
	} {
		t.Run(tt.name, func(t *testing.T) {
			if got := calculateLongestStreak(journalsOn(tt.dates...)); got != tt.want {
				t.Fatalf("longest streak = %d, want %d", got, tt.want)
			}
		})
	}
	t.Run("continuous for more than two years", func(t *testing.T) {
		var entries []journal.Journal
		start := time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)
		for i := 0; i < 1100; i++ {
			entries = append(entries, journal.Journal{Date: start.AddDate(0, 0, i).Format("2006-01-02")})
		}
		if got := calculateLongestStreak(entries); got != len(entries) {
			t.Fatalf("longest streak = %d, want %d", got, len(entries))
		}
	})
}

func TestFormatStars(t *testing.T) {
	for _, tt := range []struct {
		rating float64
		want   string
	}{
		{0, "☆☆☆☆☆"}, {1, "★☆☆☆☆"}, {3.9, "★★★☆☆"}, {5, "★★★★★"},
		{-1, "☆☆☆☆☆"}, {10, "★★★★★"}, {math.NaN(), "☆☆☆☆☆"},
		{math.Inf(1), "★★★★★"}, {math.Inf(-1), "☆☆☆☆☆"},
	} {
		if got := FormatStars(tt.rating); got != tt.want {
			t.Errorf("FormatStars(%v) = %q, want %q", tt.rating, got, tt.want)
		}
	}
}
