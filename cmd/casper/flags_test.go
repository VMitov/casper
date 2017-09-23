package main

import (
	"fmt"
	"net/url"
	"testing"
)

func TestFixPathsForFileSource(t *testing.T) {
	dir := "/config"
	cases := []struct {
		u string
		f string
	}{
		{
			u: "file:///config/source.yaml?a=1",
			f: "file:///config/source.yaml?a=1",
		},
		{
			u: "file://source.yaml?a=1",
			f: "file:///config/source.yaml?a=1",
		},
		{
			u: "file://./source.yaml?a=1",
			f: "file:///config/source.yaml?a=1",
		},
		{
			u: "file://../source.yaml?a=1",
			f: "file:///source.yaml?a=1",
		},
		{
			u: "file://../config/source.yaml?a=1",
			f: "file:///config/source.yaml?a=1",
		},
	}

	for i, tc := range cases {
		t.Run(fmt.Sprintf("Case%v", i), func(t *testing.T) {
			u, err := url.Parse(tc.u)
			if err != nil {
				t.Fatal(err)
			}

			f, err := fixPathsForFileSource(dir, u)
			if err != nil {
				t.Fatal(err)
			}

			if f != tc.f {
				t.Errorf("Got: %v; Expected: %v", f, tc.f)
			}
		})
	}
}
