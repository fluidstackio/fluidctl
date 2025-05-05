package format

import (
	"encoding/json"
	"fmt"

	"gopkg.in/yaml.v2"
)

type Format string

const (
	JSON Format = "json"
	YAML Format = "yaml"
)

type Marshal interface {
	Marshal(v any) ([]byte, error)
}

type JSONMarshaller struct{}

func (j *JSONMarshaller) Marshal(v any) ([]byte, error) {
	return json.MarshalIndent(v, "", "  ")
}

type YAMLMarshaller struct{}

func (y *YAMLMarshaller) Marshal(v any) ([]byte, error) {
	return yaml.Marshal(v)
}

type TableMarshaller struct{}

func (t *TableMarshaller) Marshal(v any) ([]byte, error) {
	return nil, fmt.Errorf("table marshalling not implemented")
}

func NewMarshaller(format Format) (Marshal, error) {
	switch format {
	case JSON:
		return &JSONMarshaller{}, nil
	case YAML:
		return &YAMLMarshaller{}, nil
	default:
		return nil, fmt.Errorf("unsupported format: %s", format)
	}
}
