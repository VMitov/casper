package main

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/miracl/casper/lib/caspertest"
)

func TestMain(t *testing.T) {
	os.Args = []string{"casper", "-c", "../../example/config.yaml", "build", "-t", "../../example/template.yaml"}
	out := caspertest.GetStdout(t, main)

	expected, err := ioutil.ReadFile("../../example/output.yaml")
	if err != nil {
		t.Fatal(err)
	}

	if out != string(expected) {
		t.Errorf("Got\n%v;\nWant\n%v;", out, string(expected))
	}
}