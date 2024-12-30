package iox_test

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/berquerant/metafind/iox"
	"github.com/stretchr/testify/assert"
)

func TestReadFileOrLiteral(t *testing.T) {
	t.Run("literal", func(t *testing.T) {
		got, err := iox.ReadFileOrLiteral("LITERAL")
		assert.Nil(t, err)
		assert.Equal(t, "LITERAL", got)
	})
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

		got, err := iox.ReadFileOrLiteral("@" + filename)
		assert.Nil(t, err)
		assert.Equal(t, content, got)
	})
}
