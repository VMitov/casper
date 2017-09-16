package diff

import (
	"fmt"
	"testing"
)

func TestDiff(t *testing.T) {
	testCases := []struct {
		changes []KVChange
		plain   string
		pretty  string
	}{
		{
			[]KVChange{
				NewAdd("keyAdd", "valAdd"),
				NewAdd("keyAddEmpty", ""),
				NewUpdate("keyUpdate", "valUpdateOld", "valUpdateNew"),
				NewRemove("keyRemove", "valRemove"),
			},
			"" +
				"+keyAdd=valAdd\n" +
				"+keyAddEmpty=\"\"\n" +
				"-keyRemove=valRemove\n" +
				"-keyUpdate=valUpdateOld\n" +
				"+keyUpdate=valUpdateNew\n",
			"" +
				"keyAdd=valAdd\n" +
				"keyAddEmpty=\"\"\n" +
				"keyRemove=valRemove\n" +
				"keyUpdate=valUpdate\033[31mOld\033[0m\033[32mNew\033[0m\n",
		},
	}

	for i, tc := range testCases {
		t.Run(fmt.Sprintf("Case%v", i), func(t *testing.T) {
			plain := Diff(tc.changes, false)
			pretty := Diff(tc.changes, true)

			if plain != tc.plain {
				t.Errorf("Got:\n%vwant:\n%v", plain, tc.plain)
			}

			if pretty != tc.pretty {
				t.Errorf("Got:\n%vwant:\n%v", pretty, tc.pretty)
			}
		})
	}
}
