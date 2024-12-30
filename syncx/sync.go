package syncx

import (
	"context"
	"errors"
)

func Done(ctx context.Context) bool {
	select {
	case <-ctx.Done():
		return true
	default:
		return false
	}
}

func IsDone(err error) bool {
	return errors.Is(err, context.Canceled)
}
