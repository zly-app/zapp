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

func (s jsonSerializer) MarshalBytes(a interface{}) ([]byte, error) {
	return json.Marshal(a)
}

func (jsonSerializer) Unmarshal(r io.Reader, a interface{}) error {
	return json.NewDecoder(r).Decode(a)
}

func (s jsonSerializer) UnmarshalBytes(data []byte, a interface{}) error {
	return json.Unmarshal(data, a)
}
