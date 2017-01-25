package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"testing"

	"github.com/miracl/casper/lib/caspertest"
)

func TestFetchRun(t *testing.T) {
	testCases := []struct {
		storage string
	}{
		{"key: val"},
	}

	for i, tc := range testCases {
		t.Run(fmt.Sprintf("Case%v", i), func(t *testing.T) {
			// Prepare storage
			strf, err := caspertest.PrepareTmpFile(fmt.Sprintf("Case%vStorage", i), []byte(tc.storage))
			if err != nil {
				t.Fatal(err)
			}
			defer os.Remove(strf.Name())

			config := map[string]interface{}{
				"path": strf.Name(),
			}

			out := caspertest.GetStdout(t, func() {
				err = fetchRun("file", config, "jsonraw")
				if err != nil {
					t.Fatal(err)
				}
			})

			exp := tc.storage + "\n"
			if out != exp {
				t.Errorf("Got %#v; want %#v", out, exp)
			}
		})
	}
}

func TestFetch(t *testing.T) {
	testCases := []struct {
		storage string
	}{
		{"key: value"},
	}

	for i, tc := range testCases {
		t.Run(fmt.Sprintf("Case%v", i), func(t *testing.T) {
			// Prepare storage
			strf, err := caspertest.PrepareTmpFile(fmt.Sprintf("Case%vStorage", i), []byte(tc.storage))
			if err != nil {
				t.Fatal(err)
			}
			defer os.Remove(strf.Name())

			configJSON, err := json.Marshal(map[string]interface{}{
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

			os.Args = []string{"casper", "fetch", "-c", cfgf.Name()}
			out := caspertest.GetStdout(t, func() {
				fetchCmd.Execute()
			})

			exp := tc.storage + "\n"
			if out != exp {
				t.Errorf("Got %#v; want %#v", out, exp)
			}
		})
	}
}
