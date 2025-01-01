package meta_test

import (
	"testing"

	"github.com/berquerant/metafind/meta"
	"github.com/stretchr/testify/assert"
)

func TestEnv(t *testing.T) {
	t.Run("NewDataFromEqualPairs", func(t *testing.T) {
		for _, tc := range []struct {
			title string
			ss    []string
			want  *meta.Data
		}{
			{
				title: "ignore broken pairs",
				ss: []string{
					"k=v",
					"k2=v2=",
					"broken",
					"left=",
				},
				want: meta.NewData(map[string]any{
					"k":    "v",
					"k2":   "v2=",
					"left": "",
				}),
			},
			{
				title: "pair",
				ss: []string{
					"k=v",
				},
				want: meta.NewData(map[string]any{
					"k": "v",
				}),
			},
			{
				title: "empty",
				want:  meta.NewData(map[string]any{}),
			},
		} {
			t.Run(tc.title, func(t *testing.T) {
				got := meta.NewDataFromEqualPairs(tc.ss)
				assert.Equal(t, tc.want.Unwrap(), got.Unwrap())
			})
		}
	})
}
