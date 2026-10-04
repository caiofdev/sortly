package criteria

import (
	"context"
	"testing"
)

func TestSizeRule(t *testing.T) {
	cases := []struct {
		bytes int64
		want  string
	}{
		{0, "size-1mb"},
		{1, "size-1mb"},
		{1048575, "size-1mb"},
		{1048576, "size-1mb"},
		{1048577, "size-2mb"},
		{2097152, "size-2mb"},
		{2097153, "size-3mb"},
	}
	for _, tc := range cases {
		got, _ := sizeRule{}.Segment(context.Background(), File{Size: tc.bytes})
		if got != tc.want {
			t.Errorf("%d bytes = %q, want %q", tc.bytes, got, tc.want)
		}
	}
}
