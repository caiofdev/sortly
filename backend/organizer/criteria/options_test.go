package criteria

import "testing"

var allOptions = Options{ByDuration: true, ByPages: true, ByResolution: true, ByDate: true, BySize: true, ByExtension: true}

func TestAny(t *testing.T) {
	cases := []struct {
		name string
		opts Options
		want bool
	}{
		{"0 critérios", Options{}, false},
		{"1 critério", Options{ByPages: true}, true},
		{"6 critérios", allOptions, true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := tc.opts.Any(); got != tc.want {
				t.Fatalf("Any = %v, want %v", got, tc.want)
			}
		})
	}
}

func TestCount(t *testing.T) {
	cases := []struct {
		opts Options
		want int
	}{
		{Options{}, 0},
		{Default, 1},
		{Options{ByDate: true, BySize: true}, 2},
		{allOptions, 6},
	}
	for _, tc := range cases {
		if got := tc.opts.Count(); got != tc.want {
			t.Errorf("Count(%+v) = %d, want %d", tc.opts, got, tc.want)
		}
	}
}

func TestGetAndWith(t *testing.T) {
	for _, key := range Keys {
		t.Run(key, func(t *testing.T) {
			on, ok := Options{}.Get(key)
			if !ok || on {
				t.Fatalf("Get em Options{} = (%v, %v), want (false, true)", on, ok)
			}
			changed, ok := Options{}.With(key, true)
			if !ok || changed.Count() != 1 {
				t.Fatalf("With(%q, true) = (%+v, %v)", key, changed, ok)
			}
			if on, _ := changed.Get(key); !on {
				t.Fatalf("Get depois de With = false")
			}
			if off, _ := changed.With(key, false); off != (Options{}) {
				t.Fatalf("With(%q, false) = %+v", key, off)
			}
		})
	}
}

func TestUnknownKey(t *testing.T) {
	if on, ok := allOptions.Get("byColor"); on || ok {
		t.Errorf("Get desconhecida = (%v, %v)", on, ok)
	}
	if got, ok := Default.With("byColor", true); ok || got != Default {
		t.Errorf("With desconhecida = (%+v, %v)", got, ok)
	}
}

func TestKeysMatchFields(t *testing.T) {
	if len(Keys) != 6 {
		t.Fatalf("Keys tem %d chaves, want 6", len(Keys))
	}
	if allOptions.Count() != len(Keys) {
		t.Fatalf("toda chave deveria corresponder a um campo de Options")
	}
}
