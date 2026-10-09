package settings

import (
	"bytes"
	"errors"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/caiofdev/sortly/backend/organizer/criteria"
)

func newService(t *testing.T, content string) *Service {
	t.Helper()
	path := filepath.Join(t.TempDir(), "settings.json")
	if content != "" {
		if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
			t.Fatal(err)
		}
	}
	return New(path, nil)
}

func TestGet(t *testing.T) {
	cases := []struct {
		name    string
		content string
		want    Settings
	}{
		{"sem arquivo", "", Default()},
		{"válido", `{"language":"en","organizationOptions":{"byDate":true,"bySize":true,"byType":false}}`,
			Settings{Language: "en", Theme: DefaultTheme, Duplicates: "rename", Options: criteria.Options{ByDate: true, BySize: true}}},
		{"vazio", " ", Default()},
		{"JSON corrompido", `{"language":`, Default()},
		{"null", "null", Default()},
		{"campos ausentes", `{}`, Default()},
		{"idioma desconhecido", `{"language":"fr","organizationOptions":{"byPages":true}}`,
			Settings{Language: DefaultLanguage, Theme: DefaultTheme, Duplicates: "rename", Options: criteria.Options{ByPages: true}}},
		{"nenhum critério ligado", `{"language":"en","organizationOptions":{"byExtension":false}}`,
			Settings{Language: "en", Theme: DefaultTheme, Duplicates: "rename", Options: criteria.Default}},
		{"chave desconhecida é ignorada", `{"language":"pt-BR","organizationOptions":{"byColor":true,"byDate":true}}`,
			Settings{Language: "pt-BR", Theme: DefaultTheme, Duplicates: "rename", Options: criteria.Options{ByDate: true}}},
		{"tema claro", `{"language":"en","theme":"light","organizationOptions":{"byDate":true}}`,
			Settings{Language: "en", Theme: "light", Duplicates: "rename", Options: criteria.Options{ByDate: true}}},
		{"tema desconhecido", `{"language":"en","theme":"sepia","organizationOptions":{"byDate":true}}`,
			Settings{Language: "en", Theme: DefaultTheme, Duplicates: "rename", Options: criteria.Options{ByDate: true}}},
		{"substituir", `{"language":"en","duplicates":"replace","organizationOptions":{"byDate":true}}`,
			Settings{Language: "en", Theme: DefaultTheme, Duplicates: "replace", Options: criteria.Options{ByDate: true}}},
		{"política desconhecida", `{"language":"en","duplicates":"mesclar","organizationOptions":{"byDate":true}}`,
			Settings{Language: "en", Theme: DefaultTheme, Duplicates: "rename", Options: criteria.Options{ByDate: true}}},
		{"6 critérios", `{"language":"en","organizationOptions":{"byDuration":true,"byPages":true,"byResolution":true,"byDate":true,"bySize":true,"byExtension":true}}`,
			Settings{Language: "en", Theme: DefaultTheme, Duplicates: "rename", Options: criteria.Options{ByDuration: true, ByPages: true, ByResolution: true, ByDate: true, BySize: true, ByExtension: true}}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := newService(t, tc.content).Get(); got != tc.want {
				t.Fatalf("Get = %+v, want %+v", got, tc.want)
			}
		})
	}
}

func TestGetUnreadableAndCorruptedAreLogged(t *testing.T) {
	cases := []struct {
		name  string
		setup func(path string) error
		log   string
	}{
		{"ilegível (é uma pasta)", func(path string) error { return os.Mkdir(path, 0o755) }, "ilegíveis"},
		{"corrompido", func(path string) error { return os.WriteFile(path, []byte("{"), 0o600) }, "corrompidas"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			path := filepath.Join(t.TempDir(), "settings.json")
			if err := tc.setup(path); err != nil {
				t.Fatal(err)
			}
			var logs bytes.Buffer
			s := New(path, slog.New(slog.NewTextHandler(&logs, nil)))
			if got := s.Get(); got != Default() {
				t.Fatalf("Get = %+v, want padrão", got)
			}
			if !strings.Contains(logs.String(), tc.log) {
				t.Fatalf("log = %q, want %q", logs.String(), tc.log)
			}
		})
	}
}

func TestGetReadsFileOnce(t *testing.T) {
	s := newService(t, `{"language":"en","organizationOptions":{"byDate":true}}`)
	first := s.Get()
	if err := os.WriteFile(s.path, []byte(`{"language":"pt-BR"}`), 0o600); err != nil {
		t.Fatal(err)
	}
	if got := s.Get(); got != first {
		t.Fatalf("segunda leitura = %+v, want o valor em memória %+v", got, first)
	}
}

func TestSetLanguage(t *testing.T) {
	t.Run("idioma válido é salvo", func(t *testing.T) {
		s := newService(t, "")
		got, err := s.SetLanguage("en")
		if err != nil || got.Language != "en" {
			t.Fatalf("SetLanguage = (%+v, %v)", got, err)
		}
		if reread := New(s.path, nil).Get(); reread != got {
			t.Fatalf("relido do disco = %+v, want %+v", reread, got)
		}
	})

	t.Run("idioma desconhecido é recusado", func(t *testing.T) {
		s := newService(t, "")
		got, err := s.SetLanguage("fr")
		if !errors.Is(err, ErrInvalidLanguage) || got != Default() {
			t.Fatalf("SetLanguage = (%+v, %v), want (padrão, ErrInvalidLanguage)", got, err)
		}
	})
}

