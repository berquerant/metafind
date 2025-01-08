package walk

import (
	"archive/zip"
	"context"
	"iter"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/berquerant/metafind/expr"
	"github.com/berquerant/metafind/syncx"
)

var _ Walker = &ZipWalker{}

func NewZip(exclude expr.Expr) *ZipWalker {
	return &ZipWalker{
		FileWalker: FileWalker{
			exclude: exclude,
		},
	}
}

type ZipWalker struct {
	FileWalker
}

func (w *ZipWalker) Walk(root string) iter.Seq[Entry] {
	WalkCount.Incr()
	w.err = nil
	root = os.ExpandEnv(root)

	return func(yield func(Entry) bool) {
		resultC := make(chan Entry, walkerBufferSize)
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		go func() {
			defer close(resultC)
			reader, err := zip.OpenReader(root)
			if err != nil {
				w.err = err
				return
			}
			defer reader.Close()

			for _, file := range reader.File {
				path := filepath.Join(root, file.Name)
				slog.Debug("ZipWalker", slog.String("root", root), slog.String("path", path))
				WalkCallCount.Incr()

				if syncx.Done(ctx) {
					return
				}

				entry := NewEntry(
					path,
					file.FileInfo(),
					NewZipEntry(
						root,
						file.Name,
						file.CompressedSize64,
						file.UncompressedSize64,
						file.Comment,
						file.NonUTF8,
					),
				)
				if w.isRejected(entry) {
					continue
				}
				if entry.Info().IsDir() {
					WalkDirCount.Incr()
					// skip dir
					continue
				}

				WalkEntryCount.Incr()
				resultC <- entry
			}
		}()

		for x := range resultC {
			if !yield(x) {
				return
			}
		}
	}
}
