package walk

import (
	"bufio"
	"context"
	"io"
	"iter"
	"log/slog"
	"os"

	"github.com/berquerant/metafind/logx"
	"github.com/berquerant/metafind/syncx"
)

var _ Walker = &ReaderWalker{}

// ReaderWalker receives paths from io.Reader and walks under them.
type ReaderWalker struct {
	r          io.Reader
	fileWalker Walker
	err        error
}

func NewReader(r io.Reader, fileWalker Walker) *ReaderWalker {
	return &ReaderWalker{
		r:          r,
		fileWalker: fileWalker,
	}
}

func (w ReaderWalker) Err() error { return w.err }

func (w *ReaderWalker) Walk(_ string) iter.Seq[Entry] {
	w.err = nil

	return func(yield func(Entry) bool) {
		resultC := make(chan Entry, walkerBufferSize)
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		go func() {
			defer close(resultC)

			if syncx.Done(ctx) {
				return
			}

			scanner := bufio.NewScanner(w.r)
			for scanner.Scan() {
				select {
				case <-ctx.Done():
					return
				default:
					path := os.ExpandEnv(scanner.Text())
					info, err := os.Stat(path)
					slog.Debug("ReaderWalker", slog.String("path", path), logx.Err(err))
					if os.IsNotExist(err) {
						continue
					}

					if err != nil {
						w.err = err
						return
					}
					if info.IsDir() {
						for x := range w.fileWalker.Walk(path) {
							resultC <- x
						}
						if err := w.fileWalker.Err(); err != nil {
							slog.Warn("ReaderWalker", slog.String("path", path), logx.Err(err))
						}
						continue
					}
					resultC <- NewEntry(path, info)
				}
			}

			if err := scanner.Err(); err != nil {
				w.err = err
			}
		}()

		for x := range resultC {
			if !yield(x) {
				return
			}
		}
	}
}
