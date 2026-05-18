package errors

import (
	"fmt"
)

type K8strikeRuntimeError struct {
	Err       error
	CustomMsg string
}

func New(text string) error {
	return &K8strikeRuntimeError{nil, text}
}

func (e *K8strikeRuntimeError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s:\n%s", e.CustomMsg, e.Err)
	} else {
		return e.CustomMsg
	}
}
