package main

import (
	"fmt"
	"net/url"
	"os"

	"github.com/miracl/casper"
	"github.com/miracl/casper/source"
	consulstorage "github.com/miracl/casper/storage/consul"
	filestorage "github.com/miracl/casper/storage/file"
	"github.com/pkg/errors"
)

type context struct {
	path     string
	template *os.File
	storage  casper.Storage
	source   *source.Source
}

func newContext(path string, opts ...func(*context) error) (*context, error) {
	config := &context{
		path: path,
	}
	for _, opt := range opts {
		if err := opt(config); err != nil {
			return nil, err
		}
	}

	return config, nil
}

// Sets the config file if it is not in the current directory.
// All files should be relative to this path.
func (c *context) withPath(path string) {
	c.path = path
}

func withPath(path string) func(*context) error {
	return func(c *context) error {
		c.withPath(path)
		return nil
	}
}

func (c *context) withSources(sources []string) error {
	sourceTypes := map[string]getSourcer{
		configScheme: getConfigSource,
		"file":       getFileSource,
	}

	sourceList := make([]source.Getter, len(sources))
	for i, s := range sources {
		u, err := url.Parse(s)
		if err != nil {
			return errors.Wrapf(err, "parsing source %v failed", s)
		}

		if u.Scheme == "" {
			// default to config
			u = &url.URL{
				Scheme:   configScheme,
				RawQuery: s,
			}
		}

		getSourcer, ok := sourceTypes[u.Scheme]
		if !ok {
			return fmt.Errorf("invalid source format %v", u.Scheme)
		}

		sourceList[i], err = getSourcer(u)
		if err != nil {
			return err
		}
	}

	var err error
	c.source, err = source.NewMultiSourcer(sourceList...)
	return err
}

func withSources(sources []string) func(*context) error {
	return func(c *context) error {
		return c.withSources(sources)
	}
}

func (c *context) withTemplate(path string) error {
	var err error
	c.template, err = os.Open(path)
	return errors.Wrapf(err, "getting template %v failed", path)
}

func withTemplate(path string) func(*context) error {
	return func(c *context) error {
		return c.withTemplate(path)
	}
}

func (c *context) withFileStorage(path string) {
	c.storage = filestorage.New(path)
}

func withFileStorage(path string) func(*context) error {
	return func(c *context) error {
		c.withFileStorage(path)
		return nil
	}
}

func (c *context) withConsulStorage(addr string) error {
	var err error
	c.storage, err = consulstorage.New(addr)
	return errors.Wrap(err, "creating Consul storage failed")
}

func withConsulStorage(addr string) func(*context) error {
	return func(c *context) error {
		return c.withConsulStorage(addr)
	}
}
