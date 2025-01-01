package meta

import (
	"encoding/json"
)

type Data struct {
	d map[string]any
}

func NewData(d map[string]any) *Data {
	return &Data{
		d: d,
	}
}

func (d Data) IsEmpty() bool                { return len(d.d) == 0 }
func (d Data) Unwrap() map[string]any       { return d.d }
func (d Data) MarshalJSON() ([]byte, error) { return json.Marshal(d.d) }
func (d Data) MarshalYAML() (any, error)    { return d.d, nil }
func (d *Data) Set(key string, value any)   { d.d[key] = value }
func (d Data) Get(key string) (any, bool) {
	v, ok := d.d[key]
	return v, ok
}
