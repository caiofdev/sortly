package store

import (
	"bytes"
	"errors"
	"log/slog"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

func TestCanUndo(t *testing.T) {
	var nilOp *Operation
	cases := []struct {
		name string
		op   *Operation
		want bool
	}{
		{"nil", nilOp, false},
		{"0 itens", &Operation{}, false},
		{"1 item", &Operation{MovedItems: []MovedItem{{From: "a", To: "b"}}}, true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := tc.op.CanUndo(); got != tc.want {
				t.Errorf("CanUndo = %v, want %v", got, tc.want)
			}
		})
	}
}

func TestDefaultPath(t *testing.T) {
	t.Run("na pasta do usuário", func(t *testing.T) {
		home := t.TempDir()
		t.Setenv("HOME", home)
		t.Setenv("USERPROFILE", home)

		got, err := DefaultPath()
		if err != nil {
			t.Fatal(err)
		}
		if want := filepath.Join(home, ".sortly", "last-operation.json"); got != want {
			t.Errorf("DefaultPath = %q, want %q", got, want)
		}
	})

	t.Run("sem pasta do usuário", func(t *testing.T) {
		t.Setenv("HOME", "")
		t.Setenv("USERPROFILE", "")
		t.Setenv("home", "")

		if _, err := DefaultPath(); err == nil {
			t.Fatal("esperava erro sem pasta do usuário")
		}
	})
}

func TestNewWithoutLogger(t *testing.T) {
	s := New(filepath.Join(t.TempDir(), "x.json"), nil)
	if s.log == nil {
		t.Fatal("logger nil deveria virar um logger que descarta")
	}
	// Não deve entrar em pânico ao registrar (#6).
	_ = s.fail("teste", errors.New("x"))
}

func TestLoadMissingOrUnreadable(t *testing.T) {
	t.Run("arquivo inexistente: nada para desfazer", func(t *testing.T) {
		op, err := newStore(t).Load()
		if op != nil || err != nil {
			t.Fatalf("Load = (%v, %v), want (nil, nil)", op, err)
		}
	})

	t.Run("erro de leitura (caminho é uma pasta)", func(t *testing.T) {
		s := newStore(t)
		if err := os.Mkdir(s.Path(), 0o755); err != nil {
			t.Fatal(err)
		}
		if _, err := s.Load(); err == nil || errors.Is(err, ErrCorrupted) {
			t.Fatalf("err = %v, want erro de leitura", err)
		}
	})

}

func TestLoadCorrupted(t *testing.T) {
	for _, content := range []string{"", "   ", "{quebrado", `{"movedItems":"texto"}`} {
		t.Run("corrompido "+strings.TrimSpace(content), func(t *testing.T) {
			s := storeWith(t, content)
			op, err := s.Load()
			if op != nil || !errors.Is(err, ErrCorrupted) {
				t.Fatalf("Load = (%v, %v), want (nil, ErrCorrupted)", op, err)
			}
		})
	}

}

func TestLoadValid(t *testing.T) {
	t.Run("JSON null", func(t *testing.T) {
		op, err := storeWith(t, "null").Load()
		if op != nil || err != nil {
			t.Fatalf("Load = (%v, %v), want (nil, nil)", op, err)
		}
	})

	t.Run("registro gerado pela versão 1.0", func(t *testing.T) {
		op, err := storeWith(t, golden(t, "electron-last-operation.json")).Load()
		if err != nil {
			t.Fatal(err)
		}
		dest := `C:\Users\ana\Organizados & Cia`
		want := &Operation{
			SourceFolderPath:      `C:\Users\ana\Downloads`,
			DestinationFolderPath: dest,
			MovedItems: []MovedItem{
				{From: `C:\Users\ana\Downloads\relatório.pdf`, To: dest + `\pdf\pages-12\relatório.pdf`},
				{From: `C:\Users\ana\Downloads\foto.jpg`, To: dest + `\jpg\1920x1080\foto (1).jpg`},
			},
			CreatedFolders: []string{dest + `\pdf\pages-12`, dest + `\jpg\1920x1080`},
		}
		if !reflect.DeepEqual(op, want) {
			t.Fatalf("Load =\n%+v\nwant\n%+v", op, want)
		}
	})

	t.Run("registro legado sem createdFolders", func(t *testing.T) {
		op, err := storeWith(t, golden(t, "legacy-without-created-folders.json")).Load()
		if err != nil {
			t.Fatal(err)
		}
		if op.CreatedFolders != nil || len(op.MovedItems) != 1 || !op.CanUndo() {
			t.Fatalf("Load = %+v, want 1 item e createdFolders nil", op)
		}
	})
}

