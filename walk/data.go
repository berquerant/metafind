package walk

import (
	"fmt"
	"iter"
	"path/filepath"
	"strings"
	"time"

	"github.com/berquerant/metafind/meta"
)

//go:generate go run github.com/berquerant/dataclass -type Entry -field "Path string|Info fs.FileInfo" -output entry_dataclass_generated.go

type Walker interface {
	Walk(root string) iter.Seq[Entry]
	Err() error
}

const (
	walkerBufferSize = 100
)

func NewMetaData(entry Entry) *meta.Data {
	var (
		path = entry.Path()
		name = entry.Info().Name()
		ext  = filepath.Ext(name)
	)
	return meta.NewData(map[string]any{
		"path":        path,
		"dir":         filepath.Dir(path),
		"name":        name,
		"ext":         ext,
		"basename":    strings.TrimRight(name, ext),
		"basepath":    strings.TrimRight(path, ext),
		"size":        entry.Info().Size(),
		"mode":        fmt.Sprintf("%o", entry.Info().Mode()),
		"mod_time":    entry.Info().ModTime().Format(time.DateTime),
		"mod_time_ts": entry.Info().ModTime().Unix(),
	})
}

func GetPathFromMetadata(v *meta.Data) string {
	x, _ := v.Get("path")
	return x.(string)
}
