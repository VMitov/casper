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

// Implementation of the storage interface that stores in file.
// This implementation is mostly for testing.
type FileStorage struct {
	path string
}

var errFilePath = errors.New("File store path is invalid type.")

func NewFileStorageConfig(config map[string]interface{}) (storage, error) {
	path, ok := config["path"].(string)
	if !ok {
		return nil, errFilePath
	}

	return &FileStorage{path}, nil
}

func (s FileStorage) String(format string) (string, error) {
	data, err := ioutil.ReadFile(s.path)
	if err != nil {
		return "", err
	}

	return string(data), nil
}

func (s FileStorage) FormatIsValid(format string) bool {
	return true
}

func (s FileStorage) DefaultFormat() string {
	return "string"
}

func (s FileStorage) GetChanges(config []byte, format, key string) (changes, error) {
	data, err := ioutil.ReadFile(s.path)
	if err != nil {
		return nil, err
	}

	if bytes.Compare(data, config) == 0 {
		return FileChanges{}, nil
	}
	return FileChanges{data, config}, nil
}

func (s FileStorage) Diff(cs changes, pretty bool) string {
	if cs.Len() == 0 {
		return ""
	}
	c := cs.(FileChanges)

	if pretty {
		dmp := diffmatchpatch.New()
		return dmp.DiffPrettyText(dmp.DiffMain(string(c.old), string(c.new), false))
	}

	return fmt.Sprintf("-%v\n+%v", string(c.old), string(c.new))
}

func (s FileStorage) Push(cs changes) error {
	c := cs.(FileChanges)

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

type FileChanges struct {
	old []byte
	new []byte
}

func (c FileChanges) Len() int {
	if len(c.old) == 0 && len(c.new) == 0 {
		return 0
	}
	return 1
}
