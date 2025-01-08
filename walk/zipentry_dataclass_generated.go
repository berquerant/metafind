// Code generated by "dataclass -type ZipEntry -field Root string|RelPath string|CompressedSize uint64|UncompressedSize uint64|Comment string|NonUTF8 bool -output zipentry_dataclass_generated.go"; DO NOT EDIT.

package walk

type ZipEntry interface {
	Root() string
	RelPath() string
	CompressedSize() uint64
	UncompressedSize() uint64
	Comment() string
	NonUTF8() bool
}
type zipEntry struct {
	root             string
	relPath          string
	compressedSize   uint64
	uncompressedSize uint64
	comment          string
	nonUTF8          bool
}

func (s *zipEntry) Root() string             { return s.root }
func (s *zipEntry) RelPath() string          { return s.relPath }
func (s *zipEntry) CompressedSize() uint64   { return s.compressedSize }
func (s *zipEntry) UncompressedSize() uint64 { return s.uncompressedSize }
func (s *zipEntry) Comment() string          { return s.comment }
func (s *zipEntry) NonUTF8() bool            { return s.nonUTF8 }
func NewZipEntry(
	root string,
	relPath string,
	compressedSize uint64,
	uncompressedSize uint64,
	comment string,
	nonUTF8 bool,
) ZipEntry {
	return &zipEntry{
		root:             root,
		relPath:          relPath,
		compressedSize:   compressedSize,
		uncompressedSize: uncompressedSize,
		comment:          comment,
		nonUTF8:          nonUTF8,
	}
}
