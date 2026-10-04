package criteria

import (
	"context"
	"time"
)

type dateRule struct{ loc *time.Location }

func (dateRule) Enabled(o Options) bool { return o.ByDate }
func (dateRule) Applies(File) bool      { return true }

// Segment usa o fuso local, não UTC: é a data que o usuário vê no gerenciador de arquivos.
func (r dateRule) Segment(_ context.Context, f File) (string, error) {
	return "date-" + f.ModTime.In(r.loc).Format("2006-01-02"), nil
}
