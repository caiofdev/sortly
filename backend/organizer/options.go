package organizer

// RawOptions são os critérios como chegam do frontend. ByExtension é ponteiro
// porque, como na versão Electron, ele vale true quando não é informado.
type RawOptions struct {
	ByDuration   bool  `json:"byDuration"`
	ByPages      bool  `json:"byPages"`
	ByResolution bool  `json:"byResolution"`
	ByDate       bool  `json:"byDate"`
	BySize       bool  `json:"bySize"`
	ByExtension  *bool `json:"byExtension"`
}

// Options são os critérios já normalizados.
type Options struct {
	ByDuration   bool
	ByPages      bool
	ByResolution bool
	ByDate       bool
	BySize       bool
	ByExtension  bool
}

// Normalize aplica o padrão de ByExtension (ligado quando ausente).
func (r RawOptions) Normalize() Options {
	return Options{
		ByDuration:   r.ByDuration,
		ByPages:      r.ByPages,
		ByResolution: r.ByResolution,
		ByDate:       r.ByDate,
		BySize:       r.BySize,
		ByExtension:  r.ByExtension == nil || *r.ByExtension,
	}
}

// Any informa se pelo menos um critério está ligado.
func (o Options) Any() bool {
	return o != Options{}
}
