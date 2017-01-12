package cmd

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"testing"

	"github.com/miracl/casper/lib/caspertest"
)

func TestBuildConfig(t *testing.T) {
	testCases := []struct {
		tmpl string
		srcs []map[string]interface{}
		out  string
		ok   bool
	}{
		{
			"key1: {{.key1}}, key2: {{.key2}}",
			[]map[string]interface{}{
				map[string]interface{}{
					"type": "config",
					"vals": map[interface{}]interface{}{
						"key1": "var1",
						"key2": "var2",
					},
				},
			},
			"key1: var1, key2: var2",
			true,
		},
		{
			"key1: {{.key1}}, key2: {{.key2}}",
			[]map[string]interface{}{
				map[string]interface{}{
					"type": "bad",
				},
			},
			"",
			false,
		},
	}

	for i, tc := range testCases {
		t.Run(fmt.Sprintf("Case%v", i), func(t *testing.T) {
			// Prepare config file
			tmlpFile, err := ioutil.TempFile("", "TestBuild")
			if err != nil {
				log.Fatal(err)
			}
			defer os.Remove(tmlpFile.Name()) // clean up

			if _, err := tmlpFile.Write([]byte(tc.tmpl)); err != nil {
				t.Fatal(err)
			}
			if err := tmlpFile.Close(); err != nil {
				t.Fatal(err)
			}

			// Build
			out, err := buildConfig(tmlpFile.Name(), tc.srcs)
			if tc.ok != (err == nil) {
				if err != nil {
					t.Fatal("Failed with", err)
				} else {
					t.Fatal("Didn't fail")
				}
			}

			if tc.ok && string(out) != tc.out {
				t.Errorf("Got '%v'; want '%v'", out, tc.out)
			}
		})
	}
}

func TestBuildRun(t *testing.T) {
	testCases := []struct {
		tmpl    string
		sources []map[string]interface{}
		output  string
	}{
		{
			`key: {{.placeholder}}`,
			[]map[string]interface{}{
				map[string]interface{}{
					"type": "config",
					"vals": map[string]string{
						"placeholder": "val",
					},
				},
			},
			"key: val",
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

			out := caspertest.GetStdout(t, func() {
				err = buildRun(tmpf.Name(), tc.sources)
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

func TestBuild(t *testing.T) {
	testCases := []struct {
		tmpl    string
		sources []map[string]interface{}
		output  string
	}{
		{
			`key: {{.placeholder}}`,
			[]map[string]interface{}{
				map[string]interface{}{
					"type": "config",
					"vals": map[string]string{
						"placeholder": "val",
					},
				},
			},
			"key: val",
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

			configJson, err := json.Marshal(map[string]interface{}{
				"template": tmplf.Name(),
				"sources":  tc.sources,
			})
			if err != nil {
				t.Fatal(err)
			}

			// Prepare config
			cfgf, err := caspertest.PrepareTmpFile(fmt.Sprintf("Case%vConfig.yaml", i), configJson)
			if err != nil {
				t.Fatal(err)
			}
			defer os.Remove(cfgf.Name())

			os.Args = []string{"casper", "build", "-c", cfgf.Name()}
			out := caspertest.GetStdout(t, func() {
				buildCmd.Execute()
			})

			if out != tc.output {
				t.Errorf("Got %#v; Want %#v", out, tc.output)
			}
		})
	}
}
