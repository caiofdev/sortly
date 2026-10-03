package apperr

// função     | CC | casos
// New        |  1 | TestError
// Error.*    |  1 | TestError
// CodeOf     |  3 | TestCodeOf: nil; com código (direto, envolvido, unido); sem código

import (
	"errors"
	"fmt"
	"testing"
)

var errSample = New("SAMPLE", "erro de exemplo")

func TestError(t *testing.T) {
	if errSample.Code() != "SAMPLE" || errSample.Error() != "erro de exemplo" {
		t.Fatalf("Error = (%q, %q)", errSample.Code(), errSample.Error())
	}
}

func TestCodeOf(t *testing.T) {
	cases := []struct {
		name string
		err  error
		want string
	}{
		{"nil", nil, ""},
		{"direto", errSample, "SAMPLE"},
		{"envolvido", fmt.Errorf("%w: /pasta", errSample), "SAMPLE"},
		{"unido a outro erro", errors.Join(errors.New("x"), errSample), "SAMPLE"},
		{"sem código", errors.New("x"), CodeUnexpected},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := CodeOf(tc.err); got != tc.want {
				t.Fatalf("CodeOf = %q, want %q", got, tc.want)
			}
		})
	}
	if !errors.Is(fmt.Errorf("%w: x", errSample), errSample) {
		t.Fatal("errors.Is deveria reconhecer a sentinela envolvida")
	}
}
