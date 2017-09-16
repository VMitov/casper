package main

type storage interface {
	String(format string) (string, error)
	FormatIsValid(format string) bool
	DefaultFormat() string
	GetChanges(config []byte, format, key string) (changes, error)
	Diff(cs changes, pretty bool) string
	Push(cs changes) error
}

type changes interface {
	Len() int
}
