package consul

import (
	"encoding/json"
	"flag"
	"fmt"
	"reflect"
	"testing"

	"github.com/hashicorp/consul/api"
)

// It is defined in each package so you can run `go test ./...`
var full = flag.Bool("full", false, "Run all tests including integration")

func TestConsulToMap(t *testing.T) {
	testCases := []struct {
		pairs api.KVPairs
		json  NestedMap
	}{
		{
			api.KVPairs{
				&api.KVPair{Key: "key1", Value: []byte("val1")},
			},
			NestedMap{
				"key1": "val1",
			},
		},
		{
			api.KVPairs{
				&api.KVPair{Key: "key1", Value: []byte("val1")},
				&api.KVPair{Key: "key2/", Value: []byte("null")},
				&api.KVPair{Key: "key2/1", Value: []byte("val2/1")},
				&api.KVPair{Key: "key2/2", Value: []byte("val2/2")},
				&api.KVPair{Key: "key3/", Value: []byte("null")},
				&api.KVPair{Key: "key3/1/", Value: []byte("null")},
				&api.KVPair{Key: "key3/1/1", Value: []byte("val3/1/1")},
				&api.KVPair{Key: "key3/2", Value: []byte("val3/2")},
				&api.KVPair{Key: "key4/", Value: []byte("null")},
			},
			NestedMap{
				"key1": "val1",
				"key2": map[string]string{
					"_value": "null",
					"1":      "val2/1",
					"2":      "val2/2",
				},
				"key3": map[string]interface{}{
					"_value": "null",
					"1": map[string]interface{}{
						"_value": "null",
						"1":      "val3/1/1",
					},
					"2": "val3/2",
				},
				"key4": map[string]string{
					"_value": "null",
				},
			},
		},
	}

	for i, tc := range testCases {
		t.Run(fmt.Sprintf("Case%v", i), func(t *testing.T) {
			got, _ := json.MarshalIndent(KVPairsToMap(tc.pairs), "", "  ")
			want, _ := json.MarshalIndent(tc.json, "", "  ")

			if string(got) != string(want) {
				t.Errorf("Got\n%v;\nwant\n%v", string(got), string(want))
			}
		})
	}
}

func TestGetChanges(t *testing.T) {
	testCases := []struct {
		pairs   api.KVPairs
		config  []byte
		format  string
		changes []Change
		ok      bool
	}{
		{nil, []byte("{}"), "json", nil, true},
		{nil, []byte(`"key1": "val1"`), "json", nil, false},
		{
			api.KVPairs{
				&api.KVPair{Key: "key1", Value: []byte("val1")},
			},
			[]byte(`{"key1": "val1"}`), "json",
			nil,
			true,
		},
		{
			api.KVPairs{
				&api.KVPair{Key: "key1", Value: []byte("val1")},
			},
			[]byte(`{"key1": "val1", "key2": "val2"}`), "json",
			[]Change{
				{ConsulAdd, "key2", "", "val2"},
			},
			true,
		},
		{
			api.KVPairs{
				&api.KVPair{Key: "key1", Value: []byte("val1")},
				&api.KVPair{Key: "key2", Value: []byte("val2")},
			},
			[]byte(`{"key1": "val1"}`), "json",
			[]Change{
				{ConsulRemove, "key2", "val2", ""},
			},
			true,
		},
		{
			api.KVPairs{
				&api.KVPair{Key: "key1", Value: []byte("val1")},
				&api.KVPair{Key: "key2", Value: []byte("val0")},
			},
			[]byte(`{"key1": "val1", "key2": "val2"}`), "json",
			[]Change{
				{ConsulUpdate, "key2", "val0", "val2"},
			},
			true,
		},
	}

	for i, tc := range testCases {
		t.Run(fmt.Sprintf("Case%v", i), func(t *testing.T) {
			changes, err := GetChanges(tc.pairs, tc.config, tc.format)

			if tc.ok != (err == nil) {
				if err != nil {
					t.Fatal(err)
				} else {
					t.Fatal("Get should have failed but haven't")
				}
			}

			if len(changes) != len(tc.changes) {
				t.Errorf("Got %v; want %v", changes, tc.changes)
			}

			for _, c := range changes {
				found := false
				for _, e := range tc.changes {
					if reflect.DeepEqual(c, e) {
						found = true
						break
					}
				}

				if !found {
					t.Errorf("Got %v that is not expected", c)
				}
			}

			for _, e := range tc.changes {
				found := false
				for _, c := range changes {
					if reflect.DeepEqual(c, e) {
						found = true
						break
					}
				}

				if !found {
					t.Errorf("%v missing", e)
				}

			}
		})
	}
}

func TestStringToMap(t *testing.T) {
	testCases := []struct {
		config []byte
		format string
		pairs  map[string]string
		isOk   bool
	}{
		{[]byte(""), "json", nil, false},
		{[]byte(`{"key1": {"_value": ["val1.1", "val1.2"]}}`), "json", nil, false},
		{
			[]byte(`
                key1: val1
                key2:
                    _value: val2
            `),
			"yaml",
			map[string]string{
				"key1":  "val1",
				"key2/": "val2",
			},
			true,
		},
		{
			[]byte(`{
				"key1": 1,
				"key2":  {"_value": 2.1}
			}`),
			"json",
			map[string]string{
				"key1":  "1",
				"key2/": "2.1",
			},
			true,
		},
		{
			[]byte(`{
				"key1": true,
				"key2":  {"_value": false}
			}`),
			"json",
			map[string]string{
				"key1":  "true",
				"key2/": "false",
			},
			true,
		},
		{
			[]byte(`{
				"key1": "val1",
				"key2": {
					"1": "val2/1",
					"2": "val2/2"
				}
			}`),
			"json",
			map[string]string{
				"key1":   "val1",
				"key2/1": "val2/1",
				"key2/2": "val2/2",
			},
			true,
		},
		{
			[]byte(`{
				"key1": "val1",
				"key2": {
					"1": "val2/1",
					"2": {
						"1": "val2/2/1",
						"2": "val2/2/2"
					}
				}
			}`),
			"json",
			map[string]string{
				"key1":     "val1",
				"key2/1":   "val2/1",
				"key2/2/1": "val2/2/1",
				"key2/2/2": "val2/2/2",
			},
			true,
		},
		{
			[]byte(`{
				"key1": "val1",
				"key2": {
					"_value": "val2",
					"1": "val2/1",
					"2": {
						"_value": "val2/2",
						"1": "val2/2/1",
						"2": "val2/2/2"
					}
				}
			}`),
			"json",
			map[string]string{
				"key1":     "val1",
				"key2/":    "val2",
				"key2/1":   "val2/1",
				"key2/2/":  "val2/2",
				"key2/2/1": "val2/2/1",
				"key2/2/2": "val2/2/2",
			},
			true,
		},
	}

	for i, tc := range testCases {
		t.Run(fmt.Sprintf("Case%v", i), func(t *testing.T) {
			pairs, err := stringToMap(tc.config, tc.format)
			if err != nil {
				if tc.isOk {
					t.Fatalf("Got error %v", err)
				}
				return
			}

			if !tc.isOk {
				t.Fatal("Got no error but should have")
			}

			if !reflect.DeepEqual(pairs, tc.pairs) {
				t.Errorf("Got %v; want %v", pairs, tc.pairs)
			}
		})
	}
}
