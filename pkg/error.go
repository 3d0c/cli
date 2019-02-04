package cli

import (
	"errors"
)

const (
	ErrHelp = iota
)

type Error struct {
	Code int
	Msg  error
}

func (e *Error) Error() string {
	return e.Msg.Error()
}

func ErrFlagRequired(flagname string) error {
	return &Error{
		Code: ErrHelp,
		Msg:  errors.New("-" + flagname + " is required"),
	}
}

func ErrWrongFormat(flagname string) error {
	return &Error{
		Code: ErrHelp,
		Msg:  errors.New("-" + flagname + " has wrong format"),
	}
}
