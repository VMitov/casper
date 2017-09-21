package casper

import (
	"bytes"
	"io"
	"io/ioutil"
	"text/template"

	"github.com/miracl/casper/source"
)

var funcMap = template.FuncMap{
	"last":    isLast,
	"notLast": isNotLast,
	"quote":   quote,
}

// BuildConfig represent a configuration
type BuildConfig struct {
	Template io.Reader
	Source   source.ValuesSourcer
}

// Build creates the config based on the template and the environment files
func (c BuildConfig) Build() ([]byte, error) {
	// Compile the template for the config
	cfgTmplBody, err := ioutil.ReadAll(c.Template)
	if err != nil {
		return nil, err
	}

	cfgTmpl, err := template.New("config").
		Funcs(funcMap).
		Parse(string(cfgTmplBody))
	if err != nil {
		return nil, err
	}

	var cfg bytes.Buffer
	if err := cfgTmpl.Execute(&cfg, c.Source.Get()); err != nil {
		return nil, err
	}

	// Convert to string
	cfgStr, err := ioutil.ReadAll(&cfg)
	if err != nil {
		return nil, err
	}

	return cfgStr, nil
}
