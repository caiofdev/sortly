package files

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/caiofdev/sortly/backend/fs/paths"
)

var ErrNoAvailableName = errors.New("files: nenhum nome disponível para o destino")

// O próprio nome mais "nome (1)" ... "nome (9999)". É variável para os testes de limite
// (#5).
var maxAttempts = 10000

// A reserva cria um arquivo vazio com O_EXCL, então dois movimentos simultâneos
// nunca recebem o mesmo nome ("nome.ext", "nome (1).ext", ...). O arquivo
// reservado deve ser substituído com Move (#5).
func Reserve(path string) (string, error) {
	for n := range maxAttempts {
		candidate := candidateName(path, n)

		f, err := os.OpenFile(paths.Native(candidate), os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0o644)
		if err == nil {
			return candidate, closeReservation(f, candidate)
		}
		if !isOccupied(candidate, err) {
			return "", err
		}
	}
	return "", fmt.Errorf("%w: %s", ErrNoAvailableName, path)
}

// n = 0 é o próprio path. A extensão segue Ext, então ".gitignore" vira ".gitignore (1)"
// (#5).
func candidateName(path string, n int) string {
	if n == 0 {
		return path
	}
	dir, base := filepath.Split(path)
	ext := paths.Ext(base)
	stem := strings.TrimSuffix(base, ext)
	return filepath.Join(dir, fmt.Sprintf("%s (%d)%s", stem, n, ext))
}

// No Windows, criar um arquivo com o nome de uma pasta devolve acesso negado,
// não "já existe": qualquer erro com algo no caminho conta como ocupado (#5).
func isOccupied(path string, err error) bool {
	if errors.Is(err, fs.ErrExist) {
		return true
	}
	_, statErr := os.Lstat(paths.Native(path))
	return statErr == nil
}

func closeReservation(f *os.File, path string) error {
	if err := f.Close(); err != nil {
		_ = os.Remove(paths.Native(path))
		return err
	}
	return nil
}
