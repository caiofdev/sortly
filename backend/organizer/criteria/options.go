package criteria

// Options diz quais critérios estão ligados.
type Options struct {
	ByDuration   bool
	ByPages      bool
	ByResolution bool
	ByDate       bool
	BySize       bool
	ByExtension  bool
}

// Default é o padrão do primeiro uso: só a extensão.
var Default = Options{ByExtension: true}

// Keys são as chaves dos critérios na ordem em que a interface os mostra. É
// diferente da ordem de aninhamento das pastas (veja New).
var Keys = []string{"byDuration", "byPages", "byResolution", "byDate", "bySize", "byExtension"}

// Any informa se pelo menos um critério está ligado.
func (o Options) Any() bool {
	return o != Options{}
}

// Count devolve quantos critérios estão ligados.
func (o Options) Count() int {
	n := 0
	for _, key := range Keys {
		if on, _ := o.Get(key); on {
			n++
		}
	}
	return n
}

// Get devolve o valor do critério com a chave dada; ok é false para chave desconhecida.
func (o Options) Get(key string) (on, ok bool) {
	field := o.field(key)
	if field == nil {
		return false, false
	}
	return *field, true
}

// With devolve uma cópia com o critério alterado; ok é false para chave desconhecida.
func (o Options) With(key string, on bool) (Options, bool) {
	field := o.field(key)
	if field == nil {
		return o, false
	}
	*field = on
	return o, true
}

func (o *Options) field(key string) *bool {
	switch key {
	case "byDuration":
		return &o.ByDuration
	case "byPages":
		return &o.ByPages
	case "byResolution":
		return &o.ByResolution
	case "byDate":
		return &o.ByDate
	case "bySize":
		return &o.BySize
	case "byExtension":
		return &o.ByExtension
	}
	return nil
}
