package settings

import "github.com/caiofdev/sortly/backend/organizer/criteria"

// Locked marca o último critério ligado, que não pode ser desmarcado (#44).
type Criterion struct {
	Key     string `json:"key"`
	Enabled bool   `json:"enabled"`
	Locked  bool   `json:"locked"`
}

// Os critérios vêm na ordem dos checkboxes, já com a regra do último critério aplicada
// (#44).
type View struct {
	Language   string      `json:"language"`
	Theme      string      `json:"theme"`
	Duplicates string      `json:"duplicates"`
	Criteria   []Criterion `json:"criteria"`
}

func (st Settings) View() View {
	lastOne := st.Options.Count() == 1
	list := make([]Criterion, 0, len(criteria.Keys))
	for _, key := range criteria.Keys {
		on, _ := st.Options.Get(key)
		list = append(list, Criterion{Key: key, Enabled: on, Locked: on && lastOne})
	}
	return View{Language: st.Language, Theme: st.Theme, Duplicates: st.Duplicates, Criteria: list}
}
