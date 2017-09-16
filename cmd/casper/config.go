package main

import (
	"fmt"
	"net/url"
	"os"
	"path/filepath"

	"github.com/miracl/casper/source"
)

type config struct {
	path     string
	template *os.File
	storage  storage
	source   *source.Source
}

func newConfig(path string, opts ...func(*config) error) (*config, error) {
	config := &config{
		path: path,
	}
	for _, opt := range opts {
		if err := opt(config); err != nil {
			return nil, err
		}
	}

	return config, nil
}

func (c *config) open(path string) (*os.File, error) {
	if filepath.IsAbs(path) {
		return os.Open(path)
	}

	if c.path == defaultPath {
		return os.Open(path)
	}

	return c.find(path)
}

// find checks if path is relative to current dir or to config file
func (c *config) find(path string) (*os.File, error) {
	absCfgPath, err := filepath.Abs(c.path)
	if err != nil {
		return nil, err
	}
	pathRelToConfig := filepath.Clean(filepath.Join(filepath.Dir(absCfgPath), path))

	pathRelToCurDir, err := filepath.Abs(path)
	if err != nil {
		return nil, err
	}

	for _, p := range []string{pathRelToCurDir, pathRelToConfig} {
		f, err := os.Open(p)
		if err == nil {
			return f, nil
		}
		if !os.IsNotExist(err) {
			return nil, err
		}
	}

	return nil, fmt.Errorf("unable to find file %v; tried: %v, %v", path, pathRelToCurDir, pathRelToConfig)
}

// withPath sets the config file if it is not in the current directory.
// All files should be relative to this path.
func (c *config) withPath(path string) {
	c.path = path
}

func withPath(path string) func(*config) error {
	return func(c *config) error {
		c.withPath(path)
		return nil
	}
}

func (c *config) withSources(sources []string) error {
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

		// TODO!
		if c.path != defaultPath && !filepath.IsAbs(u.Hostname()) {
			absCfgPath, err := filepath.Abs(c.path)
			if err != nil {
				return err
			}
			u.Host = filepath.Clean(filepath.Join(filepath.Dir(absCfgPath), u.Hostname()))
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

func withSources(sources []string) func(*config) error {
	return func(c *config) error {
		return c.withSources(sources)
	}
}

func (c *config) withTemplate(path string) error {
	var err error
	c.template, err = c.open(path)
	return err
}

func withTemplate(path string) func(*config) error {
	return func(c *config) error {
		return c.withTemplate(path)
	}
}

func (c *config) withFileStorage(path string) {
	c.storage = &fileStorage{path}
}

func withFileStorage(path string) func(*config) error {
	return func(c *config) error {
		c.withFileStorage(path)
		return nil
	}
}

func (c *config) withConsulStorage(addr string) error {
	var err error
	c.storage, err = newConsulStorage(addr)
	return err
}

func withConsulStorage(addr string) func(*config) error {
	return func(c *config) error {
		return c.withConsulStorage(addr)
	}
}
