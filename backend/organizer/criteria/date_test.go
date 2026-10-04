package criteria

import (
	"context"
	"testing"
	"time"
)

func TestDateRule(t *testing.T) {
	loc := time.FixedZone("BRT", -3*3600)
	cases := []struct {
		name string
		at   time.Time
		want string
	}{
		{"23:59:59 local", time.Date(2026, 3, 5, 23, 59, 59, 0, loc), "date-2026-03-05"},
		{"00:00:00 local", time.Date(2026, 3, 6, 0, 0, 0, 0, loc), "date-2026-03-06"},
		{"virada de ano", time.Date(2025, 12, 31, 23, 59, 59, 0, loc), "date-2025-12-31"},
		{"29 de fevereiro", time.Date(2024, 2, 29, 12, 0, 0, 0, loc), "date-2024-02-29"},
		{"UTC já é o dia seguinte, local não", time.Date(2026, 3, 6, 1, 30, 0, 0, time.UTC), "date-2026-03-05"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, _ := dateRule{loc: loc}.Segment(context.Background(), File{ModTime: tc.at})
			if got != tc.want {
				t.Fatalf("Segment = %q, want %q", got, tc.want)
			}
		})
	}
}
