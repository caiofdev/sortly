// Package apperr define os erros com código estável que o backend devolve ao
// frontend (ADR 0004). O frontend traduz o código; a mensagem em Error() é
// para logs e desenvolvedores.
package apperr

import "errors"

// CodeUnexpected é o código de qualquer erro sem código próprio.
const CodeUnexpected = "UNEXPECTED"

// Error é um erro identificado por um código estável. Use como sentinela
// (var ErrX = apperr.New(...)) e envolva com fmt.Errorf("%w: ...", ErrX).
type Error struct {
	code string
	msg  string
}

func New(code, msg string) *Error {
	return &Error{code: code, msg: msg}
}

func (e *Error) Error() string { return e.msg }

func (e *Error) Code() string { return e.code }

// CodeOf devolve o código do primeiro erro com código na cadeia de err, ou
// CodeUnexpected. Para err nil, devolve "".
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
