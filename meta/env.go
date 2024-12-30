package meta

import (
	"os"
	"strings"
)

func NewDataFromEnviron() *Data {
	d := map[string]any{}
	for _, e := range os.Environ() {
		xs := strings.SplitN(e, "=", 2)
		k, v := xs[0], xs[1]
		d[k] = v
	}
	return NewData(d)
}
