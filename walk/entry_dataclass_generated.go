// Code generated by "dataclass -type Entry -field Path string|Info fs.FileInfo|Zip ZipEntry -output entry_dataclass_generated.go"; DO NOT EDIT.

package walk

import "io/fs"

type Entry interface {
	Path() string
	Info() fs.FileInfo
	Zip() ZipEntry
}
type entry struct {
	path string
	info fs.FileInfo
	zip  ZipEntry
}

func (s *entry) Path() string      { return s.path }
func (s *entry) Info() fs.FileInfo { return s.info }
func (s *entry) Zip() ZipEntry     { return s.zip }
func NewEntry(
	path string,
	info fs.FileInfo,
	zip ZipEntry,
) Entry {
	return &entry{
		path: path,
		info: info,
		zip:  zip,
	}
}
