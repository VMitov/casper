package source

// ValuesSourcer is interface for sources
type ValuesSourcer interface {
	Get() map[string]interface{}
}

// Source is a simple ValuesSourcer implementation
type Source struct {
	body map[string]interface{}
}

// NewSource creates new Source
func NewSource(body map[string]interface{}) *Source {
	return &Source{body}
}

// Get returns the values from the source
func (s *Source) Get() map[string]interface{} {
	return s.body
}
