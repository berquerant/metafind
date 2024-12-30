package worker

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"sync"

	"github.com/berquerant/metafind/logx"
	"github.com/berquerant/metafind/syncx"
)

func New[In, Out any](
	name string,
	n int,
	f func(context.Context, In) (Out, error),
) *Worker[In, Out] {
	if n < 1 {
		n = 1
	}
	var wg sync.WaitGroup
	return &Worker[In, Out]{
		name: name,
		n:    n,
		f:    f,
		wg:   &wg,
	}
}

type Worker[In, Out any] struct {
	name string
	n    int
	f    func(context.Context, In) (Out, error)
	wg   *sync.WaitGroup
}

var (
	// ErrReject makes Worker ignore the element.
	ErrReject = errors.New("Reject")
)

func (w *Worker[In, Out]) Start(
	ctx context.Context,
	inC <-chan In,
	outC chan<- Out,
) {
	for range w.n {
		w.wg.Add(1)
		go func() {
			defer w.wg.Done()
			if syncx.Done(ctx) {
				return
			}

			for x := range inC {
				if syncx.Done(ctx) {
					return
				}
				r, err := w.f(ctx, x)
				switch {
				case err == nil:
					outC <- r
				case errors.Is(err, ErrReject):
					continue
				case syncx.IsDone(err):
					return
				default:
					slog.Warn("Worker call",
						slog.String("name", w.name),
						slog.String("in", fmt.Sprintf("%v", x)),
						logx.Err(err),
					)
					// error but send original input
					if a, ok := any(x).(Out); ok {
						outC <- a
					} else {
						slog.Warn("Worker cannot send original input because IN type != OUT type",
							slog.String("in", fmt.Sprintf("%#v", x)),
						)
					}
				}
			}
		}()
	}

	go func() {
		w.wg.Wait()
		close(outC)
	}()
}