func TestSave(t *testing.T) {
	op := Operation{
		SourceFolderPath:      "/origem",
		DestinationFolderPath: "/destino",
		MovedItems:            []MovedItem{{From: "/origem/a.pdf", To: "/destino/pdf/a.pdf"}},
		CreatedFolders:        []string{"/destino/pdf"},
	}

	t.Run("cria a pasta e grava; Load devolve o mesmo registro", func(t *testing.T) {
		s := New(filepath.Join(t.TempDir(), ".sortly", "last-operation.json"), nil)

		if err := s.Save(op); err != nil {
			t.Fatal(err)
		}
		got, err := s.Load()
		if err != nil || !reflect.DeepEqual(*got, op) {
			t.Fatalf("Load após Save = (%+v, %v), want %+v", got, err, op)
		}
		assertNoTempFiles(t, filepath.Dir(s.Path()))
	})

	t.Run("pasta não pode ser criada", func(t *testing.T) {
		dir := t.TempDir()
		blocker := filepath.Join(dir, "arquivo")
		if err := os.WriteFile(blocker, nil, 0o644); err != nil {
			t.Fatal(err)
		}
		s := New(filepath.Join(blocker, "last-operation.json"), nil)

		if err := s.Save(op); err == nil {
			t.Fatal("esperava erro ao criar pasta dentro de um arquivo")
		}
	})

	t.Run("falha na gravação vai para o log (#6)", func(t *testing.T) {
		var logs bytes.Buffer
		path := filepath.Join(t.TempDir(), "last-operation.json")
		if err := os.MkdirAll(filepath.Join(path, "conteudo"), 0o755); err != nil {
			t.Fatal(err)
		}
		s := New(path, slog.New(slog.NewTextHandler(&logs, nil)))

		if err := s.Save(op); err == nil {
			t.Fatal("esperava erro ao gravar por cima de uma pasta")
		}
		if !strings.Contains(logs.String(), "gravar") {
			t.Fatalf("o erro deveria ser registrado no log, log = %q", logs.String())
		}
	})
}

