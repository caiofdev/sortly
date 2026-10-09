package criteria

import (
	"context"
	"testing"
)

func TestTypeRule(t *testing.T) {
	tests := []struct {
		ext, want string
	}{
		{"jpg", TypeImages},
		{"heic", TypeImages},
		{"pdf", TypeDocuments},
		{"csv", TypeDocuments},
		{"zip", TypeArchives},
		{"gz", TypeArchives},
		{"exe", TypeInstallers},
		{"dmg", TypeInstallers},
		{"mp4", TypeVideos},
		{"mp3", TypeAudio},
		{"xyz", TypeOther},
		{"", TypeOther},
	}
	for _, tt := range tests {
		got, err := typeRule{}.Segment(context.Background(), File{Ext: tt.ext})
		if err != nil || got != tt.want {
			t.Errorf("Segment(%q) = (%q, %v), want %q", tt.ext, got, err, tt.want)
		}
	}
}

// Uma extensão repetida em duas categorias ficaria só na última lida do mapa,
// numa ordem que muda a cada execução (#81).
func TestTypeExtensionsDoNotRepeat(t *testing.T) {
	listed := 0
	for _, exts := range typeExtensions {
		listed += len(exts)
	}
	if listed != len(typeByExt) {
		t.Fatalf("%d extensões listadas, %d distintas: há repetidas", listed, len(typeByExt))
	}
}
