package meta

import (
	"bufio"
	"context"
	"encoding/json"
	"io"
	"log/slog"

	"github.com/berquerant/metafind/logx"
	"github.com/berquerant/metafind/syncx"
)

// Reader parses Data jsonl.
type Reader struct {
	r   io.Reader
	err error
}

func NewReader(r io.Reader) *Reader {
	return &Reader{
		r: r,
	}
}

func (r Reader) Err() error { return r.err }

const readerBufferSize = 100

func (r *Reader) Read(ctx context.Context) chan *Data {
	r.err = nil
	resultC := make(chan *Data, readerBufferSize)

	go func() {
		defer close(resultC)

		if syncx.Done(ctx) {
			return
		}

		scanner := bufio.NewScanner(r.r)
		for scanner.Scan() {
			select {
			case <-ctx.Done():
				return
			default:
				d := map[string]any{}
				if err := json.Unmarshal(scanner.Bytes(), &d); err != nil {
					slog.Warn("MetaReader: unmarshal", logx.Err(err))
					continue
				}
				resultC <- NewData(d)
			}
		}

		if err := scanner.Err(); err != nil {
			r.err = err
		}
	}()

	return resultC
}
