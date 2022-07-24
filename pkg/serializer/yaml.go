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

func (yamlSerializer) Unmarshal(in io.Reader, a interface{}) error {
	return yaml.NewDecoder(in).Decode(a)
}
