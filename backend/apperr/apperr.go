// Package apperr define os erros com código estável que o backend devolve ao
// frontend. O frontend traduz o código; a mensagem de Error() é para logs e
// desenvolvedores (ADR 0004, #8).
package apperr

import "errors"

const CodeUnexpected = "UNEXPECTED"

// Use como sentinela (var ErrX = apperr.New(...)) e envolva com fmt.Errorf("%w: ...", ErrX)
// (#8).
type Error struct {
	code string
	msg  string
}

func New(code, msg string) *Error {
	return &Error{code: code, msg: msg}
}

func (e *Error) Error() string { return e.msg }

func (e *Error) Code() string { return e.code }

// Procura o primeiro erro com código na cadeia; sem nenhum, CodeUnexpected.
// Para err nil, devolve "" (#8).
func CodeOf(err error) string {
	if err == nil {
		return ""
	}
	var coded interface{ Code() string }
	if errors.As(err, &coded) {
		return coded.Code()
	}
	return CodeUnexpected
}
