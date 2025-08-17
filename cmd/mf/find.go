package main

import (
	"context"
	"errors"
	"log/slog"

	"github.com/berquerant/metafind/logx"
	"github.com/berquerant/metafind/meta"
)

func find(ctx context.Context, c *Config) error {
	w, err := c.NewOutput()
	if err != nil {
		return err
	}
	defer w.Close()

	expression, err := c.NewExpr()
	if err != nil && !errors.Is(err, errNotSpecified) {
		return err
	}
	exprEnabled := err == nil
	slog.Debug("Expr", slog.Bool("enabled", exprEnabled))

	var (
		inC  = make(chan *meta.Data, bufferSize)
		outC = make(chan *meta.Data, bufferSize)
	)
	{
		r, err := c.NewIndexReader()
		switch {
		case err == nil:
			defer r.Close()
			rd := meta.NewReader(r.Reader())
			inC = rd.Read(ctx)
			defer func() {
				if err := rd.Err(); err != nil {
					slog.Warn("IndexReader", logx.Err(err))
				}
			}()
		case errors.Is(err, errNotSpecified):
			walker, err := c.NewRootWalker()
			if err != nil {
				return err
			}
			entryC := walker.Start(ctx)
			entryWorker := c.NewEntryWorker()
			entryWorker.Start(ctx, entryC, inC)
			defer func() {
				if err := walker.Err(); err != nil {
					slog.Warn("RootWalker", logx.Err(err))
				}
			}()
		default:
			return err
		}
	}

	join, err := c.NewProberWorkersChain()
	if err != nil {
		return err
	}
	join.Start(ctx, inC, outC)

	for x := range outC {
		if !exprEnabled {
			c.Output(w, x)
			continue
		}

		passed, err := expression.Run(x.Unwrap())
		if err != nil {
			slog.Debug("Expr error", logx.JSON("data", x.Unwrap()), logx.Err(err))
			continue
		}
		if !passed {
			continue
		}
		c.Output(w, x)
	}

	return nil
}
