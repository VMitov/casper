package main

import (
	"net/url"
	"os"
	"strings"

	"github.com/miracl/casper/source"
	"github.com/pkg/errors"
)

const configScheme = "config"

type getSourcer func(u *url.URL) (*source.Source, error)

func getConfigSource(u *url.URL) (*source.Source, error) {
	body := map[string]interface{}{}
	for k, v := range u.Query() {
		if len(v) > 1 {
			body[k] = v
		}

		body[k] = v[0]
	}

	return source.NewSource(body), nil
}

func getFileSource(u *url.URL) (*source.Source, error) {

	path := u.Hostname() + u.EscapedPath()
	pathSlice := strings.Split(path, ".")
	format := pathSlice[len(pathSlice)-1]

	r, err := os.Open(path)
	if err != nil {
		return nil, errors.Wrapf(err, "opening file %v failed", path)
	}

	s, err := source.NewFileSource(r, format)
	if err != nil {
		return nil, errors.Wrap(err, "creating new file source failed")
	}

	return s, nil
}
