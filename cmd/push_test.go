package cmd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"regexp"
	"testing"

	"golang.org/x/crypto/ssh/terminal"

	"github.com/miracl/casper/lib/caspertest"
)

func TestPushRun(t *testing.T) {
	testCases := []struct {
		storage     string
		tmpl        string
		interactive bool
		backup      bool
		sources     *[]map[string]interface{}
		force       bool
		output      string
	}{
		{
			"",
			`key: {{.placeholder}}`,
			false,
			false,
			&[]map[string]interface{}{
				{
					"type": "config",
					"vals": map[string]string{
						"placeholder": "val",
					},
				},
			},
			true,
			"\nThe following changes will be applied:\n-\n+key: val\nApplying changes...\nDone.\n",
		},
		{
			"",
			`key: {{.placeholder}}`,
			false,
			false,
			&[]map[string]interface{}{
				{
					"type": "config",
					"vals": map[string]string{
						"placeholder": "val",
					},
				},
			},
			false,
			"\nThe following changes will be applied:\n-\n+key: val\nContinue [y/N]: Canceled\n",
		},
		{
			"key: val",
			`key: {{.placeholder}}`,
			false,
			false,
			&[]map[string]interface{}{
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
		{
			"",
			`key: {{.placeholder}}`,
			true,
			true,
			&[]map[string]interface{}{
				{
					"type": "config",
					"vals": map[string]string{
						"placeholder": "val",
					},
				},
			},
			false,
			"\nThe following changes will be applied:\n-\n+key: val\nContinue [y/N]: Canceled\n",
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

			pushConf := &pushConfig{
				tmpf.Name(), false, "jsonraw", "file", "",
				tc.interactive, tc.backup, tc.sources, tc.force, false,
			}

			out := caspertest.GetStdout(t, func() {
				err = pushRun(
					pushConf,
					map[string]interface{}{"path": configf.Name()})
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
			"\nThe following changes will be applied:\n-\n+key: val\nApplying changes...\nDone.\n",
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

func TestGenerateBackupFilename(t *testing.T) {
	t.Run("Case0", func(t *testing.T) {
		expression := regexp.MustCompile("\\d{10}_backup.txt")
		generated := generateBackupFilename()

		if !expression.MatchString(generated) {
			t.Errorf("Generated filename %s did not match the expected format", generated)
		}
	})
}

func TestSaveBackup(t *testing.T) {
	testCases := []string{
		"test1234",
	}

	for i := range testCases {
		t.Run(fmt.Sprintf("Case%v", i), func(t *testing.T) {
			filename, err := saveBackup(&testCases[i])
			defer os.Remove(filename)

			result, err := ioutil.ReadFile(filename)
			if err != nil {
				t.Errorf("Could not read file %s", filename)
			}

			if bytes.Compare(result, []byte(testCases[i])) != 0 {
				t.Errorf("Wrong content: Expected \"%s\", got \"%s\"", testCases[i], string(result))
			}
		})
	}
}

func TestBackup(t *testing.T) {
	testCases := []struct {
		storage       string
		createFile    bool
		errorExpected bool
	}{
		{"test: 1234abc", true, false},
		{"", false, true},
	}

	backupFormat := "Backup has been saved as (?P<Filename>\\d{10}_backup.txt).*"
	expression := regexp.MustCompile(backupFormat)

	for i, tc := range testCases {
		storageFilename := ""

		t.Run(fmt.Sprintf("Case%v", i), func(t *testing.T) {
			// Prepare config

			if tc.createFile {
				testFilename := fmt.Sprintf("Case%vConfig.yaml", i)
				storageFile, err := caspertest.PrepareTmpFile(testFilename, []byte(tc.storage))
				if err != nil {
					t.Fatal(err)
				}

				storageFilename = storageFile.Name()
				defer os.Remove(storageFilename)
			}

			sourcesList, _ := getSliceStringMapIface("")
			pc := &pushConfig{
				"", false, tc.storage, "file", "",
				false, true, &sourcesList, false, true,
			}
			conf := map[string]interface{}{"path": storageFilename}

			out := caspertest.GetStdout(t, func() {
				err := backup(pc, conf)

				if !tc.errorExpected && err != nil {
					t.Errorf("Backup failed: %v", err)
				} else if tc.errorExpected && err == nil {
					t.Errorf("An error was expected but none was detected")
				}
			})

			if tc.errorExpected {
				return
			}

			if !expression.MatchString(out) {
				t.Errorf("Output \"%s\" did not match "+
					"the expected format: \"%s\"",
					out, backupFormat)
			} else {
				backupFile := expression.FindStringSubmatch(out)[1]
				defer os.Remove(backupFile)
			}
		})
	}
}

func TestUsePlain(t *testing.T) {
	testCases := []struct {
		plain bool
	}{
		{true},
		{false},
	}

	for i, tc := range testCases {
		t.Run(fmt.Sprintf("Case%v", i), func(t *testing.T) {
			result := usePlain(tc.plain)
			isTerminal := terminal.IsTerminal(int(os.Stdout.Fd()))
			expected := tc.plain || !isTerminal

			if result != expected {
				t.Errorf("Got %t for input %t", result, expected)
			}
		})
	}
}
