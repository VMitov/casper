package cmd

import (
	"errors"
	"fmt"
	"log"
	"os"
	"reflect"
	"testing"
)

func TestGetConfigSource(t *testing.T) {
	testCases := []struct {
		cfg map[string]interface{}
		res map[string]interface{}
		ok  bool
	}{
		{
			map[string]interface{}{
				"vals": map[string]string{
					"key1": "var1",
					"key2": "var2",
				},
			},
			map[string]interface{}{
				"key1": "var1",
				"key2": "var2",
			},
			true,
		},
		{
			map[string]interface{}{
				"opts": map[interface{}]interface{}{
					"key1": "var1",
					"key2": "var2",
				},
			},
			nil,
			false,
		},
		{map[string]interface{}{"vals": 42}, nil, false},
	}

	for i, tc := range testCases {
		t.Run(fmt.Sprintf("Case%v", i), func(t *testing.T) {

			s, err := getConfigSource(tc.cfg)
			if tc.ok != (err == nil) {
				t.Fatalf("Failed with: %v", err)
			}

			if tc.ok && !reflect.DeepEqual(s.Get(), tc.res) {
				t.Errorf("Got %v; want %v", s.Get(), tc.res)
			}
		})
	}
}

func TestGetFileSource(t *testing.T) {
	testCases := []struct {
		cfg        map[string]interface{}
		createFile bool
		file       string
		res        map[string]interface{}
		ok         bool
	}{
		{
			map[string]interface{}{"file": "TestGetFileSource"},
			true, `{"key": "var"}`,
			map[string]interface{}{"key": "var"},
			true,
		},
		{
			map[string]interface{}{"file": "TestGetFileSource"},
			true, `key: var`,
			map[string]interface{}{"key": "var"},
			false,
		},
		{
			map[string]interface{}{"file": "TestGetFileSource"},
			false, `{"key": "var"}`,
			map[string]interface{}{"key": "var"},
			false,
		},
		{
			map[string]interface{}{"file": "TestGetFileSource", "format": "yaml"},
			false, `key: var`,
			map[string]interface{}{"key": "var"},
			false,
		},
		{map[string]interface{}{}, false, "", nil, false},
		{map[string]interface{}{"file": 1}, false, "", nil, false},
		{nil, false, "", nil, false},
	}

	for i, tc := range testCases {
		t.Run(fmt.Sprintf("Case%v", i), func(t *testing.T) {
			if tc.createFile {
				// Prepare tmp file
				tmpfile, err := os.Create("TestGetFileSource")
				if err != nil {
					log.Fatal(err)
				}
				defer os.Remove(tmpfile.Name()) // clean up

				if _, err := tmpfile.Write([]byte(tc.file)); err != nil {
					t.Fatal(err)
				}
				if err := tmpfile.Close(); err != nil {
					t.Fatal(err)
				}
			}

			s, err := getFileSource(tc.cfg)
			if tc.ok != (err == nil) {
				t.Fatalf("Failed with: %v", err)
			}

			if tc.ok && !reflect.DeepEqual(s.Get(), tc.res) {
				t.Errorf("Got %v; want %v", s.Get(), tc.res)
			}
		})
	}
}

func TestGetMultiSourcer(t *testing.T) {
	testCases := []struct {
		srcs []map[string]interface{}
		res  map[string]interface{}
		ok   bool
	}{
		{
			[]map[string]interface{}{
				{
					"type": "config",
					"vals": map[interface{}]interface{}{
						"key1": "var1",
						"key2": "var2",
					},
				},
			},
			map[string]interface{}{
				"key1": "var1",
				"key2": "var2",
			},
			true,
		},
		{
			[]map[string]interface{}{
				{
					"sourceType": "config",
				},
			},
			nil,
			false,
		},
		{
			[]map[string]interface{}{
				{
					"type": 42,
				},
			},
			nil,
			false,
		},
		{
			[]map[string]interface{}{
				{
					"type": "invalid sourcer type",
				},
			},
			nil,
			false,
		},
		{
			[]map[string]interface{}{
				{
					"type": "config",
					"vals": 42,
				},
			},
			nil,
			false,
		},
	}

	for i, tc := range testCases {
		t.Run(fmt.Sprintf("Case%v", i), func(t *testing.T) {
			s, err := getMultiSourcer(tc.srcs)
			if tc.ok != (err == nil) {
				t.Fatalf("Failed with: %v", err)
			}

			if tc.ok && !reflect.DeepEqual(s.Get(), tc.res) {
				t.Errorf("Got %v; want %v", s.Get(), tc.res)
			}
		})
	}
}

