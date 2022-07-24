package serializer

import (
	"encoding/json"
	"io"
)

const JsonSerializerName = "json"

type jsonSerializer struct{}

func (jsonSerializer) Marshal(a interface{}, w io.Writer) error {
	return json.NewEncoder(w).Encode(a)
}

func (jsonSerializer) Unmarshal(in io.Reader, a interface{}) error {
	return json.NewDecoder(in).Decode(a)
}
