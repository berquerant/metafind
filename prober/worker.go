package prober

import (
	"context"

	"github.com/berquerant/metafind/meta"
	"github.com/berquerant/metafind/walk"
	"github.com/berquerant/metafind/worker"
)

type Prober = meta.Prober
type Data = meta.Data
type Worker = worker.Worker[*Data, *Data]

// AddData add metadata obtained from Prober.
func AddData(ctx context.Context, name string, p Prober, x *Data) (*Data, error) {
	path := walk.GetPathFromMetadata(x)
	y, err := p.Probe(ctx, path)
	if err != nil {
		return nil, err
	}
	x.Set(name, y.Unwrap())
	return x, nil
}

func NewWorker(p Prober, n int, name string) *Worker {
	f := func(ctx context.Context, x *Data) (*Data, error) {
		return AddData(ctx, name, p, x)
	}
	return worker.New(name, n, f)
}
