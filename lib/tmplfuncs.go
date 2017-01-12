package casper

import (
	"fmt"
	"reflect"
	"strconv"
)

func isLast(x int, a interface{}) bool {
	return x == reflect.ValueOf(a).Len()-1
}

func isNotLast(x int, a interface{}) bool {
	return x != reflect.ValueOf(a).Len()-1
}

func quote(a interface{}) string {
	if a == nil {
		return `""`
	}

	var s string
	switch v := a.(type) {
	case string:
		s = v
	case int:
		s = strconv.Itoa(v)
	default:
		s = "Unable to quote"
	}

	return fmt.Sprintf("%#v", s)
}
