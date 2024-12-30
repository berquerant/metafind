package metric_test

import (
	"sync"
	"testing"

	"github.com/berquerant/metafind/metric"
	"github.com/stretchr/testify/assert"
)

func TestCounter(t *testing.T) {
	const name = "cnt"
	c := metric.NewCounter(name)
	assert.Equal(t, name, c.Name())
	assert.Equal(t, uint64(0), c.Get())
	c.Incr()
	assert.Equal(t, uint64(1), c.Get())

	const (
		n = 999
		p = 4
	)
	var (
		wg     sync.WaitGroup
		inputC = make(chan int, n)
	)

	for range p {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for range inputC {
				c.Incr()
			}
		}()
	}

	for i := range n {
		inputC <- i
	}
	close(inputC)
	wg.Wait()
	assert.Equal(t, uint64(1+n), c.Get())
}
