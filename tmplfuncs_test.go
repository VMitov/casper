package casper

import (
	"fmt"
	"testing"
)

func TestIsLastAndIsNotLast(t *testing.T) {
	a := []int{1, 2, 3, 4}
	testCases := []struct {
		x      int
		isLast bool
	}{
		{-1, false},
		{0, false},
		{1, false},
		{2, false},
		{3, true},
		{4, false},
	}

	for i, tc := range testCases {
		t.Run(fmt.Sprintf("Case%v", i), func(t *testing.T) {
			if isLast(tc.x, a) != tc.isLast {
				t.Errorf("Got isLast=%v; want %v", isLast(tc.x, a), tc.isLast)
			}

			if isNotLast(tc.x, a) != !tc.isLast {
				t.Errorf("Got isNotLast=%v; want %v", isNotLast(tc.x, a), !tc.isLast)
			}
		})
	}
}

func TestQuote(t *testing.T) {
	testCases := []struct {
		a interface{}
		q string
	}{
		{1, `"1"`},
		{"", `""`},
		{nil, `""`},
		{"a", `"a"`},
		{false, `"Unable to quote"`},
	}

	for i, tc := range testCases {
		t.Run(fmt.Sprintf("Case%v", i), func(t *testing.T) {
			if quote(tc.a) != tc.q {
				t.Errorf("Got %v; want %v", quote(tc.a), tc.q)
			}
		})
	}
}
