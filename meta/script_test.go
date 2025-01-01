package meta_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/berquerant/metafind/meta"
	"github.com/stretchr/testify/assert"
)

func TestScript(t *testing.T) {
	for _, tc := range []struct {
		title string
		raw   string
		want  *meta.Data
		err   error
	}{
		{
			title: "all broken",
			raw:   `neither json nor equal pairs`,
			err:   meta.ErrParse,
		},
		{
			title: "equal pairs except broken",
			raw: `k=v
k2=v2=@ARG
broken`,
			want: meta.NewData(map[string]any{
				"k":  "v",
				"k2": `v2="DUMMY"`,
			}),
		},
		{
			title: "equal pair",
			raw:   `k=v`,
			want: meta.NewData(map[string]any{
				"k": "v",
			}),
		},
		{
			title: "json",
			raw: `{
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
}`,
			want: meta.NewData(map[string]any{
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
			}),
		},
	} {
		t.Run(tc.title, func(t *testing.T) {
			content := fmt.Sprintf(`cat <<EOS
%s
EOS`, tc.raw)
			s := meta.NewScript(content, "sh")
			defer s.Close()
			got, err := s.Probe(context.TODO(), "DUMMY")
			if tc.err != nil {
				assert.ErrorIs(t, err, tc.err)
				return
			}
			if !assert.Nil(t, err) {
				t.Logf("ERR:%v", err)
				return
			}
			assert.Equal(t, tc.want.Unwrap(), got.Unwrap())
		})
	}
}
