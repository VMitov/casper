package main

import (
	"flag"
	"net/url"
	"path/filepath"
	"strings"
	"syscall"

	"github.com/pkg/errors"
	cli "gopkg.in/urfave/cli.v2"
	"gopkg.in/urfave/cli.v2/altsrc"
)

// sourcesSliceFlag is the flag type that wraps cli.sourcesSliceFlag to allow
// for other values to be specified
type sourcesSliceFlag struct {
	*cli.StringSliceFlag
	set *flag.FlagSet
}

// newSourcesSliceFlag creates a new StringSliceFlag
func newSourcesSliceFlag(fl *cli.StringSliceFlag) *sourcesSliceFlag {
	return &sourcesSliceFlag{StringSliceFlag: fl, set: nil}
}

// Apply saves the flagSet for later usage calls, then calls the
// wrapped StringSliceFlag.Apply
func (f *sourcesSliceFlag) Apply(set *flag.FlagSet) {
	f.set = set
	f.StringSliceFlag.Apply(set)
}

// ApplyWithError saves the flagSet for later usage calls, then calls the
// wrapped StringSliceFlag.ApplyWithError
func (f *sourcesSliceFlag) ApplyWithError(set *flag.FlagSet) error {
	f.set = set
	return f.StringSliceFlag.ApplyWithError(set)
}

// ApplyInputSourceValue applies a StringSlice value to the flagSet if required
func (f *sourcesSliceFlag) ApplyInputSourceValue(context *cli.Context, isc altsrc.InputSourceContext) error {
	if f.set != nil {
		if !context.IsSet(f.Name) && !isEnvVarSet(f.EnvVars) {
			value, err := isc.StringSlice(f.StringSliceFlag.Name)
			if err != nil {
				return err
			}

			dir, err := getConfigDir(context)
			if err != nil {
				return err
			}

			value, err = fixPathsForSources(dir, value)
			if err != nil {
				return err
			}

			if value != nil {
				sliceValue := *(cli.NewStringSlice(value...))
				for _, name := range f.Names() {
					underlyingFlag := f.set.Lookup(name)
					if underlyingFlag != nil {
						underlyingFlag.Value = &sliceValue
					}
				}
			}
		}
	}
	return nil
}

func getConfigDir(context *cli.Context) (string, error) {
	configPath := context.String(configFlag)
	inputSourcePath, err := filepath.Abs(configPath)
	if err != nil {
		return "", errors.Wrapf(err, "resolving absolute path for config %v failed", configPath)
	}

	return filepath.Dir(inputSourcePath), nil
}

func fixPathsForSources(dir string, value []string) ([]string, error) {
	for i, v := range value {
		fixed, err := fixPathsForSource(dir, v)
		if err != nil {
			return value, err
		}
		if fixed != v {
			value[i] = fixed
		}
	}
	return value, nil
}

func fixPathsForSource(dir, value string) (string, error) {
	u, err := url.Parse(value)
	if err != nil {
		return "", errors.Wrapf(err, "parsing source %v failed", value)
	}

	if u.Scheme == "file" {
		fixed, err := fixPathsForFileSource(dir, u)
		if err != nil {
			return "", errors.Wrapf(err, "converting file source path %v to absolute failed", u)
		}

		return fixed, nil
	}

	return value, nil
}

func fixPathsForFileSource(dir string, u *url.URL) (string, error) {
	path := u.Host + u.Path
	if strings.HasPrefix(path, "/") {
		return u.String(), nil
	}

	var err error
	u.Path, err = url.PathUnescape(filepath.Join(dir, path))
	if err != nil {
		return "", err
	}

	u.Host = ""
	return u.String(), nil
}

func isEnvVarSet(envVars []string) bool {
	for _, envVar := range envVars {
		if _, ok := syscall.Getenv(envVar); ok {
			// TODO: Can't use this for bools as
			// set means that it was true or false based on
			// Bool flag type, should work for other types
			return true
		}
	}

	return false
}
