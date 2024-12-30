package iox_test

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/berquerant/metafind/iox"
	"github.com/stretchr/testify/assert"
)

func TestReaderAndCloser(t *testing.T) {
	t.Run("file", func(t *testing.T) {
		const content = "CONTENT"
		filename := filepath.Join(t.TempDir(), "some.txt")
		{
			f, err := os.Create(filename)
			if err != nil {
				t.Error(err)
			}
			fmt.Fprint(f, content)
			f.Close()
		}

		r, err := iox.NewReaderAndCloser(nil, filename)
		assert.Nil(t, err)
		got, err := io.ReadAll(r.Reader())
		assert.Nil(t, err)
		assert.Nil(t, r.Close())
		assert.Equal(t, content, string(got))
	})

	t.Run("stdin", func(t *testing.T) {
		const content = "STDINCONTENT"
		stdin := bytes.NewBufferString(content)

		r, err := iox.NewReaderAndCloser(stdin, "-")
		assert.Nil(t, err)
		got, err := io.ReadAll(r.Reader())
		assert.Nil(t, err)
		assert.Nil(t, r.Close())
		assert.Equal(t, content, string(got))
	})

	t.Run("exclusive stdin and file", func(t *testing.T) {
		_, err := iox.NewReaderAndCloser(nil, "-", "invalid")
		assert.ErrorIs(t, err, iox.ErrIO)
	})
}
