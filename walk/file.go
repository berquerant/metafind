package walk

import (
	"context"
	"io/fs"
	"iter"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/berquerant/metafind/expr"
	"github.com/berquerant/metafind/logx"
	"github.com/berquerant/metafind/metric"
)

var _ Walker = &FileWalker{}

func NewFile(exclude expr.Expr) *FileWalker {
	return &FileWalker{
		exclude: exclude,
	}
}

// FileWalker walks only files under the root.
type FileWalker struct {
	err     error
	exclude expr.Expr
}

func (w FileWalker) Err() error { return w.err }

func (w *FileWalker) isRejected(entry Entry) bool {
	if w.exclude == nil {
		return false
	}

	data := NewMetaData(entry)
	data.Set("is_dir", entry.Info().IsDir())
	rejected, err := w.exclude.Run(data.Unwrap())
	if err != nil {
		WalkExcludeErrCount.Incr()
		slog.Warn("FileWalker: exclude", slog.String("path", entry.Path()), logx.Err(err))
		return true
	}
	if rejected {
		WalkExcludeCount.Incr()
	}
	return rejected
}

var (
	WalkCount           = metric.NewCounter("Walk")
	WalkCallCount       = metric.NewCounter("WalkCall")
	WalkDirCount        = metric.NewCounter("WalkDir")
	WalkEntryCount      = metric.NewCounter("WalkEntry")
	WalkExcludeCount    = metric.NewCounter("WalkExclude")
	WalkExcludeErrCount = metric.NewCounter("WalkExcludeErr")
)

func (w *FileWalker) Walk(root string) iter.Seq[Entry] {
	WalkCount.Incr()
	w.err = nil
	root = os.ExpandEnv(root)

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

					entry := NewEntry(path, info, nil)
					if w.isRejected(entry) {
						if info.IsDir() {
							return filepath.SkipDir
						}
						return nil
					}

					if info.IsDir() {
						WalkDirCount.Incr()
						// skip dir
						return nil
					}

					WalkEntryCount.Incr()
					resultC <- entry
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
