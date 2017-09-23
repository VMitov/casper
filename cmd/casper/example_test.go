package main

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/miracl/casper/caspertest"
)

func TestMain(t *testing.T) {
	os.Args = []string{"casper", "build", "-c", "../../example/config.yaml", "-t", "../../example/template.yaml"}
	out := caspertest.GetStdout(t, main)

	expected, err := ioutil.ReadFile("../../example/output.yaml")
	if err != nil {
		t.Fatal(err)
	}

	if out != string(expected) {
		t.Errorf("Got %v; Want %v", out, string(expected))
	}
}
