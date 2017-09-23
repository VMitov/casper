package casper

// Storage is interface for storages
type Storage interface {
	String(format string) (string, error)
	GetChanges(config []byte, format, key string) (Changes, error)
	Diff(cs Changes, pretty bool) string
	Push(cs Changes) error
}

// Changes is interface for changes
type Changes interface {
	Len() int
}
