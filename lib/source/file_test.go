package source

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"reflect"
	"testing"
)

var errSuccFail = errors.New("Successful fail")

type FailingReader struct{}

func (FailingReader) Read(p []byte) (n int, err error) {
	return 0, errSuccFail
}

func TestFileSourcer(t *testing.T) {
	testCases := []struct {
		file   io.Reader
		format string
		parsed map[string]interface{}
		ok     bool
	}{
		{
			bytes.NewBufferString(`{
				"key1": "var1",
				"key2": "var2"
			}`),
			"json",
			map[string]interface{}{
				"key1": "var1",
				"key2": "var2",
			},
			true,
		},
		{
			bytes.NewBufferString("key1: var1\nkey2: var2\n"),
			"yaml",
			map[string]interface{}{
				"key1": "var1",
				"key2": "var2",
			},
			true,
		},
		{
			bytes.NewBufferString(" key1: var1\nkey2: var2\n"),
			"yaml",
			nil,
			false,
		},
		{
			bytes.NewBufferString("key1: var1\nkey2: var2\n"),
			"invalid format",
			nil,
			false,
		},
		{
			bytes.NewBufferString(`key1=var1, key2=var2`),
			"json",
			nil,
			false,
		},
		{
			bytes.NewBufferString(`key1=var1, key2=var2`),
			"json",
			nil,
			false,
		},
		{nil, "json", nil, true},
		{FailingReader{}, "json", nil, false},
	}

	for i, tc := range testCases {
		t.Run(fmt.Sprintf("Case%v", i), func(t *testing.T) {
			s, err := NewFileSource(tc.file, tc.format)
			if tc.ok != (err == nil) {
				if err != nil {
					t.Fatal(err)
				} else {
					t.Fatal("Get should have failed but haven't")
				}
			}

			// Compare
			if tc.ok && !reflect.DeepEqual(s.Get(), tc.parsed) {
				t.Errorf("Got %v; want %v", s.Get(), tc.parsed)
			}
		})
	}
}
