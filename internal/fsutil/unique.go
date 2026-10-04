package fsutil

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

// ErrNoAvailableName indica que todas as tentativas de nome estavam ocupadas.
var ErrNoAvailableName = errors.New("fsutil: nenhum nome disponível para o destino")

// maxAttempts limita a busca por nome livre: o próprio nome mais
// "nome (1)" ... "nome (9999)". É variável para os testes de limite.
var maxAttempts = 10000

// Reserve escolhe o primeiro nome livre a partir de path ("nome.ext",
// "nome (1).ext", "nome (2).ext", ...) e o reserva criando um arquivo vazio
// de forma exclusiva (O_EXCL). Assim, dois movimentos simultâneos nunca
// recebem o mesmo nome. O arquivo reservado deve ser substituído com Move.
func Reserve(path string) (string, error) {
	for n := range maxAttempts {
		candidate := candidateName(path, n)

		f, err := os.OpenFile(nativePath(candidate), os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0o644)
		if err == nil {
			return candidate, closeReservation(f, candidate)
		}
		if !isOccupied(candidate, err) {
			return "", err
		}
	}
	return "", fmt.Errorf("%w: %s", ErrNoAvailableName, path)
}

// candidateName devolve a n-ésima tentativa de nome: n = 0 é o próprio path.
// A extensão segue Ext, então ".gitignore" vira ".gitignore (1)".
func candidateName(path string, n int) string {
	if n == 0 {
		return path
	}
	dir, base := filepath.Split(path)
	ext := Ext(base)
	stem := strings.TrimSuffix(base, ext)
	return filepath.Join(dir, fmt.Sprintf("%s (%d)%s", stem, n, ext))
}

// isOccupied trata como "nome ocupado" tanto o erro de arquivo existente quanto
// os casos em que o sistema devolve outro erro, mas algo já existe no caminho
// (no Windows, criar um arquivo com o nome de uma pasta retorna acesso negado).
func isOccupied(path string, err error) bool {
	if errors.Is(err, fs.ErrExist) {
		return true
	}
	_, statErr := os.Lstat(nativePath(path))
	return statErr == nil
}

func closeReservation(f *os.File, path string) error {
	if err := f.Close(); err != nil {
		_ = os.Remove(nativePath(path))
		return err
	}
	return nil
}
