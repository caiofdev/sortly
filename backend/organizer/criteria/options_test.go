package criteria

import "testing"

func TestNormalize(t *testing.T) {
	yes, no := true, false
	cases := []struct {
		name string
		raw  RawOptions
		want Options
	}{
		{"byExtension ausente vale true", RawOptions{ByDate: true}, Options{ByDate: true, ByExtension: true}},
		{"byExtension true", RawOptions{ByExtension: &yes}, Options{ByExtension: true}},
		{"byExtension false", RawOptions{BySize: true, ByExtension: &no}, Options{BySize: true}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := tc.raw.Normalize(); got != tc.want {
				t.Fatalf("Normalize = %+v, want %+v", got, tc.want)
			}
		})
	}
}

func TestAny(t *testing.T) {
	all := Options{ByDuration: true, ByPages: true, ByResolution: true, ByDate: true, BySize: true, ByExtension: true}
	cases := []struct {
		name string
		opts Options
		want bool
	}{
		{"0 critérios", Options{}, false},
		{"1 critério", Options{ByPages: true}, true},
		{"6 critérios", all, true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := tc.opts.Any(); got != tc.want {
				t.Fatalf("Any = %v, want %v", got, tc.want)
			}
		})
	}
}
