package source

import (
	"encoding/json"
	"io"
	"io/ioutil"

	yaml "gopkg.in/yaml.v2"
)

var fileSourceFormats = []string{"json", "yaml"}

type formatError string

func (e formatError) Error() string {
	return "Invalid file source format " + string(e)
}

// NewFileSource creates a source of type file
func NewFileSource(r io.Reader, format string) (*Source, error) {
	if r == nil {
		return &Source{}, nil
	}

	data, err := ioutil.ReadAll(r)
	if err != nil {
		return nil, err
	}

	body := map[string]interface{}{}
	switch format {
	case "json":
		err = json.Unmarshal(data, &body)
	case "yaml":
		err = yaml.Unmarshal(data, &body)
	default:
		return nil, formatError(format)
	}

	return NewSource(body), err
}
