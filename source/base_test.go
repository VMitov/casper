package source

import (
	"fmt"
	"reflect"
	"testing"
)

func TestSourcer(t *testing.T) {
	testCases := []struct {
		source map[string]interface{}
	}{
		{
			map[string]interface{}{
				"key1": "var1",
				"key2": "var2",
			},
		},
	}

	for i, tc := range testCases {
		t.Run(fmt.Sprintf("Case%v", i), func(t *testing.T) {
			s := NewSource(tc.source)

			if !reflect.DeepEqual(s.Get(), tc.source) {
				t.Errorf("Got %v; want %v", s.Get(), tc.source)
			}
		})
	}
}
