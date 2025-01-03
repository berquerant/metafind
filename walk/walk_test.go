package walk_test

import (
	"bytes"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/berquerant/metafind/expr"
	"github.com/berquerant/metafind/logx"
	"github.com/berquerant/metafind/walk"
	"github.com/stretchr/testify/assert"
)

func TestWalker(t *testing.T) {
	logx.Setup(os.Stderr, slog.LevelDebug)
	d := t.TempDir()

	touch := func(t *testing.T, p string) {
		f, err := os.Create(p)
		if err != nil {
			t.Error(err)
		}
		f.Close()
	}
	mkdir := func(t *testing.T, p string) {
		if err := os.MkdirAll(p, 0755); err != nil {
			t.Error(err)
		}
	}
	join := func(p ...string) string {
		return filepath.Join(append([]string{d}, p...)...)
	}

	// d
	//   d1/
	//   d2/
	//     f1
	//   d3/
	//     f2
	//     d31/
	//       f3
	var (
		d1  = join("d1")
		d2  = join("d2")
		f1  = join("d2", "f1")
		d3  = join("d3")
		f2  = join("d3", "f2")
		d31 = join("d3", "d31")
		f3  = join("d3", "d31", "f3")
	)
	t.Run("init", func(t *testing.T) {
		mkdir(t, d1)
		mkdir(t, d2)
		touch(t, f1)
		mkdir(t, d3)
		touch(t, f2)
		mkdir(t, d31)
		touch(t, f3)
	})

	t.Run("ReaderWalker", func(t *testing.T) {
		for _, tc := range []struct {
			name  string
			input []string
			want  []string
		}{
			{
				name: "exclude not exist",
				input: []string{
					f2 + "notexist",
					f1,
				},
				want: []string{
					f1,
				},
			},
			{
				name: "exclude dir but walk dir",
				input: []string{
					d1,
					d2,
					f2,
				},
				want: []string{
					f1,
					f2,
				},
			},
			{
				name: "a file",
				input: []string{
					f1,
				},
				want: []string{
					f1,
				},
			},
			{
				name: "empty",
				want: []string{},
			},
		} {
			t.Run(tc.name, func(t *testing.T) {
				r := bytes.NewBufferString(strings.Join(tc.input, "\n"))
				w := walk.NewReader(r, walk.NewFile(nil))
				result := slices.Collect(w.Walk(""))
				if !assert.Nil(t, w.Err()) {
					t.Errorf("%#v", w.Err())
				}

				got := make([]string, len(result))
				for i, x := range result {
					got[i] = x.Path()
				}

				slices.Sort(tc.want)
				slices.Sort(got)
				assert.Equal(t, tc.want, got)
			})
		}
	})

	t.Run("FileWalker", func(t *testing.T) {
		for _, tc := range []struct {
			name    string
			root    string
			exclude expr.Expr
			want    []string
		}{
			{
				name: "d31",
				root: d31,
				want: []string{
					f3,
				},
			},
			{
				name: "d3",
				root: d3,
				want: []string{
					f2,
					f3,
				},
			},
			{
				name: "d2",
				root: d2,
				want: []string{
					f1,
				},
			},
			{
				name: "d",
				root: d,
				want: []string{
					f1,
					f2,
					f3,
				},
			},
			{
				name:    "d exclude d3",
				root:    d,
				exclude: expr.New(expr.MustNewRaw(fmt.Sprintf(`path == "%s"`, d3))),
				want: []string{
					f1,
				},
			},
			{
				name:    "d exclude f3",
				root:    d,
				exclude: expr.New(expr.MustNewRaw(fmt.Sprintf(`path == "%s"`, f3))),
				want: []string{
					f1,
					f2,
				},
			},
			{
				name: "d1",
				root: d1,
				want: []string{},
			},
		} {
			t.Run(tc.name, func(t *testing.T) {
				w := walk.NewFile(tc.exclude)
				r := slices.Collect(w.Walk(tc.root))
				if !assert.Nil(t, w.Err()) {
					t.Errorf("%#v", w.Err())
				}

				got := make([]string, len(r))
				for i, x := range r {
					got[i] = x.Path()
				}

				slices.Sort(tc.want)
				slices.Sort(got)
				assert.Equal(t, tc.want, got)
			})
		}
	})
}
