package metric

import "sync/atomic"

func incr(addr *uint64) {
	atomic.AddUint64(addr, 1)
}

// Counter counts up uint value.
type Counter struct {
	name  string
	value *uint64
}

func NewCounter(name string) *Counter {
	var v uint64
	return &Counter{
		name:  name,
		value: &v,
	}
}

// Incr increments the value by an atomic operation.
func (c *Counter) Incr()       { incr(c.value) }
func (c Counter) Name() string { return c.name }
func (c *Counter) Get() uint64 { return *c.value }
