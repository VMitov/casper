package source

type dupKeyError string

func (e dupKeyError) Error() string {
	return "Duplicated key: " + string(e)
}

// NewMultiSourcer create source that is a collection of value sources
func NewMultiSourcer(vss ...ValuesSourcer) (*Source, error) {
	vars := map[string]interface{}{}

	for _, s := range vss {
		for k, v := range s.Get() {
			if _, ok := vars[k]; ok {
				return nil, dupKeyError(k)
			}
			vars[k] = v
		}
	}

	return &Source{vars}, nil
}
