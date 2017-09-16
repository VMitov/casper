package main

import "fmt"

func strChanges(cs changes, key string, s storage, pretty bool) string {
	if cs.Len() == 0 {
		if key != "" {
			return fmt.Sprintf("No changes for key %v", key)
		}
		return fmt.Sprintf("No changes")
	}
	return s.Diff(cs, pretty)
}
