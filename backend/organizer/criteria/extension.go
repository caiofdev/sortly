package criteria

import "context"

type extensionRule struct{}

func (extensionRule) Enabled(o Options) bool { return o.ByExtension }
func (extensionRule) Applies(File) bool      { return true }

func (extensionRule) Segment(_ context.Context, f File) (string, error) {
	if f.Ext == "" {
		return "", ErrSkipNoExtension
	}
	return f.Ext, nil
}
