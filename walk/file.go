package walk

import (
	"context"
	"io/fs"
	"iter"
	"log/slog"
	"path/filepath"

	"github.com/berquerant/metafind/logx"
	"github.com/berquerant/metafind/metric"
)

var _ Walker = &FileWalker{}

func NewFile() *FileWalker {
	return &FileWalker{}
}

// FileWalker walks only files under the root.
type FileWalker struct {
	err error
}

func (w FileWalker) Err() error { return w.err }

var (
	WalkCount      = metric.NewCounter("Walk")
	WalkCallCount  = metric.NewCounter("WalkCall")
	WalkDirCount   = metric.NewCounter("WalkDir")
	WalkEntryCount = metric.NewCounter("WalkEntry")
)

func (w *FileWalker) Walk(root string) iter.Seq[Entry] {
	WalkCount.Incr()
	w.err = nil

	return func(yield func(Entry) bool) {
		resultC := make(chan Entry, walkerBufferSize)
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		go func() {
			defer close(resultC)

			w.err = filepath.Walk(root, func(path string, info fs.FileInfo, err error) error {
				slog.Debug("FileWalker", slog.String("path", path), logx.Err(err))
				WalkCallCount.Incr()

				select {
				case <-ctx.Done():
					return filepath.SkipAll
				default:
					if err != nil {
						return err
					}
					if info.IsDir() {
						WalkDirCount.Incr()
						// skip dir
						return nil
					}

					WalkEntryCount.Incr()
					resultC <- NewEntry(path, info)
					return nil
				}
			})
		}()

		for x := range resultC {
			if !yield(x) {
				return
			}
		}
	}
}
