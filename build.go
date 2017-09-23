package casper

import (
	"bytes"
	"io"
	"io/ioutil"
	"text/template"

	"github.com/miracl/casper/source"
	"github.com/pkg/errors"
)

var funcMap = template.FuncMap{
	"last":    isLast,
	"notLast": isNotLast,
	"quote":   quote,
}

// BuildConfig represent a configuration
type BuildConfig struct {
	Template io.Reader
	Source   source.Getter
}

// Build creates the config based on the template and the environment files
func (c BuildConfig) Build() ([]byte, error) {
	// Compile the template for the config
	cfgTmplBody, err := ioutil.ReadAll(c.Template)
	if err != nil {
		return nil, errors.Wrap(err, "reading template failed")
	}

	cfgTmpl, err := template.New("config").
		Funcs(funcMap).
		Parse(string(cfgTmplBody))
	if err != nil {
		return nil, errors.Wrap(err, "template error")
	}

	var cfg bytes.Buffer
	if err := cfgTmpl.Execute(&cfg, c.Source.Get()); err != nil {
		return nil, errors.Wrap(err, "executing template failed")
	}

	// Convert to string
	cfgStr, err := ioutil.ReadAll(&cfg)
	if err != nil {
		return nil, errors.Wrap(err, "serializing template to string failed")
	}

	return cfgStr, nil
}
