package settings

import "github.com/caiofdev/sortly/backend/organizer/criteria"

// Criterion é um critério como a interface o mostra. Locked marca o último
// critério ligado, que não pode ser desmarcado.
type Criterion struct {
	Key     string `json:"key"`
	Enabled bool   `json:"enabled"`
	Locked  bool   `json:"locked"`
}

// View são as preferências prontas para a interface: o idioma e os critérios
// na ordem dos checkboxes, já com a regra do último critério aplicada.
type View struct {
	Language string      `json:"language"`
	Criteria []Criterion `json:"criteria"`
}

// View monta a visão das preferências para a interface.
func (st Settings) View() View {
	lastOne := st.Options.Count() == 1
	list := make([]Criterion, 0, len(criteria.Keys))
	for _, key := range criteria.Keys {
		on, _ := st.Options.Get(key)
		list = append(list, Criterion{Key: key, Enabled: on, Locked: on && lastOne})
	}
	return View{Language: st.Language, Criteria: list}
}
