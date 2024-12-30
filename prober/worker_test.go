package prober_test

import (
	"context"
	"errors"
	"testing"

	"github.com/berquerant/metafind/meta"
	"github.com/berquerant/metafind/prober"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type mockProber struct {
	mock.Mock
}

func (p *mockProber) Probe(ctx context.Context, path string) (*meta.Data, error) {
	args := p.Called(ctx, path)
	return args.Get(0).(*meta.Data), args.Error(1)
}

func TestAddData(t *testing.T) {
	const (
		path = "PATH"
		name = "NAME"
	)
	var (
		d = meta.NewData(map[string]any{
			"path": path,
		})
		someErr = errors.New("some error")
	)

	for _, tc := range []struct {
		title   string
		data    *meta.Data
		err     error
		want    *meta.Data
		wantErr error
	}{
		{
			title: "probe success",
			data: meta.NewData(map[string]any{
				"v": 1,
			}),
			want: meta.NewData(map[string]any{
				"path": path,
				name: map[string]any{
					"v": 1,
				},
			}),
		},
		{
			title: "no data",
			data:  meta.NewData(nil),
			want:  d,
		},
		{
			title:   "probe error",
			err:     someErr,
			wantErr: someErr,
		},
	} {
		t.Run(tc.title, func(t *testing.T) {
			p := new(mockProber)
			p.On("Probe", context.TODO(), path).Return(tc.data, tc.err)
			got, err := prober.AddData(context.TODO(), name, p, d)
			if tc.wantErr != nil {
				assert.ErrorIs(t, err, tc.wantErr)
				return
			}
			assert.Nil(t, err)
			assert.Equal(t, tc.want, got)
		})
	}
}
