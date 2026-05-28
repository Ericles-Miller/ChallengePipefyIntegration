package AppError

import "errors"

var (
	ErrNotFound       = errors.New("not found")
	ErrBadRequest     = errors.New("bad request")
	ErrUnauthorized   = errors.New("unauthorized")
	ErrForbidden      = errors.New("forbidden")
	ErrInternalServer = errors.New("internal server error")
)

type appError struct {
	msg  string
	kind error
}

func (e *appError) Error() string { return e.msg }
func (e *appError) Unwrap() error { return e.kind }

func New(msg string, kind error) error {
	return &appError{msg: msg, kind: kind}
}