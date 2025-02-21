package walk

import (
	"fmt"
	"iter"
	"path/filepath"
	"strings"
	"time"

	"github.com/berquerant/metafind/meta"
)

//go:generate go tool dataclass -type Entry -field "Path string|Info fs.FileInfo|Zip ZipEntry" -output entry_dataclass_generated.go
//go:generate go tool dataclass -type ZipEntry -field "Root string|RelPath string|CompressedSize uint64|UncompressedSize uint64|Comment string|NonUTF8 bool" -output zipentry_dataclass_generated.go

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
	data := meta.NewData(map[string]any{
		"path":        path,
		"dir":         filepath.Dir(path),
		"name":        name,
		"ext":         ext,
		"basename":    strings.TrimSuffix(name, ext),
		"basepath":    strings.TrimSuffix(path, ext),
		"size":        entry.Info().Size(),
		"mode":        fmt.Sprintf("%o", entry.Info().Mode()),
		"mod_time":    entry.Info().ModTime().Format(time.DateTime),
		"mod_time_ts": entry.Info().ModTime().Unix(),
	})
	data.Merge(newZipMetadata(entry.Zip()))
	return data
}

func newZipMetadata(entry ZipEntry) *meta.Data {
	if entry == nil {
		return nil
	}
	return meta.NewData(map[string]any{
		"root":              entry.Root(),
		"relpath":           entry.RelPath(),
		"compressed_size":   entry.CompressedSize(),
		"uncompressed_size": entry.UncompressedSize(),
		"comment":           entry.Comment(),
		"non_utf8":          entry.NonUTF8(),
	})
}

func GetPathFromMetadata(v *meta.Data) string {
	x, _ := v.Get("path")
	return x.(string)
}
