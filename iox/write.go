package iox

import (
	"io"
	"os"
)

type psuedoWriteCloser struct {
	w io.Writer
}

func AsWriteCloser(w io.Writer) io.WriteCloser {
	return &psuedoWriteCloser{
		w: w,
	}
}

func (w *psuedoWriteCloser) Write(p []byte) (int, error) { return w.w.Write(p) }
func (psuedoWriteCloser) Close() error                   { return nil }

func NewWriteCloser(stdout io.Writer, file string) (io.WriteCloser, error) {
	if file == StdoutMark {
		return AsWriteCloser(stdout), nil
	}
	return os.OpenFile(file, os.O_WRONLY|os.O_CREATE, 0644)
}
