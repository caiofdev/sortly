// Package settings guarda as preferências do usuário (idioma, tema e critérios) em
// ~/.sortly/settings.json. Fora do WebView, elas sobrevivem à limpeza dos dados
// do navegador embutido e às atualizações do app. O caminho do arquivo é
// decidido na composição, em app.NewDefault (#44).
package settings

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"log/slog"
	"os"
	"sync"

	"github.com/caiofdev/sortly/backend/apperr"
	"github.com/caiofdev/sortly/backend/fs/files"
	"github.com/caiofdev/sortly/backend/organizer/criteria"
)

var (
	ErrInvalidLanguage  = apperr.New("INVALID_LANGUAGE", "settings: idioma desconhecido")
	ErrInvalidTheme     = apperr.New("INVALID_THEME", "settings: tema desconhecido")
	ErrUnknownCriterion = apperr.New("UNKNOWN_CRITERION", "settings: critério desconhecido")
	ErrLastCriterion    = apperr.New("LAST_CRITERION", "settings: pelo menos um critério precisa ficar ligado")
	ErrNotSaved         = apperr.New("SETTINGS_NOT_SAVED", "settings: preferências não foram salvas")
)

// O escuro é o tema principal do design system, o da logo (#76).
const (
	DefaultLanguage = "pt-BR"
	DefaultTheme    = "dark"
)

var (
	languages = map[string]bool{"pt-BR": true, "en": true}
	themes    = map[string]bool{"dark": true, "light": true}
)

type Settings struct {
	Language string
	Theme    string
	Options  criteria.Options
}

func Default() Settings {
	return Settings{Language: DefaultLanguage, Theme: DefaultTheme, Options: criteria.Default}
}

// Seguro para uso concorrente: os bindings do Wails rodam em goroutines separadas (#44).
type Service struct {
	mu      sync.Mutex
	path    string
	log     *slog.Logger
	current *Settings
}

// Com log nil, nada é registrado (#44).
func New(path string, log *slog.Logger) *Service {
	if log == nil {
		log = slog.New(slog.DiscardHandler)
	}
	return &Service{path: path, log: log}
}

func (s *Service) Get() Settings {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.load()
}

func (s *Service) SetLanguage(lang string) (Settings, error) {
	return s.update(func(st Settings) (Settings, error) {
		if !languages[lang] {
			return st, fmt.Errorf("%w: %q", ErrInvalidLanguage, lang)
		}
		st.Language = lang
		return st, nil
	})
}

func (s *Service) SetTheme(theme string) (Settings, error) {
	return s.update(func(st Settings) (Settings, error) {
		if !themes[theme] {
			return st, fmt.Errorf("%w: %q", ErrInvalidTheme, theme)
		}
		st.Theme = theme
		return st, nil
	})
}

// Desligar o último critério ligado é recusado (#44).
func (s *Service) SetCriterion(key string, enabled bool) (Settings, error) {
	return s.update(func(st Settings) (Settings, error) {
		opts, ok := st.Options.With(key, enabled)
		if !ok {
			return st, fmt.Errorf("%w: %q", ErrUnknownCriterion, key)
		}
		if !opts.Any() {
			return st, ErrLastCriterion
		}
		st.Options = opts
		return st, nil
	})
}

// Se a validação ou a gravação falhar, as preferências em memória continuam as anteriores
// (#44).
func (s *Service) update(change func(Settings) (Settings, error)) (Settings, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	old := s.load()
	next, err := change(old)
	if err != nil || next == old {
		return old, err
	}
	if err := s.save(next); err != nil {
		return old, err
	}
	s.current = &next
	return next, nil
}

// Lido só na primeira chamada. Arquivo ausente, ilegível ou corrompido vale
// como as preferências padrão; só os dois últimos vão para o log (#44).
func (s *Service) load() Settings {
	if s.current != nil {
		return *s.current
	}
	st := Default()
	data, err := os.ReadFile(s.path)
	if err == nil {
		st = s.decode(data)
	}
	if err != nil && !errors.Is(err, fs.ErrNotExist) {
		s.log.Warn("preferências ilegíveis; usando o padrão", "path", s.path, "err", err)
	}
	s.current = &st
	return st
}

// Os critérios ficam num mapa para que chaves desconhecidas (de uma versão
// futura) sejam ignoradas em vez de quebrar a leitura (#44).
type fileFormat struct {
	Language            string          `json:"language"`
	Theme               string          `json:"theme"`
	OrganizationOptions map[string]bool `json:"organizationOptions"`
}

// Só valores válidos: idioma ou tema desconhecido (ou ausente, num arquivo
// anterior ao tema) e nenhum critério ligado voltam ao padrão (#44, #76).
func (s *Service) decode(data []byte) Settings {
	st := Default()
	var f fileFormat
	if err := json.Unmarshal(data, &f); err != nil {
		s.log.Warn("preferências corrompidas; usando o padrão", "path", s.path, "err", err)
		return st
	}
	if languages[f.Language] {
		st.Language = f.Language
	}
	if themes[f.Theme] {
		st.Theme = f.Theme
	}
	var opts criteria.Options
	for key, on := range f.OrganizationOptions {
		opts, _ = opts.With(key, on)
	}
	if opts.Any() {
		st.Options = opts
	}
	return st
}

func (s *Service) save(st Settings) error {
	f := fileFormat{Language: st.Language, Theme: st.Theme, OrganizationOptions: map[string]bool{}}
	for _, key := range criteria.Keys {
		f.OrganizationOptions[key], _ = st.Options.Get(key)
	}
	// Marshal de strings e map[string]bool não falha (#44).
	data, _ := json.Marshal(f)
	if err := files.WriteAtomic(s.path, data); err != nil {
		s.log.Error("preferências não salvas", "path", s.path, "err", err)
		return fmt.Errorf("%w: %w", ErrNotSaved, err)
	}
	return nil
}
