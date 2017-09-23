package source

import (
	"fmt"
	"reflect"
	"testing"
)

func TestMultiSourcer(t *testing.T) {
	testCases := []struct {
		s  []Getter
		r  map[string]interface{}
		ok bool
	}{
		{
			[]Getter{
				NewSource(map[string]interface{}{
					"key1": "var1",
					"key2": "var2",
				}),
			},
			map[string]interface{}{
				"key1": "var1",
				"key2": "var2",
			},
			true,
		},
		{
			[]Getter{
				NewSource(map[string]interface{}{
					"key1": "var1",
					"key2": "var2",
				}),
				NewSource(map[string]interface{}{
					"key3": "var3",
					"key4": "var4",
					"key5": "var5",
				}),
			},
			map[string]interface{}{
				"key1": "var1",
				"key2": "var2",
				"key3": "var3",
				"key4": "var4",
				"key5": "var5",
			},
			true,
		},
		{
			[]Getter{
				NewSource(map[string]interface{}{
					"key1": "var1",
					"key2": "var2",
				}),
				NewSource(map[string]interface{}{
					"key1": "var3",
				}),
			},
			nil,
			false,
		},
	}

	for i, tc := range testCases {
		t.Run(fmt.Sprintf("Case%v", i), func(t *testing.T) {
			s, err := NewMultiSourcer(tc.s...)
			if tc.ok != (err == nil) {
				if err != nil {
					t.Fatal(err)
				} else {
					t.Fatal("Get should have failed but haven't")
				}
			}

			if tc.ok && !reflect.DeepEqual(s.Get(), tc.r) {
				t.Errorf("Got %#v; want %#v", s.Get(), tc.r)
			}
		})
	}
}
