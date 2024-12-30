package worker

import (
	"context"
)

type Chain[T any] struct {
	workers []*Worker[T, T]
	n       int
}

func NewChain[T any](workers []*Worker[T, T], n int) *Chain[T] {
	return &Chain[T]{
		workers: workers,
		n:       n,
	}
}

func (j *Chain[T]) Start(ctx context.Context, inC <-chan T, outC chan<- T) {
	switch len(j.workers) {
	case 0:
		w := New[T, T]("Noop", j.n, func(_ context.Context, x T) (T, error) {
			return x, nil
		})
		j.workers = []*Worker[T, T]{w}
		w.Start(ctx, inC, outC)
	case 1:
		j.workers[0].Start(ctx, inC, outC)
	default:
		var (
			ic = inC
			oc chan T
		)
		for i, w := range j.workers {
			if i == len(j.workers)-1 {
				w.Start(ctx, ic, outC)
				return
			}
			oc = make(chan T, j.n)
			w.Start(ctx, ic, oc)
			ic = oc
		}
	}
}
