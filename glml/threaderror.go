// Copyright © 2012 Popog
package glml

type ThreadError interface {
	error
	Fatal() bool // returns true if the error is fatal
}

type threadError struct {
	error
	fatal bool
}

func (c threadError) Fatal() bool { return c.fatal }

func NewThreadError(err error, fatal bool) ThreadError {
	return threadError{
		error: err,
		fatal: fatal,
	}
}
