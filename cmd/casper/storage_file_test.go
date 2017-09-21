package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"testing"
)

func TestFileStorageString(t *testing.T) {
	testCases := []struct {
		data string
	}{
		{`{"key": "val"}`},
	}

	for i, tc := range testCases {
		name := fmt.Sprintf("Case%v", i)
		t.Run(name, func(t *testing.T) {
			f, err := prepareTmpFile(name, []byte(tc.data))
			if err != nil {
				t.Fatal(err)
			}
			defer os.Remove(f.Name())

			s := &fileStorage{path: f.Name()}
			out, err := s.String("string")
			if err != nil {
				t.Fatal(err)
			}

			if out != tc.data {
				t.Errorf("Got `%v`; want `%v`", out, tc.data)
			}
		})
	}
}

func TestFileStorageDiff(t *testing.T) {
	testCases := []struct {
		data   string
		conf   string
		plain  string
		pretty string
	}{
		{
			`{"key": "val"}`,
			`{"key": "val"}`,
			"", "",
		},
		{
			`{"key": "val"}`,
			`{"key": "val", "keyNew": "valNew"}`,
			"" +
				`-{"key": "val"}` + "\n" +
				`+{"key": "val", "keyNew": "valNew"}`,
			"{\"key\": \"val\"\033[32m, \"keyNew\": \"valNew\"\033[0m}",
		},
	}

	for i, tc := range testCases {
		name := fmt.Sprintf("Case%v", i)
		t.Run(name, func(t *testing.T) {
			f, err := prepareTmpFile(name, []byte(tc.data))
			if err != nil {
				t.Fatal(err)
			}
			defer os.Remove(f.Name())

			s := &fileStorage{path: f.Name()}
			changes, err := s.GetChanges([]byte(tc.conf), "string", "")
			if err != nil {
				t.Fatal(err)
			}

			plain := s.Diff(changes, false)
			if plain != tc.plain {
				t.Errorf("Got `%v`; want `%v`", plain, tc.plain)
			}

			pretty := s.Diff(changes, true)
			if pretty != tc.pretty {
				t.Errorf("Got `%v`; want `%v`", pretty, tc.pretty)
			}
		})
	}
}

func TestFileStoragePush(t *testing.T) {
	testCases := []struct {
		data string
		conf string
	}{
		{
			`{"key": "val"}`,
			`{"key": "val", "keyNew": "valNew"}`,
		},
	}

	for i, tc := range testCases {
		name := fmt.Sprintf("Case%v", i)
		t.Run(name, func(t *testing.T) {
			f, err := prepareTmpFile(name, []byte(tc.data))
			if err != nil {
				t.Fatal(err)
			}
			defer os.Remove(f.Name())

			s := &fileStorage{path: f.Name()}
			changes, err := s.GetChanges([]byte(tc.conf), "string", "")
			if err != nil {
				t.Fatal(err)
			}

			if err := s.Push(changes); err != nil {
				t.Fatal(err)
			}

			dat, err := ioutil.ReadFile(f.Name())
			if err != nil {
				t.Fatal(err)
			}

			if string(dat) != tc.conf {
				t.Errorf("Got `%v`; want `%v`", string(dat), tc.conf)
			}
		})
	}
}

func TestFileStorageFormats(t *testing.T) {
	testCases := []struct {
		fmt string
		def string
	}{
		{
			"fmt1",
			"string",
		},
		{
			"fmt2",
			"string",
		},
		{
			"fmt3",
			"string",
		},
	}

	for i, tc := range testCases {
		t.Run(fmt.Sprintf("Case%v", i), func(t *testing.T) {
			s := &fileStorage{}

			if !s.FormatIsValid(tc.fmt) {
				t.Errorf("%v should have been valid", tc.fmt)
			}

			if s.DefaultFormat() != tc.def {
				t.Errorf("Default format should have been %v, not %v", tc.def, s.DefaultFormat())
			}
		})
	}
}

// prepareTmpFile create a file with the given content
func prepareTmpFile(name string, data []byte) (*os.File, error) {
	f, err := os.Create(name)
	if err != nil {
		return nil, err
	}

	if _, err := f.Write(data); err != nil {
		return nil, err
	}
	if err := f.Close(); err != nil {
		return nil, err
	}

	return f, nil
}
