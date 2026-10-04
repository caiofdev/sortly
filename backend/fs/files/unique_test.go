package files

// função            | CC | casos
// Reserve           |  4 | TestReserve: livre; ocupado; vários ocupados; pasta com o nome; pasta pai inexistente; limite
// candidateName     |  2 | TestCandidateName: n = 0; n > 0 (com e sem extensão, oculto, ponto final)
// isOccupied        |  2 | TestIsOccupied: ErrExist; outro erro com algo no caminho; outro erro sem nada
// closeReservation  |  2 | TestCloseReservation: sucesso (via TestReserve); arquivo já fechado
//
// Valor-limite: tentativas = limite - 1 (ainda encontra nome) e = limite (ErrNoAvailableName);
// nome que já termina em " (1)".

import (
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"testing"
)

func TestCandidateName(t *testing.T) {
	dir := t.TempDir()

	cases := []struct {
		name string
		n    int
		want string
	}{
		{"foto.jpg", 0, "foto.jpg"},
		{"foto.jpg", 1, "foto (1).jpg"},
		{"foto.jpg", 2, "foto (2).jpg"},
		{"LEIAME", 1, "LEIAME (1)"},
		{".gitignore", 1, ".gitignore (1)"},
		{"backup.tar.gz", 1, "backup.tar (1).gz"},
		{"arquivo.", 1, "arquivo (1)."},
		{"foto (1).jpg", 1, "foto (1) (1).jpg"},
	}

	for _, tc := range cases {
		t.Run(tc.want, func(t *testing.T) {
			got := candidateName(filepath.Join(dir, tc.name), tc.n)
			if want := filepath.Join(dir, tc.want); got != want {
				t.Errorf("candidateName(%q, %d) = %q, want %q", tc.name, tc.n, got, want)
			}
		})
	}
}

func TestReserve(t *testing.T) {
	t.Run("nome livre", func(t *testing.T) {
		dir := t.TempDir()
		got, err := Reserve(filepath.Join(dir, "foto.jpg"))
		assertReserved(t, got, err, filepath.Join(dir, "foto.jpg"))
	})

	t.Run("nome ocupado vira (1)", func(t *testing.T) {
		dir := t.TempDir()
		touch(t, filepath.Join(dir, "foto.jpg"))
		got, err := Reserve(filepath.Join(dir, "foto.jpg"))
		assertReserved(t, got, err, filepath.Join(dir, "foto (1).jpg"))
	})

	t.Run("nome e (1) ocupados vira (2)", func(t *testing.T) {
		dir := t.TempDir()
		touch(t, filepath.Join(dir, "foto.jpg"))
		touch(t, filepath.Join(dir, "foto (1).jpg"))
		got, err := Reserve(filepath.Join(dir, "foto.jpg"))
		assertReserved(t, got, err, filepath.Join(dir, "foto (2).jpg"))
	})

	t.Run("nome sem extensão", func(t *testing.T) {
		dir := t.TempDir()
		touch(t, filepath.Join(dir, "LEIAME"))
		got, err := Reserve(filepath.Join(dir, "LEIAME"))
		assertReserved(t, got, err, filepath.Join(dir, "LEIAME (1)"))
	})

	t.Run("pasta com o mesmo nome conta como ocupado", func(t *testing.T) {
		dir := t.TempDir()
		if err := os.Mkdir(filepath.Join(dir, "foto.jpg"), 0o755); err != nil {
			t.Fatal(err)
		}
		got, err := Reserve(filepath.Join(dir, "foto.jpg"))
		assertReserved(t, got, err, filepath.Join(dir, "foto (1).jpg"))
	})

	t.Run("pasta pai inexistente devolve o erro", func(t *testing.T) {
		got, err := Reserve(filepath.Join(t.TempDir(), "nao-existe", "foto.jpg"))
		if err == nil || errors.Is(err, ErrNoAvailableName) {
			t.Fatalf("Reserve = (%q, %v), want erro de pasta inexistente", got, err)
		}
	})

	t.Run("limite - 1: última tentativa ainda é usada", func(t *testing.T) {
		withMaxAttempts(t, 3)
		dir := t.TempDir()
		touch(t, filepath.Join(dir, "foto.jpg"))
		touch(t, filepath.Join(dir, "foto (1).jpg"))
		got, err := Reserve(filepath.Join(dir, "foto.jpg"))
		assertReserved(t, got, err, filepath.Join(dir, "foto (2).jpg"))
	})

	t.Run("limite atingido devolve ErrNoAvailableName", func(t *testing.T) {
		withMaxAttempts(t, 3)
		dir := t.TempDir()
		touch(t, filepath.Join(dir, "foto.jpg"))
		touch(t, filepath.Join(dir, "foto (1).jpg"))
		touch(t, filepath.Join(dir, "foto (2).jpg"))
		_, err := Reserve(filepath.Join(dir, "foto.jpg"))
		if !errors.Is(err, ErrNoAvailableName) {
			t.Fatalf("err = %v, want ErrNoAvailableName", err)
		}
	})
}

func TestIsOccupied(t *testing.T) {
	dir := t.TempDir()
	existing := filepath.Join(dir, "existe")
	touch(t, existing)
	otherErr := errors.New("acesso negado")

	if !isOccupied(filepath.Join(dir, "qualquer"), fs.ErrExist) {
		t.Error("ErrExist deveria contar como ocupado")
	}
	if !isOccupied(existing, otherErr) {
		t.Error("outro erro com algo no caminho deveria contar como ocupado")
	}
	if isOccupied(filepath.Join(dir, "nada"), otherErr) {
		t.Error("outro erro sem nada no caminho não deveria contar como ocupado")
	}
}

func TestCloseReservation(t *testing.T) {
	path := filepath.Join(t.TempDir(), "reserva")
	f, err := os.Create(path)
	if err != nil {
		t.Fatal(err)
	}
	_ = f.Close()

	if err := closeReservation(f, path); err == nil {
		t.Fatal("fechar um arquivo já fechado deveria falhar")
	}
	if _, err := os.Stat(path); !errors.Is(err, fs.ErrNotExist) {
		t.Fatalf("a reserva deveria ter sido removida, stat err = %v", err)
	}
}

func assertReserved(t *testing.T, got string, err error, want string) {
	t.Helper()
	if err != nil {
		t.Fatalf("Reserve: %v", err)
	}
	if got != want {
		t.Fatalf("Reserve = %q, want %q", got, want)
	}
	info, statErr := os.Stat(got)
	if statErr != nil || info.Size() != 0 {
		t.Fatalf("o nome reservado deveria existir vazio: info=%v err=%v", info, statErr)
	}
}

func touch(t *testing.T, path string) {
	t.Helper()
	if err := os.WriteFile(path, []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
}

func withMaxAttempts(t *testing.T, n int) {
	t.Helper()
	old := maxAttempts
	maxAttempts = n
	t.Cleanup(func() { maxAttempts = old })
}
