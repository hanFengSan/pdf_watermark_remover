package pipeline

import "errors"

const (
	ExitInvalidInput = 2
	ExitProcessing   = 3
)

type ExitError struct {
	Code int
	Msg  string
}

func (e *ExitError) Error() string {
	return e.Msg
}

func AsExitError(err error) (*ExitError, bool) {
	if err == nil {
		return nil, false
	}
	var ee *ExitError
	if errors.As(err, &ee) {
		return ee, true
	}
	return nil, false
}
