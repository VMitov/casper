package cmd

import (
	"fmt"
	"testing"
)

func TestGetSliceStringMapIface(t *testing.T) {
	testCases := []struct {
		val interface{}
		res []map[string]interface{}
		ok  bool
	}{
		{nil, nil, false},
		{5, nil, false},
		{
			interface{}(
				[]interface{}{
					map[interface{}]interface{}{
						"key1": "var1",
						"key2": "var2",
					},
					map[interface{}]interface{}{
						"key3": "var3",
						"key4": "var4",
					},
				},
			),
			[]map[string]interface{}{
				{
					"key1": "var1",
					"key2": "var2",
				},
				{
					"key3": "var3",
					"key4": "var4",
				},
			},
			true,
		},
	}

	for i, tc := range testCases {
		t.Run(fmt.Sprintf("Case%v", i), func(t *testing.T) {
			res, ok := getSliceStringMapIface(tc.val)
			if ok != tc.ok {
				t.Fatalf("Conversion should have been ok")
			}

			if len(res) != len(tc.res) {
				t.Errorf("Got %v; want %v", res, tc.res)
			}
		})
	}
}
