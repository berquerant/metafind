package iox

import "errors"

const (
	StdinMark  = "-"
	StdoutMark = "-"
	FileMark   = "@"
)

var (
	ErrIO = errors.New("IO")
)
