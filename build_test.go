package casper

import (
	"bytes"
	"fmt"
	"testing"

	"github.com/miracl/casper/source"
)

func TestBuild(t *testing.T) {
	testCases := []struct {
		tmpl   string
		source source.Getter
		res    string
	}{
		{
			`{"cfg1": "{{.key1}}", "cfg2": "{{.key2}}"}`,
			source.NewSource(map[string]interface{}{
				"key1": "var1",
				"key2": "var2",
			}),
			`{"cfg1": "var1", "cfg2": "var2"}`,
		},
		{
			`{"cfg1": "{{.key1}}", "cfg2": "{{if .key2}}{{.key2}}{{end}}", "cfg3": "{{.key3}}"}`,
			source.NewSource(map[string]interface{}{
				"key1": "var1",
			}),
			`{"cfg1": "var1", "cfg2": "", "cfg3": "<no value>"}`,
		},
		{
			``,
			source.NewSource(map[string]interface{}{
				"key1": "var1",
			}),
			``,
		},
	}

	for i, tc := range testCases {
		t.Run(fmt.Sprintf("Case%v", i), func(t *testing.T) {

			// Build
			config, err := BuildConfig{
				Template: bytes.NewBufferString(tc.tmpl),
				Source:   tc.source,
			}.Build()
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
