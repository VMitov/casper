package cmd

import "github.com/spf13/cast"

func getSliceStringMapIface(val interface{}) ([]map[string]interface{}, bool) {
	slice, ok := val.([]interface{})
	if !ok {
		return nil, false
	}

	res := make([]map[string]interface{}, len(slice))
	for i, v := range slice {
		res[i] = cast.ToStringMap(v)
	}
	return res, true
}
