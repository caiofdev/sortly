package criteria

type Options struct {
	ByDuration   bool
	ByPages      bool
	ByResolution bool
	ByDate       bool
	BySize       bool
	ByExtension  bool
}

var Default = Options{ByExtension: true}

// Ordem em que a interface mostra os critérios, diferente da ordem de
// aninhamento das pastas (veja New) (#44).
var Keys = []string{"byDuration", "byPages", "byResolution", "byDate", "bySize", "byExtension"}

func (o Options) Any() bool {
	return o != Options{}
}

func (o Options) Count() int {
	n := 0
	for _, key := range Keys {
		if on, _ := o.Get(key); on {
			n++
		}
	}
	return n
}

func (o Options) Get(key string) (on, ok bool) {
	field := o.field(key)
	if field == nil {
		return false, false
	}
	return *field, true
}

func (o Options) With(key string, on bool) (Options, bool) {
	field := o.field(key)
	if field == nil {
		return o, false
	}
	*field = on
	return o, true
}

var fields = map[string]func(*Options) *bool{
	"byDuration":   func(o *Options) *bool { return &o.ByDuration },
	"byPages":      func(o *Options) *bool { return &o.ByPages },
	"byResolution": func(o *Options) *bool { return &o.ByResolution },
	"byDate":       func(o *Options) *bool { return &o.ByDate },
	"bySize":       func(o *Options) *bool { return &o.BySize },
	"byExtension":  func(o *Options) *bool { return &o.ByExtension },
}

func (o *Options) field(key string) *bool {
	get, ok := fields[key]
	if !ok {
		return nil
	}
	return get(o)
}
