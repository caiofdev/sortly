package settings

import (
	"reflect"
	"testing"

	"github.com/caiofdev/sortly/backend/organizer/criteria"
)

func TestView(t *testing.T) {
	t.Run("um critério ligado fica travado", func(t *testing.T) {
		got := Default().View()
		want := View{Language: DefaultLanguage, Theme: DefaultTheme, Criteria: []Criterion{
			{Key: "byDuration"}, {Key: "byPages"}, {Key: "byResolution"},
			{Key: "byDate"}, {Key: "bySize"}, {Key: "byExtension", Enabled: true, Locked: true},
		}}
		if !reflect.DeepEqual(got, want) {
			t.Fatalf("View = %+v, want %+v", got, want)
		}
	})

	t.Run("sem critério ligado, nenhum fica travado", func(t *testing.T) {
		for _, c := range (Settings{Language: "en"}).View().Criteria {
			if c.Enabled || c.Locked {
				t.Fatalf("%s = %+v, want desligado e destravado", c.Key, c)
			}
		}
	})

	t.Run("com dois ligados, nenhum fica travado", func(t *testing.T) {
		got := Settings{Language: "en", Options: criteria.Options{ByDate: true, ByExtension: true}}.View()
		for _, c := range got.Criteria {
			if c.Locked {
				t.Fatalf("%s travado com dois critérios ligados", c.Key)
			}
		}
		if got.Language != "en" || len(got.Criteria) != len(criteria.Keys) {
			t.Fatalf("View = %+v", got)
		}
	})
}
