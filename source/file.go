package source

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"

	"github.com/pkg/errors"
	yaml "gopkg.in/yaml.v2"
)

// NewFileSource creates new file source.
func NewFileSource(r io.Reader, format string) (*Source, error) {
	if r == nil {
		return &Source{}, nil
	}

	data, err := ioutil.ReadAll(r)
	if err != nil {
		return nil, errors.Wrap(err, "reading file source failed")
	}

	body := map[string]interface{}{}
	switch format {
	case "json":
		if err := json.Unmarshal(data, &body); err != nil {
			return nil, errors.Wrap(err, "parsing json failed")
		}
	case "yaml":
		if err := yaml.Unmarshal(data, &body); err != nil {
			return nil, errors.Wrap(err, "parsing yaml failed")
		}
	default:
		return nil, fmt.Errorf("unsupported file source format '%v'", format)
	}

	return NewSource(body), nil
}
