package iox

import (
	"fmt"
	"io"
	"log/slog"
	"os"
	"strings"
)

func Open(file ...string) ([]*os.File, error) {
	fs := make([]*os.File, len(file))

	for i, x := range file {
		f, err := os.Open(x)
		if err != nil {
			for j := range i {
				fs[j].Close()
			}
			return nil, err
		}
		fs[i] = f
	}

	return fs, nil
}

func ReadFileOrLiteral(s string) (string, error) {
	if !strings.HasPrefix(s, FileMark) {
		slog.Debug("ReadFileOrLiteral Literal", slog.String("s", s))
		return s, nil
	}

	fileName := s[1:]
	slog.Debug("ReadFileOrLiteral File", slog.String("file", fileName))
	f, err := os.Open(fileName)
	if err != nil {
		return "", fmt.Errorf("%w: path %s", err, s)
	}
	defer f.Close()
	b, err := io.ReadAll(f)
	if err != nil {
		return "", fmt.Errorf("%w: path %s", err, s)
	}
	return string(b), nil
}