func TestClear(t *testing.T) {
	t.Run("apaga o registro", func(t *testing.T) {
		s := storeWith(t, "{}")
		if err := s.Clear(); err != nil {
			t.Fatal(err)
		}
		if op, err := s.Load(); op != nil || err != nil {
			t.Fatalf("após Clear, Load = (%v, %v)", op, err)
		}
	})

	t.Run("sem registro não é erro", func(t *testing.T) {
		if err := newStore(t).Clear(); err != nil {
			t.Fatal(err)
		}
	})

	t.Run("não removível devolve erro", func(t *testing.T) {
		s := newStore(t)
		if err := os.MkdirAll(filepath.Join(s.Path(), "conteudo"), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := s.Clear(); err == nil {
			t.Fatal("esperava erro ao apagar pasta não vazia")
		}
	})
}

func TestFailLogs(t *testing.T) {
	var logs bytes.Buffer
	s := New("/x/last-operation.json", slog.New(slog.NewTextHandler(&logs, nil)))
	cause := errors.New("sem permissão")

	err := s.fail("gravar", cause)

	if !errors.Is(err, cause) {
		t.Fatalf("err = %v deveria envolver a causa", err)
	}
	for _, want := range []string{"level=ERROR", "acao=gravar", "sem permissão"} {
		if !strings.Contains(logs.String(), want) {
			t.Errorf("log %q não contém %q", logs.String(), want)
		}
	}
}

func TestEncode(t *testing.T) {
	item := []MovedItem{{From: "a/x.txt", To: "b/txt/x.txt"}}
	folder := []string{"b/txt"}
	const movedJSON = `[{"from":"a/x.txt","to":"b/txt/x.txt"}]`
	tests := []struct {
		name string
		op   Operation
		want string
	}{
		{"listas nil viram []", Operation{SourceFolderPath: "a"},
			`{"sourceFolderPath":"a","destinationFolderPath":"","movedItems":[],"createdFolders":[]}`},
		{"só movedItems preenchida", Operation{SourceFolderPath: "a", MovedItems: item},
			`{"sourceFolderPath":"a","destinationFolderPath":"","movedItems":` + movedJSON + `,"createdFolders":[]}`},
		{"só createdFolders preenchida", Operation{SourceFolderPath: "a", CreatedFolders: folder},
			`{"sourceFolderPath":"a","destinationFolderPath":"","movedItems":[],"createdFolders":["b/txt"]}`},
		{"as duas preenchidas", Operation{SourceFolderPath: "a", DestinationFolderPath: "b", MovedItems: item, CreatedFolders: folder},
			`{"sourceFolderPath":"a","destinationFolderPath":"b","movedItems":` + movedJSON + `,"createdFolders":["b/txt"]}`},
		// Os campos da #82 só aparecem quando há substituídos (ADR 0008, #82).
		{"com substituídos", Operation{SourceFolderPath: "a", MovedItems: item,
			ReplacedItems: []ReplacedItem{{Path: "b/txt/x.txt", Backup: "s/1/txt/x.txt"}}, BackupFolder: "s/1"},
			`{"sourceFolderPath":"a","destinationFolderPath":"","movedItems":` + movedJSON +
				`,"createdFolders":[],"replacedItems":[{"path":"b/txt/x.txt","backup":"s/1/txt/x.txt"}],"backupFolder":"s/1"}`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := encode(tt.op)
			if err != nil || string(got) != tt.want {
				t.Fatalf("encode = (%s, %v), want %s", got, err, tt.want)
			}
		})
	}

	t.Run("& e acentos sem escape", func(t *testing.T) {
		got, _ := encode(Operation{SourceFolderPath: "Fotos & Vídeos <2026>"})
		if !strings.Contains(string(got), `"Fotos & Vídeos <2026>"`) {
			t.Fatalf("encode = %s", got)
		}
	})

	t.Run("grava byte a byte igual à versão 1.0", func(t *testing.T) {
		want := golden(t, "electron-last-operation.json")
		op, err := storeWith(t, want).Load()
		if err != nil {
			t.Fatal(err)
		}
		got, err := encode(*op)
		if err != nil {
			t.Fatal(err)
		}
		if string(got) != want {
			t.Fatalf("encode =\n%s\nwant\n%s", got, want)
		}
	})
}

func newStore(t *testing.T) *FileStore {
	t.Helper()
	return New(filepath.Join(t.TempDir(), "last-operation.json"), nil)
}

func storeWith(t *testing.T, content string) *FileStore {
	t.Helper()
	s := newStore(t)
	if err := os.WriteFile(s.Path(), []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
	return s
}

func golden(t *testing.T, name string) string {
	t.Helper()
	data, err := os.ReadFile(filepath.Join("testdata", name))
	if err != nil {
		t.Fatal(err)
	}
	return strings.TrimRight(string(data), "\r\n")
}

func assertNoTempFiles(t *testing.T, dir string) {
	t.Helper()
	matches, _ := filepath.Glob(filepath.Join(dir, ".last-operation.json-*.tmp"))
	if len(matches) > 0 {
		t.Fatalf("arquivos temporários esquecidos: %v", matches)
	}
}

func TestReplacedItemsRoundTrip(t *testing.T) {
	s := newStore(t)
	op := Operation{SourceFolderPath: "a", MovedItems: []MovedItem{{From: "a/x", To: "b/x"}},
		ReplacedItems: []ReplacedItem{{Path: "b/x", Backup: "s/1/x"}}, BackupFolder: "s/1"}
	if err := s.Save(op); err != nil {
		t.Fatal(err)
	}
	got, err := s.Load()
	if err != nil || !reflect.DeepEqual(got.ReplacedItems, op.ReplacedItems) || got.BackupFolder != "s/1" {
		t.Fatalf("Load = (%+v, %v)", got, err)
	}
}
