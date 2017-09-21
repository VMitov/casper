package main

import (
	"flag"
	"net/url"
	"path/filepath"
	"strings"
	"syscall"

	cli "gopkg.in/urfave/cli.v2"
	"gopkg.in/urfave/cli.v2/altsrc"
)

func getConfigDir(context *cli.Context) (string, error) {
	inputSourcePath, err := filepath.Abs(context.String(configFlag))
	if err != nil {
		return "", err
	}

	return filepath.Dir(inputSourcePath), nil
}

// pathFlag is the flag type that wraps cli.pathFlag to allow
// for other values to be specified
type pathFlag struct {
	*cli.StringFlag
	set *flag.FlagSet
}

// newPathFlag creates a new StringFlag
func newPathFlag(fl *cli.StringFlag) *pathFlag {
	return &pathFlag{StringFlag: fl, set: nil}
}

// Apply saves the flagSet for later usage calls, then calls the
// wrapped StringFlag.Apply
func (f *pathFlag) Apply(set *flag.FlagSet) {
	f.set = set
	f.StringFlag.Apply(set)
}

// ApplyWithError saves the flagSet for later usage calls, then calls the
// wrapped StringFlag.ApplyWithError
func (f *pathFlag) ApplyWithError(set *flag.FlagSet) error {
	f.set = set
	return f.StringFlag.ApplyWithError(set)
}

// ApplyInputSourceValue applies a String value to the flagSet if required
func (f *pathFlag) ApplyInputSourceValue(context *cli.Context, isc altsrc.InputSourceContext) error {
	if f.set != nil {
		if !(context.IsSet(f.Name) || isEnvVarSet(f.EnvVars)) {
			value, err := isc.String(f.StringFlag.Name)
			if err != nil {
				return err
			}

			if value != "" {
				dir, err := getConfigDir(context)
				if err != nil {
					return err
				}

				for _, name := range f.Names() {
					f.set.Set(name, absToFile(dir, value))
				}
			}
		}
	}
	return nil
}

func absToFile(dir, path string) string {
	if dir == "" || filepath.IsAbs(path) {
		return path
	}

	return filepath.Join(dir, path)
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
		return "", err
	}

	if u.Scheme == "file" {
		fixed, err := fixPathsForFileSource(dir, u)
		if err != nil {
			return "", err
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
