package organizer

import "github.com/caiofdev/sortly/backend/apperr"

var (
	ErrInvalidSource      = apperr.New("INVALID_SOURCE", "organizer: pasta de origem inválida")
	ErrInvalidDestination = apperr.New("INVALID_DESTINATION", "organizer: pasta de destino inválida")
	ErrNoCriteria         = apperr.New("NO_CRITERIA", "organizer: nenhum critério de organização selecionado")
	// Os arquivos foram movidos, mas o registro para desfazer não pôde ser gravado (#8).
	ErrRecordNotSaved = apperr.New("RECORD_NOT_SAVED", "organizer: registro da organização não foi salvo")
)
