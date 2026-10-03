package organizer

import (
	"errors"

	"github.com/caiofdev/sortly/internal/apperr"
)

// Erros com código devolvidos ao frontend (ADR 0004).
var (
	ErrInvalidSource      = apperr.New("INVALID_SOURCE", "organizer: pasta de origem inválida")
	ErrInvalidDestination = apperr.New("INVALID_DESTINATION", "organizer: pasta de destino inválida")
	ErrNoCriteria         = apperr.New("NO_CRITERIA", "organizer: nenhum critério de organização selecionado")
	// ErrRecordNotSaved: os arquivos foram movidos, mas o registro para
	// desfazer não pôde ser gravado.
	ErrRecordNotSaved = apperr.New("RECORD_NOT_SAVED", "organizer: registro da organização não foi salvo")
)

// ErrSkipNoExtension sinaliza que o arquivo não tem extensão e o critério de
// extensão está ligado: ele não é movido (conta em IgnoredWithoutExtension).
var ErrSkipNoExtension = errors.New("organizer: arquivo sem extensão")
