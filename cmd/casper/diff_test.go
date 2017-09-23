package main

import (
	"encoding/json"
	"fmt"
	"os"
	"testing"

	"github.com/miracl/casper/caspertest"
)

func TestDiffRun(t *testing.T) {
	testCases := []struct {
		storage string
		tmpl    string
		key     string
		sources []map[string]interface{}
		output  string
	}{
		{
			`key: oldval`,
			`key: {{.placeholder}}`,
			"",
			[]map[string]interface{}{
				{
					"type": "config",
					"vals": map[string]string{
						"placeholder": "val",
					},
				},
			},
			"" +
				"-key: oldval\n" +
				"+key: val\n",
		},
		{
			`key: val`,
			`key: {{.placeholder}}`,
			"",
			[]map[string]interface{}{
				{
					"type": "config",
					"vals": map[string]string{
						"placeholder": "val",
					},
				},
			},
			"No changes\n",
		},
		{
			`key: oldval`,
			`key: {{.placeholder}}`,
			"key",
			[]map[string]interface{}{
				{
					"type": "config",
					"vals": map[string]string{
						"placeholder": "val",
					},
				},
			},
			"" +
				"-key: oldval\n" +
				"+key: val\n",
		},
		{
			`key: val`,
			`key: {{.placeholder}}`,
			"key",
			[]map[string]interface{}{
				{
					"type": "config",
					"vals": map[string]string{
						"placeholder": "val",
					},
				},
			},
			"No changes for key key\n",
		},
	}

	for i, tc := range testCases {
		t.Run(fmt.Sprintf("Case%v", i), func(t *testing.T) {
			// Prepare template
			tmpf, err := caspertest.PrepareTmpFile(fmt.Sprintf("Case%vTmpl", i), []byte(tc.tmpl))
			if err != nil {
				t.Fatal(err)
			}
			defer os.Remove(tmpf.Name())

			// Prepare config
			configf, err := caspertest.PrepareTmpFile(fmt.Sprintf("Case%vConfig", i), []byte(tc.storage))
			if err != nil {
				t.Fatal(err)
			}
			defer os.Remove(configf.Name())

			out := caspertest.GetStdout(t, func() {
				err = diffRun(tmpf.Name(), "yaml", tc.key, tc.sources, "file", map[string]interface{}{"path": configf.Name()}, false)
				if err != nil {
					t.Fatal(err)
				}
			})

			if out != tc.output {
				t.Errorf("Got %#v; want %#v", out, tc.output)
			}
		})
	}
}

func TestDiff(t *testing.T) {
	testCases := []struct {
		storage string
		tmpl    string
		sources []map[string]interface{}
		output  string
	}{
		{
			`key: oldval`,
			`key: {{.placeholder}}`,
			[]map[string]interface{}{
				{
					"type": "config",
					"vals": map[string]string{
						"placeholder": "val",
					},
				},
			},
			"" +
				"-key: oldval\n" +
				"+key: val\n",
		},
		{
			`key: val`,
			`key: {{.placeholder}}`,
			[]map[string]interface{}{
				{
					"type": "config",
					"vals": map[string]string{
						"placeholder": "val",
					},
				},
			},
			"No changes\n",
		},
	}

	for i, tc := range testCases {
		t.Run(fmt.Sprintf("Case%v", i), func(t *testing.T) {
			// Prepare template
			tmplf, err := caspertest.PrepareTmpFile(fmt.Sprintf("Case%vTmpl", i), []byte(tc.tmpl))
			if err != nil {
				t.Fatal(err)
			}
			defer os.Remove(tmplf.Name())

			// Prepare storage
			strf, err := caspertest.PrepareTmpFile(fmt.Sprintf("Case%vStorage", i), []byte(tc.storage))
			if err != nil {
				t.Fatal(err)
			}
			defer os.Remove(strf.Name())

			configJSON, err := json.Marshal(map[string]interface{}{
				"template": tmplf.Name(),
				"format":   "yaml",
				"sources":  tc.sources,
				"storage": map[string]interface{}{
					"type": "file",
					"config": map[string]string{
						"path": strf.Name(),
					},
				},
			})
			if err != nil {
				t.Fatal(err)
			}

			// Prepare config
			cfgf, err := caspertest.PrepareTmpFile(fmt.Sprintf("Case%vConfig.yaml", i), configJSON)
			if err != nil {
				t.Fatal(err)
			}
			defer os.Remove(cfgf.Name())

			os.Args = []string{"casper", "diff", "-c", cfgf.Name(), "-p"}
			out := caspertest.GetStdout(t, func() {
				diffCmd.Execute()
			})

			if out != tc.output {
				t.Errorf("Got %#v; Want %#v", out, tc.output)
			}
		})
	}
}