func TestSetTheme(t *testing.T) {
	t.Run("tema claro é salvo", func(t *testing.T) {
		s := newService(t, "")
		got, err := s.SetTheme("light")
		if err != nil || got.Theme != "light" {
			t.Fatalf("SetTheme = (%+v, %v)", got, err)
		}
		if reread := New(s.path, nil).Get(); reread != got {
			t.Fatalf("relido do disco = %+v, want %+v", reread, got)
		}
	})

	t.Run("tema desconhecido é recusado", func(t *testing.T) {
		s := newService(t, "")
		got, err := s.SetTheme("sepia")
		if !errors.Is(err, ErrInvalidTheme) || got != Default() {
			t.Fatalf("SetTheme = (%+v, %v), want (padrão, ErrInvalidTheme)", got, err)
		}
	})
}

func TestSetDuplicates(t *testing.T) {
	t.Run("ignorar é salvo", func(t *testing.T) {
		s := newService(t, "")
		got, err := s.SetDuplicates("skip")
		if err != nil || got.Duplicates != "skip" {
			t.Fatalf("SetDuplicates = (%+v, %v)", got, err)
		}
		if reread := New(s.path, nil).Get(); reread != got {
			t.Fatalf("relido do disco = %+v, want %+v", reread, got)
		}
	})

	t.Run("política desconhecida é recusada", func(t *testing.T) {
		s := newService(t, "")
		got, err := s.SetDuplicates("mesclar")
		if !errors.Is(err, ErrInvalidDuplicates) || got != Default() {
			t.Fatalf("SetDuplicates = (%+v, %v), want (padrão, ErrInvalidDuplicates)", got, err)
		}
	})
}

func TestSetCriterion(t *testing.T) {
	t.Run("liga um critério", func(t *testing.T) {
		got, err := newService(t, "").SetCriterion("byDate", true)
		want := criteria.Options{ByDate: true, ByExtension: true}
		if err != nil || got.Options != want {
			t.Fatalf("SetCriterion = (%+v, %v), want %+v", got, err, want)
		}
	})

	t.Run("desliga um de dois", func(t *testing.T) {
		s := newService(t, `{"language":"pt-BR","organizationOptions":{"byDate":true,"byExtension":true}}`)
		got, err := s.SetCriterion("byExtension", false)
		if err != nil || got.Options != (criteria.Options{ByDate: true}) {
			t.Fatalf("SetCriterion = (%+v, %v)", got, err)
		}
	})

	t.Run("desligar o último é recusado", func(t *testing.T) {
		got, err := newService(t, "").SetCriterion("byExtension", false)
		if !errors.Is(err, ErrLastCriterion) || got != Default() {
			t.Fatalf("SetCriterion = (%+v, %v), want (padrão, ErrLastCriterion)", got, err)
		}
	})

	t.Run("chave desconhecida é recusada", func(t *testing.T) {
		got, err := newService(t, "").SetCriterion("byColor", true)
		if !errors.Is(err, ErrUnknownCriterion) || got != Default() {
			t.Fatalf("SetCriterion = (%+v, %v), want (padrão, ErrUnknownCriterion)", got, err)
		}
	})
}

func TestUpdateWithoutChangeDoesNotWrite(t *testing.T) {
	s := newService(t, "")
	if _, err := s.SetLanguage(DefaultLanguage); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(s.path); !os.IsNotExist(err) {
		t.Fatalf("o arquivo não deveria ter sido criado: %v", err)
	}
}

func TestSaveFailureKeepsPreviousSettings(t *testing.T) {
	dir := t.TempDir()
	blocker := filepath.Join(dir, "arquivo")
	if err := os.WriteFile(blocker, nil, 0o600); err != nil {
		t.Fatal(err)
	}
	var logs bytes.Buffer
	s := New(filepath.Join(blocker, "settings.json"), slog.New(slog.NewTextHandler(&logs, nil)))

	got, err := s.SetLanguage("en")
	if !errors.Is(err, ErrNotSaved) || got != Default() {
		t.Fatalf("SetLanguage = (%+v, %v), want (padrão, ErrNotSaved)", got, err)
	}
	if s.Get() != Default() {
		t.Fatal("a falha na gravação não deveria mudar as preferências em memória")
	}
	if !strings.Contains(logs.String(), "não salvas") {
		t.Fatalf("log = %q", logs.String())
	}
}

func TestSavedFileFormat(t *testing.T) {
	s := newService(t, "")
	if _, err := s.SetCriterion("bySize", true); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(s.path)
	if err != nil {
		t.Fatal(err)
	}
	want := `{"language":"pt-BR","theme":"dark","duplicates":"rename","organizationOptions":{"byDate":false,"byDuration":false,"byExtension":true,"byPages":false,"byResolution":false,"bySize":true,"byType":false}}`
	if string(data) != want {
		t.Fatalf("arquivo =\n%s\nwant\n%s", data, want)
	}
}
