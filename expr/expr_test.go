package expr_test

import (
	"testing"

	"github.com/berquerant/metafind/expr"
	"github.com/stretchr/testify/assert"
)

func TestProgram(t *testing.T) {
	for _, tc := range []struct {
		title      string
		code       string
		env        map[string]any
		compileErr bool
		runtimeErr bool
		want       bool
	}{
		{
			title: "no env const 1",
			code:  `1`,
			want:  true,
		},
		{
			title: "no env const 0",
			code:  `0`,
			want:  false,
		},
		{
			title: "env 1",
			code:  `v`,
			env: map[string]any{
				"v": 1,
			},
			want: true,
		},
		{
			title: "op err",
			code:  `x + y`,
			env: map[string]any{
				"x": 1,
				"y": "one",
			},
			runtimeErr: true,
		},
		{
			title: "nested env",
			code:  `v.x`,
			env: map[string]any{
				"v": map[string]any{
					"x": 1,
				},
			},
			want: true,
		},
	} {
		t.Run(tc.title, func(t *testing.T) {
			r, err := expr.NewRaw(tc.code)
			if tc.compileErr {
				assert.NotNil(t, err)
				return
			}
			if !assert.Nil(t, err) {
				return
			}
			p := expr.New(r)
			got, err := p.Run(tc.env)
			if tc.runtimeErr {
				assert.NotNil(t, err)
				return
			}
			assert.Nil(t, err)
			assert.Equal(t, tc.want, got)
		})
	}
}

func TestAsBool(t *testing.T) {
	truthy := []any{
		true,
		1,
		-1,
		uint(1),
		1.1,
		float64(1.1),
		"x",
		map[string]any{"k": 1},
		[]string{"e"},
		[1]string{"e"},
	}
	falsy := []any{
		nil,
		false,
		0,
		uint(0),
		0.0,
		float64(0.0),
		"",
		map[string]any{},
		[]string{},
		[0]string{},
		struct{}{},
	}

	t.Run("true", func(t *testing.T) {
		for _, v := range truthy {
			assert.True(t, expr.AsBool(v), "%#v", v)
		}
	})
	t.Run("false", func(t *testing.T) {
		for _, v := range falsy {
			assert.False(t, expr.AsBool(v), "%#v", v)
		}
	})
}
