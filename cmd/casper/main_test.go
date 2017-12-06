package main

import (
	"bytes"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/miracl/casper/storage/consul"
)

// It is defined in each package so you can run `go test ./...`
var full = flag.Bool("full", false, "Run all tests including integration")

var consulAddr = flag.String("consul-addr", "http://172.17.0.1:8500/?token=the_one_ring", "Consul instance to run tests agains")

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

	cases := []struct {
		cmd  string // command
		out  string // expected output
		pwd  string // change the directory relative to working dir if set
		copy bool   // copy the example folder in temporary dir if set
		err  string // string representation of expected error
	}{
		{cmd: "casper fetch", out: outputFile + "\n", pwd: "../../example"},
		{cmd: "casper diff -key somekey", out: "No changes for key somekey\n", pwd: "../../example"},

		// without config file
		{cmd: "casper build -t ../../example/template.yaml -s placeholder1=val1 -s placeholder2=val2", out: outputFile},

		// with bad source
		{cmd: "casper build -t ../../example/template.yaml -s placeholder1=val1 -s source.yaml", out: outputFile, err: "creating context failed: invalid source: source.yaml"},

		// without sources
		{cmd: "casper build -t ../../example/output.yaml", out: outputFile},
		{cmd: "casper diff -t ../../example/output.yaml -storage file -file-path ../../example/output.yaml", out: noChanges},

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

			if tc.err != "" {
				app := newApp()
				err := app.Run(strings.Split(tc.cmd, " "))
				if err.Error() != tc.err {
					t.Fatalf("\nunexpected error: %v\n\texpected: %v", err, tc.err)
				}
				return
			}

			if tc.pwd != "" {
				os.Chdir(tc.pwd)
			}

			os.Args = strings.Split(tc.cmd, " ")
			out := getStdout(t, main)
			if out != tc.out {
				t.Errorf("\n%vtest:/$ %v\n%v;\nExpected:\n%v;", tc.pwd, tc.cmd, out, tc.out)
			}
		})
	}
}

func TestConsulIntegration(t *testing.T) {
	if !*full {
		t.SkipNow()
	}

	// Cleanup
	defer func() {
		s, err := consul.New(*consulAddr)
		if err != nil {
			t.Fatalf("cleanup failed: %v", err)
		}

		c, err := s.GetChanges([]byte{}, "yaml", "")
		if err != nil {
			t.Fatalf("cleanup failed: %v", err)
		}

		err = s.Push(c)
		if err != nil {
			t.Fatalf("cleanup failed: %v", err)
		}
	}()

	steps := []struct {
		cmd string
		exp string
	}{
		{
			cmd: fmt.Sprintf("casper fetch -format yaml -storage consul -consul-addr %v", *consulAddr),
			exp: "{}\n\n",
		},

		{
			cmd: "casper diff -plain" +
				" -storage consul -consul-addr " + *consulAddr +
				" -template ../../example/template.yaml -s placeholder1=val1 -s placeholder2=val2",
			exp: "+key1=val1\n+key2=val2\n\n",
		},
		{
			cmd: "casper push -plain -force" +
				" -storage consul -consul-addr " + *consulAddr +
				" -template ../../example/template.yaml -s placeholder1=val1 -s placeholder2=val2",
			exp: "+key1=val1\n+key2=val2\n\nApplying changes...\n",
		},
		{
			cmd: "casper diff -plain" +
				" -storage consul -consul-addr " + *consulAddr +
				" -template ../../example/template.yaml -s placeholder1=val1 -s placeholder2=val2",
			exp: "No changes\n",
		},
		{
			cmd: "casper fetch -format yaml -storage consul -consul-addr " + *consulAddr,
			exp: "key1: val1\nkey2: val2\n\n",
		},

		{
			cmd: "casper diff -plain" +
				" -storage consul -consul-addr " + *consulAddr +
				" -template ../../example/template.yaml -s placeholder1=diffval1 -s placeholder2=diffval2",
			exp: "-key1=val1\n+key1=diffval1\n-key2=val2\n+key2=diffval2\n\n",
		},
		{
			cmd: "casper push -plain -force" +
				" -storage consul -consul-addr " + *consulAddr +
				" -template ../../example/template.yaml -s placeholder1=diffval1 -s placeholder2=diffval2",
			exp: "-key1=val1\n+key1=diffval1\n-key2=val2\n+key2=diffval2\n\nApplying changes...\n",
		},
		{
			cmd: "casper diff -plain" +
				" -storage consul -consul-addr " + *consulAddr +
				" -template ../../example/template.yaml -s placeholder1=diffval1 -s placeholder2=diffval2",
			exp: "No changes\n",
		},
		{
			cmd: "casper fetch -format yaml -storage consul -consul-addr " + *consulAddr,
			exp: "key1: diffval1\nkey2: diffval2\n\n",
		},
	}

	for i, step := range steps {
		os.Args = strings.Split(step.cmd, " ")
		out := getStdout(t, main)
		if out != step.exp {
			t.Errorf("\nstep%v:/$ %v\n%v;\nExpected:\n%v;", i, step.cmd, out, step.exp)
		}
	}
}

func TestAppErrors(t *testing.T) {
	cases := []struct {
		cmd string
		err string
	}{
		// no configurations
		{cmd: "casper fetch", err: "reading file casper.yaml failed: open casper.yaml: no such file or directory"},
		{cmd: "casper build", err: "creating context failed: getting template template.yaml failed: open template.yaml: no such file or directory"},
		{cmd: "casper diff", err: "creating context failed: getting template template.yaml failed: open template.yaml: no such file or directory"},
		{cmd: "casper push", err: "creating context failed: getting template template.yaml failed: open template.yaml: no such file or directory"},

		// invalid storage
		{cmd: "casper fetch -storage invalid", err: "invalid storage type 'invalid'"},
		{cmd: "casper diff -storage invalid -t ../../example/template.yaml -s key=value", err: "invalid storage type 'invalid'"},
		{cmd: "casper push -storage invalid -t ../../example/template.yaml -s key=value", err: "invalid storage type 'invalid'"},

		// file storage - invalid file
		{cmd: "casper fetch -storage file -file-path invalid.txt", err: "reading file invalid.txt failed: open invalid.txt: no such file or directory"},
		{cmd: "casper diff -storage file -file-path invalid.txt", err: "creating context failed: getting template template.yaml failed: open template.yaml: no such file or directory"},
		{cmd: "casper push -storage file -file-path invalid.txt", err: "creating context failed: getting template template.yaml failed: open template.yaml: no such file or directory"},

		// invalid source
		{cmd: "casper build -t ../../example/template.yaml -s key:val", err: "creating context failed: invalid source format key"},
		{cmd: "casper diff -t ../../example/template.yaml -s key:val", err: "creating context failed: invalid source format key"},
		{cmd: "casper push -t ../../example/template.yaml -s key:val", err: "creating context failed: invalid source format key"},
	}

	for i, tc := range cases {
		t.Run(fmt.Sprintf("Case%v", i), func(t *testing.T) {
			args := strings.Split(tc.cmd, " ")
			app := newApp()
			err := app.Run(args)

			if err != nil && tc.err == "" {
				t.Fatalf("unexpected error '%v'", err)
			}

			if err.Error() != tc.err {
				t.Errorf("\ntest:/$ %v\n%v\nExpected:\n%v;", tc.cmd, err, tc.err)
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
		t.Errorf("an error wasn't expected: %v", err)
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
