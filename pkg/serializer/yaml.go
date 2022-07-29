package serializer

import (
	"io"

	"gopkg.in/yaml.v3"
)

const YamlSerializerName = "yaml"

type yamlSerializer struct{}

func (yamlSerializer) Marshal(a interface{}, w io.Writer) error {
	return yaml.NewEncoder(w).Encode(a)
}

func (s yamlSerializer) MarshalBytes(a interface{}) ([]byte, error) {
	return yaml.Marshal(a)
}

func (yamlSerializer) Unmarshal(r io.Reader, a interface{}) error {
	return yaml.NewDecoder(r).Decode(a)
}

func (s yamlSerializer) UnmarshalBytes(data []byte, a interface{}) error {
	return yaml.Unmarshal(data, a)
}
