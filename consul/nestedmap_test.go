package consul

import (
	"encoding/json"
	"fmt"
	"testing"
)

func TestNestedMap(t *testing.T) {
	type KV struct {
		key string
		val string
	}

	testCases := []struct {
		kv  []KV
		res map[string]interface{}
	}{
		{
			[]KV{
				{"key1", "val1"},
			},
			map[string]interface{}{
				"key1": "val1",
			},
		},
		{
			[]KV{
				{"key1", "val1"},
				{"key2", "val2"},
			},
			map[string]interface{}{
				"key1": "val1",
				"key2": "val2",
			},
		},
		{
			[]KV{
				{"key1", "val1"},
				{"key2/1", "val2/1"},
				{"key2/2", "val2/2"},
			},
			map[string]interface{}{
				"key1": "val1",
				"key2": map[string]string{
					"1": "val2/1",
					"2": "val2/2",
				},
			},
		},
		{
			[]KV{
				{"key1", "val1"},
				{"key2/", "null"},
				{"key2/1", "val2/1"},
				{"key2/2", "val2/2"},
			},
			map[string]interface{}{
				"key1": "val1",
				"key2": map[string]string{
					"1":      "val2/1",
					"2":      "val2/2",
					"_value": "null",
				},
			},
		},
		{
			[]KV{
				{"key1", "val1"},
				{"key2", "null"},
				{"key2/1", "val2/1"},
				{"key2/2", "val2/2"},
			},
			map[string]interface{}{
				"key1": "val1",
				"key2": map[string]string{
					"_value": "null",
					"1":      "val2/1",
					"2":      "val2/2",
				},
			},
		},
		{
			[]KV{
				{"key1/1", "val1"},
				{"key1", "val2"},
			},
			map[string]interface{}{
				"key1": map[string]string{
					"1":      "val1",
					"_value": "val2",
				},
			},
		},
		{
			[]KV{
				{"key1", "val1"},
				{"key1", "val2"},
			},
			map[string]interface{}{
				"key1": "val2",
			},
		},

		// Test that the order doesn't matter
		{
			[]KV{
				{"key1", "val1"},
				{"key2/", "null"},
				{"key2/1", "val2/1"},
				{"key2/2", "val2/2"},
				{"key3/", "null"},
				{"key3/1/", "null"},
				{"key3/1/1", "val3/1/1"},
				{"key3/2", "val3/2"},
				{"key4/", "null"},
			},
			map[string]interface{}{
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
		{
			// Folders first, then values in order
			[]KV{
				{"key2/", "null"},
				{"key3/", "null"},
				{"key3/1/", "null"},
				{"key4/", "null"},
				{"key1", "val1"},
				{"key2/1", "val2/1"},
				{"key2/2", "val2/2"},
				{"key3/2", "val3/2"},
				{"key3/1/1", "val3/1/1"},
			},
			map[string]interface{}{
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
		{
			[]KV{
				{"key3/1/1", "val3/1/1"},
				{"key3/2", "val3/2"},
				{"key2/2", "val2/2"},
				{"key2/1", "val2/1"},
				{"key1", "val1"},
				{"key4/", "null"},
				{"key3/1/", "null"},
				{"key3/", "null"},
				{"key2/", "null"},
			},
			map[string]interface{}{
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
			j := &NestedMap{}
			for _, item := range tc.kv {
				j.Add(item.key, item.val)
			}

			got, _ := json.MarshalIndent(j, "", "  ")
			want, _ := json.MarshalIndent(tc.res, "", "  ")
			if string(got) != string(want) {
				t.Errorf("Got\n%v;\nwant\n%v", string(got), string(want))
			}
		})
	}
}
