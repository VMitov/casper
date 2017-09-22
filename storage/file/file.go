package file

import (
	"bufio"
	"bytes"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/miracl/casper"
	"github.com/pkg/errors"
	"github.com/sergi/go-diff/diffmatchpatch"
)

// Storage is an implementation of the storage interface that stores in file.
// This implementation is mostly for testing.
type Storage struct {
	path string
}

// New returns new file storage
func New(path string) *Storage {
	return &Storage{path: path}
}

func (s Storage) String(format string) (string, error) {
	data, err := ioutil.ReadFile(s.path)
	if err != nil {
		return "", errors.Wrapf(err, "reading file %v failed", s.path)
	}

	return string(data), nil
}

// FormatIsValid returns if the format is valid for this storage
func (s Storage) FormatIsValid(format string) bool {
	return true
}

// DefaultFormat returns the default format
func (s Storage) DefaultFormat() string {
	return "string"
}

// GetChanges returns changes between the config and the Storage content
func (s Storage) GetChanges(config []byte, format, key string) (casper.Changes, error) {
	data, err := ioutil.ReadFile(s.path)
	if err != nil {
		return nil, errors.Wrapf(err, "reading file %v failed", s.path)
	}

	if bytes.Compare(data, config) == 0 {
		return fileChanges{}, nil
	}
	return fileChanges{data, config}, nil
}

// Diff returns the visual representation of the changes
func (s Storage) Diff(cs casper.Changes, pretty bool) string {
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

// Push changes to the storage
func (s Storage) Push(cs casper.Changes) error {
	c := cs.(fileChanges)

	f, err := os.OpenFile(s.path, os.O_WRONLY, 0777)
	if err != nil {
		return errors.Wrapf(err, "opening file %v failed", s.path)
	}
	defer f.Close()

	w := bufio.NewWriter(f)

	if _, err := w.Write(c.new); err != nil {
		return errors.Wrapf(err, "writing to file %v failed", s.path)
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
