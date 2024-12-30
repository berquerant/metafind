package iox

import (
	"context"
	"errors"

	"github.com/berquerant/metafind/syncx"
	"github.com/berquerant/metafind/walk"
)

const walkerBufferSize = 100

type Walker struct {
	roots  []string
	walker walk.Walker
	err    error
}

func NewWalker(walker walk.Walker, root ...string) *Walker {
	return &Walker{
		walker: walker,
		roots:  root,
	}
}

func (w *Walker) Err() error { return w.err }

func (w *Walker) Start(ctx context.Context) <-chan walk.Entry {
	entryC := make(chan walk.Entry, walkerBufferSize)

	go func() {
		defer close(entryC)

		var errs []error
		for _, root := range w.roots {
			if syncx.Done(ctx) {
				break
			}
			for e := range w.walker.Walk(root) {
				if syncx.Done(ctx) {
					break
				}
				entryC <- e
			}
			if err := w.walker.Err(); err != nil {
				errs = append(errs, err)
			}
		}

		w.err = errors.Join(errs...)
	}()

	return entryC
}
