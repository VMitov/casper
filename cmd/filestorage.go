package cmd

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/sergi/go-diff/diffmatchpatch"
)

// fileStorage is an implementation of the storage interface that stores in file.
// This implementation is mostly for testing.
type fileStorage struct {
	path string
}

var errFilePath = errors.New("file store path is invalid type")

func newFileStorageConfig(config map[string]interface{}) (storage, error) {
	path, ok := config["path"].(string)
	if !ok {
		return nil, errFilePath
	}

	return &fileStorage{path}, nil
}

func (s fileStorage) String(format string) (string, error) {
	data, err := ioutil.ReadFile(s.path)
	if err != nil {
		return "", err
	}

	return string(data), nil
}

// FormatIsValid returns if the format is valid for this storage
func (s fileStorage) FormatIsValid(format string) bool {
	return true
}

// DefaultFormat returns the default format
func (s fileStorage) DefaultFormat() string {
	return "string"
}

// GetChanges returns changes between the config and the fileStorage content
func (s fileStorage) GetChanges(config []byte, format, key string) (changes, error) {
	data, err := ioutil.ReadFile(s.path)
	if err != nil {
		return nil, err
	}

	if bytes.Compare(data, config) == 0 {
		return fileChanges{}, nil
	}
	return fileChanges{data, config}, nil
}

// Diff returns the visual representation of the changes
func (s fileStorage) Diff(cs changes, pretty bool) string {
	if cs.Len() == 0 {
		return ""
	}
	c := cs.(fileChanges)

	if pretty {
		dmp := diffmatchpatch.New()
		return dmp.DiffPrettyText(dmp.DiffMain(string(c.old), string(c.new), false))
	}

	return fmt.Sprintf("-%v\n+%v", string(c.old), string(c.new))
}

// Push changes the storage with the given changes
func (s fileStorage) Push(cs changes) error {
	c := cs.(fileChanges)

	f, err := os.OpenFile(s.path, os.O_WRONLY, 0777)
	if err != nil {
		return err
	}
	defer f.Close()

	w := bufio.NewWriter(f)

	if _, err := w.Write(c.new); err != nil {
		return err
	}
	w.Flush()
	return nil
}

type fileChanges struct {
	old []byte
	new []byte
}

func (c fileChanges) Len() int {
	if len(c.old) == 0 && len(c.new) == 0 {
		return 0
	}
	return 1
}
