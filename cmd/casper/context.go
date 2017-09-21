package main

import (
	"net/url"
	"os"

	"github.com/miracl/casper/source"
)

type context struct {
	path     string
	template *os.File
	storage  storage
	source   *source.Source
}

type storage interface {
	String(format string) (string, error)
	FormatIsValid(format string) bool
	DefaultFormat() string
	GetChanges(config []byte, format, key string) (changes, error)
	Diff(cs changes, pretty bool) string
	Push(cs changes) error
}

type changes interface {
	Len() int
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

// withPath sets the config file if it is not in the current directory.
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

	sourceList := make([]source.ValuesSourcer, len(sources))
	for i, s := range sources {
		u, err := url.Parse(s)
		if err != nil {
			return err
		}

		if u.Scheme == "" {
			// Default to config
			u = &url.URL{
				Scheme:   configScheme,
				RawQuery: s,
			}
		}

		getSourcer, ok := sourceTypes[u.Scheme]
		if !ok {
			return errSourceFormat
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
	return err
}

func withTemplate(path string) func(*context) error {
	return func(c *context) error {
		return c.withTemplate(path)
	}
}

func (c *context) withFileStorage(path string) {
	c.storage = &fileStorage{path}
}

func withFileStorage(path string) func(*context) error {
	return func(c *context) error {
		c.withFileStorage(path)
		return nil
	}
}

func (c *context) withConsulStorage(addr string) error {
	var err error
	c.storage, err = newConsulStorage(addr)
	return err
}

func withConsulStorage(addr string) func(*context) error {
	return func(c *context) error {
		return c.withConsulStorage(addr)
	}
}
