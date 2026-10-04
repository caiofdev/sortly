package criteria

import (
	"context"
	"testing"
)

func TestDurationRule(t *testing.T) {
	cases := []struct {
		name string
		meta fakeMeta
		want string
	}{
		{"ok", fakeMeta{seconds: 61}, "duration-00h01m01s"},
		{"falha de leitura", fakeMeta{durErr: errRead}, "duration-unknown"},
		{"duração 0 sem erro", fakeMeta{seconds: 0}, "duration-unknown"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, _ := durationRule{meta: tc.meta}.Segment(context.Background(), File{})
			if got != tc.want {
				t.Fatalf("Segment = %q, want %q", got, tc.want)
			}
		})
	}
}

func TestFormatDuration(t *testing.T) {
	cases := []struct {
		seconds float64
		want    string
	}{
		{-1, "00h00m00s"},
		{0.4, "00h00m00s"},
		{0.5, "00h00m01s"},
		{59, "00h00m59s"},
		{59.5, "00h01m00s"},
		{60, "00h01m00s"},
		{3599, "00h59m59s"},
		{3600, "01h00m00s"},
		{59.98, "00h01m00s"},
		{360000, "100h00m00s"},
	}
	for _, tc := range cases {
		if got := formatDuration(tc.seconds); got != tc.want {
			t.Errorf("formatDuration(%v) = %q, want %q", tc.seconds, got, tc.want)
		}
	}
}
