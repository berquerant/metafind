package meta_test

import (
	"context"
	"testing"

	"github.com/berquerant/metafind/meta"
	"github.com/stretchr/testify/assert"
)

func TestScript(t *testing.T) {
	const (
		content = `cat <<EOS
{
  "b": true,
  "i": 1,
  "f": 1.2,
  "s": "str",
  "m": {
    "x": 10,
    "y": false
  },
  "a": [
    "x",
    true
  ],
  "arg": @ARG
}
EOS`
		shell = "sh"
	)
	want := map[string]any{
		"b": true,
		"i": float64(1),
		"f": 1.2,
		"s": "str",
		"m": map[string]any{
			"x": float64(10),
			"y": false,
		},
		"a": []any{
			"x",
			true,
		},
		"arg": "DUMMY",
	}

	s := meta.NewScript(content, shell)
	defer s.Close()
	got, err := s.Probe(context.TODO(), "DUMMY")
	if !assert.Nil(t, err) {
		t.Logf("ERR:%v", err)
		return
	}
	assert.Equal(t, want, got.Unwrap())
}
