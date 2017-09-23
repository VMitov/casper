package source

import "fmt"

// NewMultiSourcer create source that is a collection of value sources.
func NewMultiSourcer(vss ...Getter) (*Source, error) {
	vars := map[string]interface{}{}

	for _, s := range vss {
		for k, v := range s.Get() {
			if _, ok := vars[k]; ok {
				return nil, fmt.Errorf("duplicated key '%v'", k)
			}
			vars[k] = v
		}
	}

	return &Source{vars}, nil
}
