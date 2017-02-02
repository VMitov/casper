package cmd

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"testing"

	"github.com/miracl/casper/lib/caspertest"
)

func TestPushRun(t *testing.T) {
	testCases := []struct {
		storage string
		tmpl    string
		sources []map[string]interface{}
		force   bool
		output  string
	}{
		{
			"",
			`key: {{.placeholder}}`,
			[]map[string]interface{}{
				{
					"type": "config",
					"vals": map[string]string{
						"placeholder": "val",
					},
				},
			},
			true,
			"" +
				"-\n" +
				"+key: val\n" +
				"Applying changes...\n",
		},
		{
			"",
			`key: {{.placeholder}}`,
			[]map[string]interface{}{
				{
					"type": "config",
					"vals": map[string]string{
						"placeholder": "val",
					},
				},
			},
			false,
			"" +
				"-\n" +
				"+key: val\n" +
				"Continue[y/N]: Canceled\n",
		},
		{
			"key: val",
			`key: {{.placeholder}}`,
			[]map[string]interface{}{
				{
					"type": "config",
					"vals": map[string]string{
						"placeholder": "val",
					},
				},
			},
			true,
			"No changes\n",
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
				err = pushRun(tmpf.Name(), "jsonraw", "", tc.sources, "file", map[string]interface{}{"path": configf.Name()}, tc.force, false)
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

func TestPush(t *testing.T) {
	testCases := []struct {
		storage string
		tmpl    string
		sources []map[string]interface{}
		output  string
		result  string
	}{
		{
			"",
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
				"-\n" +
				"+key: val\n" +
				"Applying changes...\n",
			`key: val`,
		},
		{
			"key: val",
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
			`key: val`,
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

			os.Args = []string{"casper", "push", "-c", cfgf.Name(), "-p", "--force"}
			out := caspertest.GetStdout(t, func() {
				pushCmd.Execute()
			})

			if out != tc.output {
				t.Errorf("Got %#v; Want %#v", out, tc.output)
			}

			// Check the storage is correct
			result, err := ioutil.ReadFile(strf.Name())
			if err != nil {
				t.Fatal(err)
			}
			if string(result) != tc.result {
				t.Errorf("Got %#v; Want %#v", string(result), tc.result)
			}
		})
	}
}
