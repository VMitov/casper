package source

// Value source interface
type ValuesSourcer interface {
	Get() map[string]interface{}
}

// Simple value source implementation
type Source struct {
	body map[string]interface{}
}

func NewSource(body map[string]interface{}) *Source {
	return &Source{body}
}

func (s *Source) Get() map[string]interface{} {
	return s.body
}
