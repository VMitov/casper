package casper

import (
	"bytes"
	"fmt"
	"testing"

	"github.com/miracl/casper/lib/source"
)

func TestBuild(t *testing.T) {
	testCases := []struct {
		tmpl     string
		source   source.ValuesSourcer
		validate bool
		res      string
	}{
		{
			`{"cfg1": "{{.key1}}", "cfg2": "{{.key2}}"}`,
			source.NewSource(map[string]interface{}{
				"key1": "var1",
				"key2": "var2",
			}),
			true,
			`{"cfg1": "var1", "cfg2": "var2"}`,
		},
		{
			`{"cfg1": "{{.key1}}", "cfg2": "{{if .key2}}{{.key2}}{{end}}", "cfg3": "{{.key3}}"}`,
			source.NewSource(map[string]interface{}{
				"key1": "var1",
			}),
			true,
			`{"cfg1": "var1", "cfg2": "", "cfg3": "<no value>"}`,
		},
		{
			``,
			source.NewSource(map[string]interface{}{
				"key1": "var1",
			}),
			true,
			``,
		},
		{
			`{"cfg1": "{{something}}", "cfg2": "{{somethingElse}}"}`,
			source.NewSource(map[string]interface{}{
				"key1": "var1",
				"key2": "var2",
			}),
			false,
			`{"cfg1": "{{something}}", "cfg2": "{{somethingElse}}"}`,
		},
	}

	for i, tc := range testCases {
		t.Run(fmt.Sprintf("Case%v", i), func(t *testing.T) {

			// Build
			conf := BuildConfig{
				Tmlp:   bytes.NewBufferString(tc.tmpl),
				Source: tc.source,
			}

			var config []byte
			var err error

			if tc.validate {
				config, err = conf.Build()
			} else {
				config, err = conf.BuildNoValidation()
			}

			if err != nil {
				t.Fatal(err)
			}

			// Compare
			if string(config) != tc.res {
				t.Errorf("Got %v; want %v", string(config), tc.res)
			}
		})
	}
}
