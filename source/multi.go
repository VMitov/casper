package source

import (
	"fmt"
)

// NewMultiSourcer create source that is a collection of value sources.
func NewMultiSourcer(vss ...Getter) (*Source, error) {
	vars := map[string]interface{}{}

	for _, s := range vss {
		for k, v := range s.Get() {
			if prev, ok := vars[k]; ok {
				switch p := prev.(type) {
				case interface{}:
					vars[k] = []interface{}{p, v}
				case []interface{}:
					vars[k] = append(p, v)
				default:
					return nil, fmt.Errorf("multy sources merging failed")
				}

				continue

			}
			vars[k] = v
		}
	}

	return &Source{vars}, nil
}
