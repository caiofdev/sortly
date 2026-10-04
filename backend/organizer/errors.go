package organizer

import "github.com/caiofdev/sortly/backend/apperr"

// Erros com código devolvidos ao frontend (ADR 0004).
var (
	ErrInvalidSource      = apperr.New("INVALID_SOURCE", "organizer: pasta de origem inválida")
	ErrInvalidDestination = apperr.New("INVALID_DESTINATION", "organizer: pasta de destino inválida")
	ErrNoCriteria         = apperr.New("NO_CRITERIA", "organizer: nenhum critério de organização selecionado")
	// ErrRecordNotSaved: os arquivos foram movidos, mas o registro para
	// desfazer não pôde ser gravado.
	ErrRecordNotSaved = apperr.New("RECORD_NOT_SAVED", "organizer: registro da organização não foi salvo")
)
