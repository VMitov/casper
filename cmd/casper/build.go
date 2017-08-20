package main

import (
	"fmt"
	"os"

	casper "github.com/miracl/casper/lib"
)

func buildRun(tmplF string, srcs []string) error {
	out, err := buildConfig(tmplF, srcs)
	if err != nil {
		return err
	}

	fmt.Print(string(out))
	return nil
}

func buildConfig(tmplF string, srcs []string) ([]byte, error) {
	tmpl, err := os.Open(tmplF)
	if err != nil {
		return nil, err
	}

	source, err := getMultiSourcer(srcs)
	if err != nil {
		return nil, err
	}

	cfg, err := casper.BuildConfig{
		Tmlp:   tmpl,
		Source: source,
	}.Build()
	if err != nil {
		return nil, err
	}

	return cfg, nil
}
