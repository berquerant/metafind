package meta

import (
	"os"
	"strings"
)

func NewDataFromEnviron() *Data {
	return NewDataFromEqualPairs(os.Environ())
}

func NewDataFromEqualPairs(ss []string) *Data {
	d := map[string]any{}
	for _, s := range ss {
		xs := strings.SplitN(s, "=", 2)
		if len(xs) != 2 {
			continue
		}
		d[xs[0]] = xs[1]
	}
	return &Data{
		d: d,
	}
}
