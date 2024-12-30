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

func TestWriteCloser(t *testing.T) {
	const content = "CONTENT"
	t.Run("stdout", func(t *testing.T) {
		var stdout bytes.Buffer
		w, err := iox.NewWriteCloser(&stdout, "-")
		assert.Nil(t, err)
		fmt.Fprint(w, content)
		assert.Nil(t, w.Close())
		assert.Equal(t, content, stdout.String())
	})

	t.Run("file", func(t *testing.T) {
		filename := filepath.Join(t.TempDir(), "out.txt")
		w, err := iox.NewWriteCloser(nil, filename)
		assert.Nil(t, err)
		fmt.Fprint(w, content)
		assert.Nil(t, w.Close())

		f, err := os.Open(filename)
		assert.Nil(t, err)
		b, err := io.ReadAll(f)
		assert.Nil(t, err)
		assert.Equal(t, content, string(b))
	})
}