func TestToStringMapString(t *testing.T) {
	testCases := []struct {
		cfg interface{}
		res map[string]interface{}
		ok  bool
	}{
		{
			map[string]string{
				"key1": "var1",
				"key2": "var2",
			},
			map[string]interface{}{
				"key1": "var1",
				"key2": "var2",
			},
			true,
		},
		{
			map[string]interface{}{
				"key1": "var1",
				"key2": 2,
				"key3": false,
			},
			map[string]interface{}{
				"key1": "var1",
				"key2": 2,
				"key3": false,
			},
			true,
		},
		{
			map[interface{}]interface{}{
				"key1": "var1",
				"key2": "var2",
			},
			map[string]interface{}{
				"key1": "var1",
				"key2": "var2",
			},
			true,
		},
		{
			map[string]interface{}{"key1": 2.5},
			map[string]interface{}{"key1": 2.5},
			true,
		},
		{
			map[interface{}]interface{}{"key1": 0.1},
			map[string]interface{}{"key1": 0.1},
			true,
		},
		{
			map[int]string{
				1: "var1",
				2: "var2",
			},
			nil,
			false,
		},
		{map[interface{}]interface{}{0.1: "var1"}, nil, false},
	}

	for i, tc := range testCases {
		t.Run(fmt.Sprintf("Case%v", i), func(t *testing.T) {

			res, err := toStringMapString(tc.cfg)
			if tc.ok != (err == nil) {
				t.Fatalf("Failed with: %v", err)
			}

			if tc.ok && !reflect.DeepEqual(res, tc.res) {
				t.Errorf("Got %v; want %v", res, tc.res)
			}
		})
	}
}

func TestSourceFormatError(t *testing.T) {
	testCases := []struct {
		inputMessage string
		inputError   error
		expected     string
	}{
		{
			"test1234", nil,
			"Invalid source definition: test1234",
		},
		{
			"test1234", errors.New("test444"),
			"Invalid source definition: test1234 (Err:test444)",
		},
	}

	for i, tc := range testCases {
		t.Run(fmt.Sprintf("Case%v", i), func(t *testing.T) {
			output := sourceFormatError{
				tc.inputMessage, tc.inputError,
			}.Error()

			if output != tc.expected {
				t.Errorf("Got \"%s\"; want \"%s\"", output, tc.expected)
			}
		})
	}
}

func TestConvertError(t *testing.T) {
	testCases := []struct {
		inputInterface interface{}
		inputType      string
		expected       string
	}{
		{
			"test string", "int32",
			"Unable to convert \"test string\" to int32",
		},
	}

	for i, tc := range testCases {
		t.Run(fmt.Sprintf("Case%v", i), func(t *testing.T) {
			output := convertError{
				tc.inputInterface, tc.inputType,
			}.Error()

			if output != tc.expected {
				t.Errorf("Got \"%s\"; want \"%s\"", output, tc.expected)
			}
		})
	}
}

func TestToString(t *testing.T) {
	testCases := []struct {
		input       interface{}
		expectedStr string
		expectedOK  bool
	}{
		{"24abc", "24abc", true},
		{24, "24", true},
		{true, "true", true},
		{false, "false", true},
		{-150.29, "", false},
	}

	for i, tc := range testCases {
		t.Run(fmt.Sprintf("Case%v", i), func(t *testing.T) {
			str, ok := toString(tc.input)

			if str != tc.expectedStr || ok != tc.expectedOK {
				t.Errorf("Got \"%s\", %t; want \"%s\", %t",
					str, ok, tc.expectedStr, tc.expectedOK)
			}
		})
	}
}
