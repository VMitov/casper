package main

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestExample(t *testing.T) {
	wd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}

	outputFileName := filepath.Join(wd, "../../example/output.yaml")
	outputFileData, err := ioutil.ReadFile(outputFileName)
	if err != nil {
		t.Fatal(err)
	}
	outputFile := string(outputFileData)

	noChanges := "No changes\n"
	changes := "-key1: val1\n" +
		"key2: val2\n" +
		"\n" +
		"+key1: val1a\n" +
		"key2: val2a\n" +
		"\n"
	prompt := "Continue[y/N]: "
	applyingChanges := "Applying changes...\n"
	canceled := "Canceled"

	configAbsPath := filepath.Join(wd, "../../example/config.yaml")
	templateAbsPath := filepath.Join(wd, "../../example/template.yaml")
	fmt.Println(configAbsPath, templateAbsPath)

	cases := []struct {
		cmd  string // command
		out  string // expected output
		pwd  string // change the directory relative to working dir if set
		copy bool   // copy the example folder in temporary dir if set
	}{
		{cmd: "casper fetch", out: outputFile + "\n", pwd: "../../example"},

		// without config file
		{cmd: "casper build -t ../../example/template.yaml -s placeholder1=val1 -s placeholder2=val2", out: outputFile},

		// with overwritten placeholders so it there are differences
		{cmd: "casper diff -s placeholder1=val1a -s placeholder2=val2a --plain", out: changes, pwd: "../../example"},
		{cmd: "casper push -s placeholder1=val1a -s placeholder2=val2a --plain --force", out: changes + applyingChanges, pwd: "../../example"},
		{cmd: "casper push -s placeholder1=val1a -s placeholder2=val2a --plain", out: changes + prompt + canceled + "\n", pwd: "../../example"},

		//
		// Tests for correct relative path resolving.
		//

		// Build tests

		// from the same directory where the config is
		{cmd: "casper build", out: outputFile, pwd: "../../example"},
		{cmd: "casper -c ./config.yaml build", out: outputFile, pwd: "../../example"},
		{cmd: "casper -c ../example/config.yaml build -t ../example/template.yaml", out: outputFile, pwd: "../../example"},

		// from different directory where the config is
		{cmd: "casper -c ../../example/config.yaml build", out: outputFile},
		{cmd: "casper -c ../../example/config.yaml build -t ../../example/template.yaml", out: outputFile},

		// with abs paths
		{cmd: fmt.Sprintf("casper -c %v build -t %v", configAbsPath, templateAbsPath), out: outputFile},

		// Diff tests

		// from the same directory where the config is
		{cmd: "casper diff", out: noChanges, pwd: "../../example"},
		{cmd: "casper -c ./config.yaml diff", out: noChanges, pwd: "../../example"},
		{cmd: "casper -c ../example/config.yaml diff -t ../example/template.yaml", out: noChanges, pwd: "../../example"},

		// from different directory where the config is
		{cmd: "casper -c ../../example/config.yaml diff", out: noChanges},
		{cmd: "casper -c ../../example/config.yaml diff -t ../../example/template.yaml", out: noChanges},

		// with abs paths
		{cmd: fmt.Sprintf("casper -c %v build -t %v", configAbsPath, templateAbsPath), out: outputFile},
	}

	fmt.Println(fmt.Sprintf("casper -c %v build -t %v", configAbsPath, templateAbsPath))
	for i, tc := range cases {
		t.Run(fmt.Sprintf("Case%v", i), func(t *testing.T) {
			os.Chdir(wd)

			// revert any changes to the output.yaml
			defer func() {
				err := ioutil.WriteFile(outputFileName, outputFileData, 0664)
				if err != nil {
					t.Fatal(err)
				}
			}()

			if tc.pwd != "" {
				os.Chdir(tc.pwd)
			}

			os.Args = strings.Split(tc.cmd, " ")
			out := getStdout(t, main)

			if out != tc.out {
				t.Errorf("\n%v> %v\n%v;\nExpected:\n%v;", tc.pwd, tc.cmd, out, tc.out)
			}
		})
	}
}

// Runs a function and returns the stdout from it.
func getStdout(t *testing.T, f func()) string {
	old := os.Stdout // keep backup of the real stdout
	defer func() { os.Stdout = old }()
	r, w, err := os.Pipe()
	if err != nil {
		t.Error(fmt.Errorf("An error wasn't expected: %v", err))
	}
	os.Stdout = w

	f() // executes the main function

	outC := make(chan string)
	// copy the output in a separate goroutine so printing can't block indefinitely
	go func() {
		var buf bytes.Buffer
		io.Copy(&buf, r)
		outC <- buf.String()
	}()

	// back to normal state
	w.Close()
	out := <-outC

	return out
}
