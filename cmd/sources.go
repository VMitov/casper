package cmd

import (
	"errors"
	"fmt"
	"os"
	"strconv"

	"github.com/miracl/casper/lib/source"
)

var errSourceFormat = errors.New("Sources invalid format")

type sourceFormatError struct {
	msg string
	err error
}

func (e sourceFormatError) Error() string {
	s := fmt.Sprintf("Invalid source definition: %v", e.msg)
	if e.err != nil {
		s = fmt.Sprintf("%v (Err:%v)", s, e.err)
	}
	return s
}

func getMultiSourcer(srcs []map[string]interface{}) (*source.Source, error) {
	sourceTypes := map[string]getSourcer{
		"config": getConfigSource,
		"file":   getFileSource,
	}

	sourceList := make([]source.ValuesSourcer, len(srcs))
	for i, sourceCfg := range srcs {
		fI, ok := sourceCfg["type"]
		if !ok {
			return nil, sourceFormatError{"no type", nil}
		}

		file, ok := fI.(string)
		if !ok {
			return nil, sourceFormatError{"not string", nil}
		}

		f, ok := sourceTypes[file]
		if !ok {
			return nil, sourceFormatError{"invalid type", nil}
		}

		s, err := f(sourceCfg)
		if err != nil {
			return nil, err
		}

		sourceList[i] = s
	}

	return source.NewMultiSourcer(sourceList...)
}

type getSourcer func(map[string]interface{}) (*source.Source, error)

func getConfigSource(cfg map[string]interface{}) (*source.Source, error) {
	valsI, ok := cfg["vals"]
	if !ok {
		return nil, sourceFormatError{"no 'vals' for config source", nil}
	}

	body, err := toStringMapString(valsI)
	if err != nil {
		return nil, err
	}

	return source.NewSource(body), nil
}

func getFileSource(cfg map[string]interface{}) (*source.Source, error) {
	pathI, ok := cfg["file"]
	if !ok {
		return nil, sourceFormatError{"no 'file' for file source", nil}
	}

	path, ok := pathI.(string)
	if !ok {
		return nil, sourceFormatError{"invalid 'file' for file source", nil}
	}

	format := "json"
	formatI, ok := cfg["format"]
	if ok {
		format = formatI.(string)
	}

	r, err := os.Open(path)
	if err != nil {
		return nil, err
	}

	s, err := source.NewFileSource(r, format)
	if err != nil {
		return nil, sourceFormatError{"unable to create file source", err}
	}

	return s, nil
}

type convertError struct {
	in interface{}
	t  string
}

func (e convertError) Error() string {
	return fmt.Sprintf("Unable to convert %#v to %v", e.in, e.t)
}

func toStringMapString(in interface{}) (map[string]interface{}, error) {
	body := map[string]interface{}{}
	switch val := in.(type) {
	case map[string]interface{}:
		body = val
	case map[string]string:
		for k, v := range val {
			body[k] = v
		}
	case map[interface{}]interface{}:
		for k, v := range val {
			s, ok := toString(k)
			if !ok {
				return nil, convertError{k, "string"}
			}
			body[s] = v
		}
	default:
		return nil, convertError{in, "map[string]interface{}"}
	}

	return body, nil
}

func toString(s interface{}) (string, bool) {
	switch a := s.(type) {
	case string:
		return a, true
	case int:
		return strconv.Itoa(a), true
	case bool:
		return strconv.FormatBool(a), true
	default:
		return "", false
	}
}
