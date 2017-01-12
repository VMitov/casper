package cmd

import "fmt"

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

var (
	storages = map[string]func(map[string]interface{}) (storage, error){
		"consul": NewConsulStorageConfig,
		"file":   NewFileStorageConfig,
	}
)

type storageError string

func (e storageError) Error() string {
	return fmt.Sprintf("Invalid storage type %v", e)
}

func getStorage(t string, cfg map[string]interface{}) (storage, error) {
	n, ok := storages[t]
	if !ok {
		return nil, storageError(t)
	}

	return n(cfg)
}
