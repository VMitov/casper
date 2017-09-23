package main

import (
	"fmt"
	"os"
	"testing"
)

func TestConfigPath(t *testing.T) {
	currDir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}

	if err := os.Chdir("/"); err != nil {
		t.Fatal(err)
	}

	testCases := []struct {
		cfg string
		p   string
		res string
	}{
		{"", "./example", "./example"},
		{"./config/file", "./example", "/config/example"},
	}

	for i, tc := range testCases {
		t.Run(fmt.Sprintf("Case%v", i), func(t *testing.T) {
			res := configPath(tc.cfg, tc.p)

			if res != tc.res {
				t.Errorf("Got %v; want %v", res, tc.res)
			}
		})
	}

	if err := os.Chdir(currDir); err != nil {
		t.Fatal(err)
	}
}
