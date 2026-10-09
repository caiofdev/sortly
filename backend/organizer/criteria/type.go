package criteria

import "context"

// Nomes neutros, como os dos outros critérios; a interface traduz o rótulo
// (ADR 0004, #81).
const (
	TypeImages     = "images"
	TypeDocuments  = "documents"
	TypeArchives   = "archives"
	TypeInstallers = "installers"
	TypeVideos     = "videos"
	TypeAudio      = "audio"
	TypeOther      = "other"
)

// A tabela completa está em docs/organization-rules.md (#81).
var typeExtensions = map[string][]string{
	TypeImages: {"jpg", "jpeg", "png", "gif", "bmp", "webp", "svg", "tif", "tiff", "heic", "heif", "ico",
		"raw", "cr2", "nef", "arw", "dng", "psd"},
	TypeDocuments: {"pdf", "doc", "docx", "odt", "rtf", "txt", "md", "xls", "xlsx", "ods", "csv",
		"ppt", "pptx", "odp", "epub"},
	TypeArchives:   {"zip", "rar", "7z", "tar", "gz", "tgz", "bz2", "xz", "zst"},
	TypeInstallers: {"exe", "msi", "msix", "dmg", "pkg", "deb", "rpm", "appimage", "apk"},
	TypeVideos:     {"mp4", "mkv", "avi", "mov", "wmv", "webm", "m4v", "flv", "mpg", "mpeg", "3gp"},
	TypeAudio:      {"mp3", "wav", "flac", "aac", "ogg", "oga", "m4a", "wma", "opus", "aiff"},
}

var typeByExt = categories(typeExtensions)

func categories(byType map[string][]string) map[string]string {
	byExt := map[string]string{}
	for category, exts := range byType {
		for _, ext := range exts {
			byExt[ext] = category
		}
	}
	return byExt
}

// Vale para todo arquivo: sem extensão ou com uma desconhecida, vai para
// "other". Com a extensão também ligada, o arquivo sem extensão continua
// ignorado, como na 1.0 (#81).
type typeRule struct{}

func (typeRule) Enabled(o Options) bool { return o.ByType }
func (typeRule) Applies(File) bool      { return true }

func (typeRule) Segment(_ context.Context, f File) (string, error) {
	if category, ok := typeByExt[f.Ext]; ok {
		return category, nil
	}
	return TypeOther, nil
}
