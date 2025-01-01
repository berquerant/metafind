package main

import (
	"time"

	"github.com/berquerant/metafind/expr"
	"github.com/berquerant/metafind/meta"
	"github.com/berquerant/metafind/metric"
	"github.com/berquerant/metafind/walk"
)

func NewMetrics(duration time.Duration) any {
	list := []*metric.Counter{
		walk.WalkCount,
		walk.WalkCallCount,
		walk.WalkDirCount,
		walk.WalkEntryCount,
		walk.WalkExcludeCount,
		walk.WalkExcludeErrCount,
		expr.RunCount,
		expr.ErrCount,
		expr.TrueCount,
		expr.FalseCount,
		meta.ScriptCount,
		meta.ProbeCount,
		meta.ProbeSuccessCount,
		meta.ProbeFailureCount,
		AcceptCount,
	}

	d := map[string]any{
		"Duration":    duration.String(),
		"DurationSec": int(duration.Seconds()),
	}
	for _, x := range list {
		d[x.Name()] = x.Get()
	}
	return d
}
